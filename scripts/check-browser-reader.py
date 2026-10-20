"""Actual bundled OCR on explicitly SYNTHETIC tile images, not photo accuracy.

Requires generated reader assets, Playwright/Pillow and system Chromium.
Optionally pass a real image path and zone for local proposal export. That does
not create ground truth or certify recognition; original images are not stored
in the repository or sent anywhere by this script.
"""
import argparse
import base64
import io
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont
from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--image', type=Path)
parser.add_argument('--zone', choices=['table', 'rack'], default='rack')
parser.add_argument('--output', type=Path)
args = parser.parse_args()
root = Path(__file__).resolve().parents[1] / 'phone/web'
font_path = '/usr/share/fonts/truetype/open-sans/OpenSans-ExtraBold.ttf'
font = ImageFont.truetype(font_path, 70)
colors = {'red': '#e00b20', 'black': '#080808', 'blue': '#078b92', 'orange': '#ff6310'}


def fixture(values, touching=False, joker=False):
    image = Image.new('RGB', (1200, 1100), (170, 120, 75))
    draw = ImageDraw.Draw(image)
    for i, (number, color) in enumerate(values):
        x, y = 55+(i % 7)*(145 if touching else 160), 110+(i//7)*370
        draw.rounded_rectangle((x, y, x+145, y+220), radius=8, fill='#f8efdb')
        draw.text((x+73, y+52), str(number), font=font, fill=colors[color], anchor='mt')
    if joker:
        x, y = 55+3*160, 480
        draw.rounded_rectangle((x, y, x+145, y+220), radius=8, fill='#f8efdb')
        draw.ellipse((x+28, y+34, x+116, y+122), outline='#080808', width=4)
        for dx in [48, 92]:
            draw.ellipse((x+dx-3, y+62, x+dx+3, y+68), fill='#080808')
        draw.arc((x+48, y+70, x+96, y+104), 0, 180, fill='#080808', width=4)
    return image


def payload(image):
    buffer = io.BytesIO()
    image.save(buffer, format='PNG')
    return {'name': 'synthetic.png', 'mimeType': 'image/png', 'buffer': buffer.getvalue()}


with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    context = browser.new_context(offline=True, viewport={'width': 390, 'height': 844})
    requested, errors, console_errors = [], [], []

    def serve(route):
        assert route.request.method == 'GET', 'Unexpected upload/request method'
        url = route.request.url
        assert url.startswith('https://rummy.test/'), 'Unexpected network destination: '+url
        name = url.removeprefix('https://rummy.test/').split('?')[0] or 'index.html'
        target = root / name
        assert target.resolve().is_relative_to(root)
        requested.append(name)
        route.fulfill(status=200, body=target.read_bytes(), content_type={'.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css'}.get(target.suffix, 'text/plain'))

    context.route('**/*', serve)
    page = context.new_page()
    page.on('pageerror', lambda error: errors.append(str(error)))
    page.on('console', lambda message: console_errors.append(message.text) if message.type == 'error' else None)
    page.goto('https://rummy.test/')
    page.add_script_tag(url='https://rummy.test/recognition/reader-core.js')
    page.add_script_tag(url='https://rummy.test/recognition/reader-data.js')
    read = '''async url => { const image=new Image();image.src=url;await image.decode();return RummyReader.read(image); }'''
    if args.image:
        import hashlib
        data = args.image.read_bytes()
        uri = 'data:image/'+('png' if args.image.suffix.lower() == '.png' else 'jpeg')+';base64,'+base64.b64encode(data).decode()
        result = page.evaluate(read, uri)
        report = {args.image.stem: {'source': hashlib.sha256(data).hexdigest(), 'zone': args.zone, **result}}
        output = json.dumps(report, indent=2)
        if args.output:
            args.output.write_text(output+'\n')
        else:
            print(output)
    else:
        values = [(n, ['red', 'black', 'blue', 'orange'][(n-1) % 4]) for n in range(1, 14)]
        for touching in [False, True]:
            data = payload(fixture(values, touching))['buffer']
            result = page.evaluate(read, 'data:image/png;base64,'+base64.b64encode(data).decode())
            actual = [o['candidates'][0]['tile'] if o['candidates'] else None for o in result['observations']]
            expected = [{'number': n, 'color': c} for n, c in values]
            assert actual == expected, {'touching': touching, 'expected': expected, 'actual': actual, 'raw': result['observations']}
        # Actual browser import invokes the reader. No injected observations.
        page.locator('#edition').select_option('4')
        page.locator('#table-file').set_input_files(payload(Image.new('RGB', (300, 160), '#f8efdb')))
        page.get_by_role('button', name='Confirm complete table reading', exact=True).click()
        page.locator('#rack-file').set_input_files(payload(fixture([(10, 'red'), (10, 'blue'), (10, 'black')])))
        page.locator('#rack-photo .tile-editor').nth(2).wait_for()
        page.get_by_role('button', name='Find move', exact=True).click()
        assert 'confirmation' in page.locator('#status').inner_text()
        for _ in range(3):
            page.get_by_role('button', name='Confirm tile', exact=True).first.click()
        page.get_by_role('button', name='Confirm complete rack reading', exact=True).click()
        page.get_by_role('button', name='Find move', exact=True).click()
        page.wait_for_function("() => document.querySelector('#status').textContent.includes('3 rack tiles')")
        assert page.locator('#result [data-played="true"]').count() == 3
        # Figure-like joker art must stay an unconfirmed proposal.
        data = payload(fixture([(8, 'black'), (10, 'black')], joker=True))['buffer']
        result = page.evaluate(read, 'data:image/png;base64,'+base64.b64encode(data).decode())
        assert any(o['candidates'] and o['candidates'][0]['tile'].get('joker') for o in result['observations']), result
        assert not page.evaluate('sessions.rack.observations.some(o => !o.correction)'), 'Unexpected state mutation from a separate read'
        print('Bundled OCR: all 13 numbers/colors in separated and touching synthetic strips, teal ink, possible-joker uncertainty, actual file import, confirmation gate, and real-engine opening passed.')
    assert not errors, errors
    assert not console_errors, console_errors
    assert set(requested) <= {'index.html','style.css','engine/wasm_exec.js','engine/wasm_data.js','engine.js','photo-state.js','browser-reader.js','app.js','recognition/reader-core.js','recognition/reader-data.js'}, requested
    browser.close()
print('All requests were bundled GET assets; no photo upload, external model fetch, remote solver, or CSP violation. Not a real-photo or physical-iPhone accuracy claim.')
