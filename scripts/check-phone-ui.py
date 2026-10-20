"""Desktop real Go/Wasm UI with synthetic photo observations, not recognition."""
from pathlib import Path
import base64
import io
from PIL import Image
from playwright.sync_api import sync_playwright

root = Path(__file__).resolve().parents[1] / 'phone/web'
blank = io.BytesIO()
Image.new('RGB', (300,160), '#faf3de').save(blank, format='PNG')
photo = 'data:image/png;base64,'+base64.b64encode(blank.getvalue()).decode()
requests, errors = [], []
with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    context = browser.new_context(viewport={'width':390,'height':844}, offline=True)
    def serve(route):
        url=route.request.url;requests.append(url)
        assert route.request.method=='GET' and url.startswith('https://rummy.test/'), url
        name=url.removeprefix('https://rummy.test/').split('?')[0] or 'index.html'
        target=root/name;assert target.resolve().is_relative_to(root)
        kind={'.html':'text/html','.js':'text/javascript','.css':'text/css','.png':'image/png'}
        route.fulfill(body=target.read_bytes(),content_type=kind.get(target.suffix,'text/plain'))
    context.route('**/*',serve)
    page=context.new_page();page.on('pageerror',lambda error:errors.append(str(error)))
    page.on('console',lambda message:errors.append(message.text) if message.type=='error' else None)
    page.goto('https://rummy.test/')
    page.wait_for_function("typeof rummyExecute==='function'")
    assert page.locator('#edition').input_value()=='4'
    assert page.locator('#budget,#drawn').count()==0
    assert 'Redistributions in binary form' in page.locator('#licenses').text_content()
    page.locator('#opened').check();page.locator('#opening').uncheck()
    assert page.locator('#opened').is_disabled() and page.locator('#opened').is_checked()
    page.locator('#opening').check();assert page.locator('#opened').is_enabled() and page.locator('#opened').is_checked()
    page.locator('#opened').uncheck()
    page.evaluate('payload=>acceptNativePhoto(payload)',{'zone':'table','source':'table-fixture','image':photo,'observations':[]})
    observations=[{'text':str(tile),'rect':{'x':i/3,'y':0,'width':1/3,'height':1},'candidates':[{'tile':tile,'confidence':.9}]} for i,tile in enumerate([
        {'color':'red','number':10},{'color':'blue','number':10},{'joker':True}])]
    page.evaluate('payload=>acceptNativePhoto(payload)',{'zone':'rack','source':'rack-fixture','image':photo,'observations':observations})
    assert page.get_by_role('button',name='Confirm tile',exact=True).count()==0
    page.get_by_role('button',name='Find move',exact=True).click()
    page.wait_for_function("document.querySelector('#status').textContent.includes('maximum rack tiles played')")
    assert '3 rack tiles' in page.locator('#status').inner_text()
    assert page.locator('#result [data-piece-id]').count()==3
    assert page.locator('#result .tile-art-played[fill="none"]').count()==3
    assert page.locator('#result .tile-art-caption').text_content().startswith('Joker as ')
    assert page.locator('#result .table-board').count()==1
    page.screenshot(path='/tmp/rummy-phone-ui.png',full_page=True)
    # Blank current drafts block a move, even when OCR had a valid old proposal.
    page.locator('#rack-photo input[type=number]').first.fill('')
    assert page.locator('#result .table-board').count()==0
    page.get_by_role('button',name='Find move',exact=True).click()
    assert 'Could not read rack tile' in page.locator('#status').inner_text()
    assert 'error' in page.locator('#status').get_attribute('class').split()
    page.locator('#rack-photo input[type=number]').first.fill('10')
    page.get_by_role('button',name='Find move',exact=True).click()
    page.wait_for_function("document.querySelector('#status').textContent.includes('3 rack tiles')")
    # Removed detections disappear, while raw report remains unmodified.
    page.locator('#rack-photo').get_by_role('button',name='Remove tile',exact=True).first.click()
    assert page.locator('#rack-photo .tile-editor').count()==2
    raw=page.evaluate('rummyObservationReport()')
    assert len(raw['rack']['observations'])==3 and raw['rack']['observations'][0]['candidates'][0]['tile']['number']==10
    ambiguous={'rect':{'x':0,'y':0,'width':1,'height':1},'candidates':[
        {'tile':{'color':'red','number':10},'confidence':.9},
        {'tile':{'color':'red','number':11},'confidence':.8}]}
    page.evaluate('payload=>acceptNativePhoto(payload)',{'zone':'rack','source':'ambiguous-rack','image':photo,'observations':[ambiguous]})
    assert page.locator('#rack-photo input[type=number]').input_value()==''
    assert page.locator('#rack-photo .tile-error').is_visible()
    page.get_by_role('button',name='Find move',exact=True).click()
    assert 'Could not read rack tile' in page.locator('#status').inner_text()
    # An actual reader failure cannot masquerade as a successfully empty table.
    page.evaluate('payload=>acceptNativePhoto(payload)',{'zone':'table','source':'failed-table','image':photo,'observations':[],'readError':'Test decoder failure'})
    page.get_by_role('button',name='Check table',exact=True).click()
    assert 'Could not read the table photo' in page.locator('#status').inner_text()
    assert page.locator('#result [data-piece-id]').count()==0
    assert not errors,errors
    browser.close()
print('Desktop offline UI: default4, dependent opening state, immediate real-engine solve, responsive exact artwork, invalid draft/read-error blocking and red-X/raw isolation passed. Synthetic observations, no phone accuracy claim.')
