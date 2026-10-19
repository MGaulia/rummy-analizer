"""Check the generated Pages package in desktop Chromium at a phone viewport.

Uses synthetic images and MANUAL tile entry, not simulated recognition.
This is not an iPhone/Safari compatibility or real-photo accuracy test.
"""
import argparse
import io
import json
from pathlib import Path

from PIL import Image
from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('preview_directory', type=Path)
args = parser.parse_args()
root = args.preview_directory.resolve()
metadata = json.loads((root / 'preview-build.json').read_text())
base = 'https://rummy.test/rummy-analizer/'
requests = []
errors = []


def image_file(name, color):
    buffer = io.BytesIO()
    Image.new('RGB', (300, 160), color).save(buffer, format='PNG')
    return {'name': name, 'mimeType': 'image/png', 'buffer': buffer.getvalue()}


with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium',
                                headless=True, args=['--no-sandbox'])
    context = browser.new_context(viewport={'width': 390, 'height': 844},
                                  is_mobile=True, has_touch=True, offline=True)

    def serve(route):
        request = route.request
        requests.append(request.url)
        assert request.method == 'GET', 'Unexpected outgoing data: ' + request.method
        assert request.url.startswith(base), 'Unexpected destination: ' + request.url
        name = request.url.removeprefix(base).split('?')[0] or 'index.html'
        assert name in metadata['files_sha256'], 'Unexpected asset: ' + name
        target = root / name
        kind = {'.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css'}
        route.fulfill(status=200, body=target.read_bytes(),
                      content_type=kind.get(target.suffix, 'text/plain'))

    context.route('**/*', serve)
    page = context.new_page()
    page.on('pageerror', lambda error: errors.append(str(error)))
    page.goto(base)
    page.wait_for_function("document.querySelector('#status').textContent.includes('Local engine ready')")
    assert page.get_by_role('heading', name='Safari test preview').count() == 1
    assert page.locator('[data-camera]').count() == 0
    assert metadata['source_commit'][:12] in page.content()
    page.select_option('#edition', '4')
    page.locator('#table-file').set_input_files(image_file('table.png', '#f5eada'))
    page.get_by_role('button', name='Confirm complete table reading', exact=True).click()
    page.get_by_role('button', name='Check table', exact=True).click()
    page.wait_for_function("document.querySelector('#status').textContent === 'Confirmed table is valid.'")
    page.locator('#rack-file').set_input_files(image_file('rack.png', '#dce8ec'))
    page.locator('#rack-photo .photo-preview').wait_for()
    for color in ['red', 'blue', 'black']:
        page.locator('#rack-photo .photo-preview').click(position={'x': 40, 'y': 40})
        editor = page.locator('#rack-photo .tile-editor').last
        editor.get_by_label('Tile color', exact=True).select_option(color)
        editor.get_by_label('Tile number', exact=True).fill('10')
        editor.get_by_role('button', name='Confirm tile', exact=True).click()
    page.get_by_role('button', name='Confirm complete rack reading', exact=True).click()
    page.get_by_role('button', name='Find move', exact=True).click()
    page.wait_for_function("document.querySelector('#status').textContent.includes('maximum rack tiles played')")
    assert '3 rack tiles' in page.locator('#status').inner_text()
    assert page.locator('#result [data-played="true"]').count() == 3
    assert page.locator('#result [data-piece-id]').count() == 3
    assert not errors, errors
    assert len(requests) >= 7, requests
    browser.close()

print('Preview package: nested Pages path, manual file import/correction, empty-table validation, '
      'real Wasm three-tile opening and exact rendering passed; no uploads or remote solver requests.')
print('Desktop Chromium at a mobile viewport; physical iPhone/Safari remains untested.')
