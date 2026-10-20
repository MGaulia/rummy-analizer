"""Desktop browser integration of real Go/Wasm and correction UI.

Photo observations below are synthetic test input, NOT recognizer evaluation.
All assets are fulfilled in-process; any other request fails the test.
"""
from pathlib import Path
import base64
import io
from PIL import Image
from playwright.sync_api import sync_playwright

root = Path(__file__).resolve().parents[1] / 'phone' / 'web'
blank = io.BytesIO()
Image.new('RGB', (300, 160), '#faf3de').save(blank, format='PNG')
photo = 'data:image/png;base64,' + base64.b64encode(blank.getvalue()).decode()
requests = []
with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    context = browser.new_context(viewport={'width':390,'height':844}, offline=True)
    def serve(route):
        url = route.request.url
        requests.append(url)
        assert url.startswith('https://rummy.test/'), f'Unexpected network destination: {url}'
        target = root / url.removeprefix('https://rummy.test/').split('?')[0]
        if target == root:
            target = root / 'index.html'
        assert target.resolve().is_relative_to(root.resolve())
        kind = {'.html':'text/html','.js':'text/javascript','.css':'text/css','.wasm':'application/wasm'}.get(target.suffix,'application/octet-stream')
        route.fulfill(status=200,body=target.read_bytes(),content_type=kind)
    context.route('**/*', serve)
    page = context.new_page()
    errors = []
    page.on('pageerror', lambda error: errors.append(str(error)))
    page.goto('https://rummy.test/')
    page.wait_for_function("document.getElementById('status').textContent.includes('Local engine ready')")
    assert 'Redistributions in binary form' in page.locator('#licenses').text_content()
    page.select_option('#edition','4')
    page.evaluate('payload => acceptNativePhoto(payload)', {'zone':'table','source':'table-fixture','image':photo,'observations':[]})
    page.get_by_role('button',name='Confirm complete table reading',exact=True).click()
    # Raw proposals cannot be silently treated as confirmed values.
    observations=[]
    for i,tile in enumerate([{'color':'red','number':10},{'color':'blue','number':10},{'joker':True}]):
        observations.append({'text':str(tile),'rect':{'x':i/3,'y':0,'width':1/3,'height':1},'candidates':[{'tile':tile,'confidence':.9}]})
    page.evaluate('payload => acceptNativePhoto(payload)',{'zone':'rack','source':'rack-fixture','image':photo,'observations':observations})
    page.get_by_role('button',name='Find move',exact=True).click()
    assert 'confirmation' in page.locator('#status').inner_text()
    for _ in range(3):
        page.get_by_role('button',name='Confirm tile',exact=True).first.click()
    page.get_by_role('button',name='Confirm complete rack reading',exact=True).click()
    page.get_by_role('button',name='Find move',exact=True).click()
    page.wait_for_function("document.getElementById('status').textContent.includes('maximum rack tiles played')")
    assert 'primary upper bound' not in page.locator('#status').inner_text()
    assert '3 rack tiles' in page.locator('#status').inner_text()
    assert page.locator('#result [data-piece-id]').count() == 3
    assert page.locator('#result [data-played="true"]').count() == 3
    assert 'Joker as' in page.locator('#result').inner_html()
    page.screenshot(path='/tmp/rummy-phone-ui.png',full_page=True)
    # Editing a confirmed reading invalidates both the old result and its consent.
    page.locator('#rack-photo input[type=number]').first.fill('9')
    assert page.locator('#result svg').count()==0
    page.get_by_role('button',name='Find move',exact=True).click()
    assert 'confirmation' in page.locator('#status').inner_text()
    # Manual additions/corrections must not contaminate raw recognizer metrics.
    page.evaluate("sessions.rack.add({x:0,y:0,width:.1,height:.1})")
    report=page.evaluate("rummyObservationReport()")
    assert len(report['rack']['observations'])==3
    assert report['rack']['observations'][0]['candidates'][0]['tile']['number']==10
    assert not errors, errors
    assert all(url.startswith('https://rummy.test/') for url in requests)
    browser.close()
print('Desktop offline UI: real Wasm solve/render, separate synthetic photo corrections, confirmation gate, invalidation, and no external requests passed.')
