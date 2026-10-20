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
import hashlib
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont
from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
input_group=parser.add_mutually_exclusive_group()
input_group.add_argument('--image', type=Path)
input_group.add_argument('--report', type=Path, help='Rerun raw local photos from an explicitly saved phone reading report')
parser.add_argument('--zone', choices=['table', 'rack'], default='rack')
parser.add_argument('--output', type=Path)
parser.add_argument('--expectations', type=Path, help='External, manually labeled exact photo regression (source hash, zone and ordered tiles/melds)')
parser.add_argument('--preview-directory', type=Path, help='Check the exact generated Pages package instead of source assets')
args = parser.parse_args()
if args.expectations and not (args.image or args.report):
    parser.error('--expectations requires --image or --report')
root = args.preview_directory.resolve() if args.preview_directory else Path(__file__).resolve().parents[1] / 'phone/web'
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
    if not page.evaluate('Boolean(globalThis.TesseractCore && globalThis.RUMMY_READER_MODEL)'):
        page.add_script_tag(url='https://rummy.test/recognition/reader-core.js')
        page.add_script_tag(url='https://rummy.test/recognition/reader-data.js')
    read = '''async url => { const image=new Image();image.src=url;await image.decode();return RummyReader.read(image); }'''
    if args.image or args.report:
        if args.image:
            data=args.image.read_bytes()
            uri='data:image/'+('png' if args.image.suffix.lower()=='.png' else 'jpeg')+';base64,'+base64.b64encode(data).decode()
            samples={args.image.stem:{'source':hashlib.sha256(data).hexdigest(),'zone':args.zone,'image':uri}}
        else:
            report=json.loads(args.report.read_text())
            assert report['version']==1 and report['kind']=='rummy-photo-reading-report', 'Unknown report format'
            samples=report['photos']
        predictions={}
        for ident,sample in samples.items():
            assert sample['zone'] in ['table','rack']
            uri=sample['image']
            assert uri.startswith('data:image/') and ';base64,' in uri, 'Report must contain local image bytes'
            data=base64.b64decode(uri.split(';base64,',1)[1],validate=True)
            assert hashlib.sha256(data).hexdigest()==sample['source'], 'Photo checksum mismatch'
            result=page.evaluate(read,uri)
            predictions[ident]={'source':sample['source'],'zone':sample['zone'],**result}
        output = json.dumps(predictions, indent=2)
        if args.output:
            args.output.write_text(output+'\n')
        else:
            print(output)
        if args.expectations:
            expected=json.loads(args.expectations.read_text())
            assert set(predictions)==set(expected), 'Regression photo set changed'
            for ident,result in predictions.items():
                truth=expected[ident]
                assert result['source']==truth['source'] and result['zone']==truth['zone'], ('Regression source mismatch',ident)
                actual=[{'tile':o['candidates'][0]['tile'] if o['candidates'] else None,
                         'meld':o['suggestedMeld'],'order':o['suggestedOrder']} for o in result['observations']]
                assert actual==truth['tiles'], ('Exact photo regression failed',ident,{'expected':truth['tiles'],'actual':actual})
            print('Exact external photo regression passed: '+str(len(predictions))+' photos, '+str(sum(len(t['tiles']) for t in expected.values()))+' ordered tiles/melds; source hashes matched.')
    else:
        values = [(n, ['red', 'black', 'blue', 'orange'][(n-1) % 4]) for n in range(1, 14)]
        for touching in [False, True]:
            data = payload(fixture(values, touching))['buffer']
            result = page.evaluate(read, 'data:image/png;base64,'+base64.b64encode(data).decode())
            actual = [o['candidates'][0]['tile'] if o['candidates'] else None for o in result['observations']]
            expected = [{'number': n, 'color': c} for n, c in values]
            assert actual == expected, {'touching': touching, 'expected': expected, 'actual': actual, 'raw': result['observations']}
        # Spatial suppression must preserve separate physical copies, even when
        # their values/colors match and their faces touch.
        data = payload(fixture([(11, 'red'), (11, 'red'), (10, 'black')], touching=True))['buffer']
        result = page.evaluate(read, 'data:image/png;base64,'+base64.b64encode(data).decode())
        assert [o['candidates'][0]['tile'] for o in result['observations']] == [
            {'number': 11, 'color': 'red'}, {'number': 11, 'color': 'red'}, {'number': 10, 'color': 'black'}], result
        # A tilted tile can let a narrow, dark background streak into the face
        # crop. Reproduce the cross-color contamination seen in a private photo;
        # this generated image is a mechanism regression, not that real photo.
        image = fixture([(11, 'red')])
        ImageDraw.Draw(image).polygon([(68, 130), (71, 130), (68, 245)], fill='#080808')
        data = payload(image)['buffer']
        result = page.evaluate(read, 'data:image/png;base64,'+base64.b64encode(data).decode())
        assert len(result['observations']) == 1, result
        assert [c['tile'] for c in result['observations'][0]['candidates']] == [
            {'number': 11, 'color': 'red'}], ('A dark background streak must not obscure red 11', result)
        # Actual browser import invokes the reader. No injected observations.
        page.locator('#edition').select_option('4')
        page.locator('#table-file').set_input_files(payload(Image.new('RGB', (300, 160), '#f8efdb')))
        page.get_by_role('button', name='Confirm complete table reading', exact=True).click()
        rack_file = payload(fixture([(10, 'red'), (10, 'blue'), (10, 'black')]))
        for _ in range(2):
            page.locator('#rack-file').set_input_files(rack_file)
            page.locator('#rack-photo .tile-editor').nth(2).wait_for()
            assert page.locator('#rack-photo .tile-editor').count() == 3, 'Reimporting the same photo appended duplicate readings'
        page.get_by_role('button', name='Find move', exact=True).click()
        assert 'confirmation' in page.locator('#status').inner_text()
        for _ in range(3):
            page.get_by_role('button', name='Confirm tile', exact=True).first.click()
        page.get_by_role('button', name='Confirm complete rack reading', exact=True).click()
        page.get_by_role('button', name='Find move', exact=True).click()
        page.wait_for_function("() => document.querySelector('#status').textContent.includes('3 rack tiles')")
        assert page.locator('#result [data-played="true"]').count() == 3
        # Fault-injection regression, NOT an OCR accuracy measurement: a single
        # page-segmentation mode reports a confident partial "1" for visible "12".
        # Other modes still run the actual engine. Image geometry must prevent
        # the partial reading from silently winning because of its score.
        data = payload(fixture([(12, 'orange')]))['buffer']
        result = page.evaluate('''async url => {
          const {api}=await RummyReader.ready();
          const set=api.SetVariable.bind(api),text=api.GetUTF8Text.bind(api),score=api.MeanTextConf.bind(api);
          let mode='';
          api.SetVariable=(name,value)=>{if(name==='tessedit_pageseg_mode')mode=value;return set(name,value)};
          api.GetUTF8Text=()=>mode==='7'?'1\\n':text();
          api.MeanTextConf=()=>mode==='7'?90:score();
          try {const image=new Image();image.src=url;await image.decode();return await RummyReader.read(image);}
          finally {api.SetVariable=set;api.GetUTF8Text=text;api.MeanTextConf=score;}
        }''', 'data:image/png;base64,'+base64.b64encode(data).decode())
        assert result['observations'][0]['candidates'][0]['tile']['number'] == 12, ('A high-scoring partial digit cannot override two visible digits', result)
        # Figure-like joker art must stay an unconfirmed proposal.
        data = payload(fixture([(8, 'black'), (10, 'black')], joker=True))['buffer']
        result = page.evaluate(read, 'data:image/png;base64,'+base64.b64encode(data).decode())
        assert len(result['observations']) == 3, ('A light region inside a joker face must not become a second tile', result)
        assert any(o['candidates'] and o['candidates'][0]['tile'].get('joker') for o in result['observations']), result
        assert not page.evaluate('sessions.rack.observations.some(o => !o.correction)'), 'Unexpected state mutation from a separate read'
        page.locator('#reading-help summary').click()
        with page.expect_download() as download:
            page.get_by_role('button', name='Save reading report', exact=True).click()
        saved=json.loads(Path(download.value.path()).read_text())
        assert saved['kind']=='rummy-photo-reading-report' and saved['version']==1
        assert set(saved['photos'])=={'table','rack'}
        for zone,sample in saved['photos'].items():
            raw=base64.b64decode(sample['image'].split(',',1)[1])
            assert hashlib.sha256(raw).hexdigest()==sample['source']
            assert sample['zone']==zone and 'recognizer' in sample
            assert all('correction' not in observation for observation in sample['observations'])
        assert len(saved['photos']['rack']['observations'])==3
        print('Bundled OCR: all 13 synthetic numbers/colors, touching and identical tiles, cross-color background rejection, nested-joker suppression, high-score truncation fault injection, repeated file import, confirmation, and real-engine opening passed.')
    assert not errors, errors
    assert not console_errors, console_errors
    assert set(requested) <= {'index.html','style.css','engine/wasm_exec.js','engine/wasm_data.js','engine.js','photo-state.js','browser-reader.js','app.js','recognition/reader-core.js','recognition/reader-data.js'}, requested
    browser.close()
print('All requests were bundled GET assets; no photo upload, external model fetch, remote solver, or CSP violation. Desktop execution; no general accuracy or physical-iPhone acceptance claim.')
