"""Guard delayed engine readiness with synthetic photos and actual Go/Wasm.

The readiness promise is deliberately delayed after the bundled engine starts.
Photo observations are synthetic; this checks job locking and stale-input
rejection, not recognition accuracy, startup timing, or physical-phone behavior.
"""
import argparse
import base64
import io
import json
from pathlib import Path

from PIL import Image
from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--preview-directory', type=Path)
args = parser.parse_args()
root = args.preview_directory.resolve() if args.preview_directory else Path(__file__).resolve().parents[1] / 'phone/web'
buffer = io.BytesIO()
Image.new('RGB', (300, 160), '#faf3de').save(buffer, format='PNG')
photo = 'data:image/png;base64,' + base64.b64encode(buffer.getvalue()).decode()
results = {}

with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    for case in ('edit-during-startup', 'two-clicks-during-startup'):
        context = browser.new_context(offline=True)
        base = 'https://rummy.test/rummy-analizer/'
        errors = []

        def serve(route):
            assert route.request.method == 'GET' and route.request.url.startswith(base), route.request.url
            name = route.request.url.removeprefix(base).split('?')[0] or 'index.html'
            target = root / name
            assert target.resolve().is_relative_to(root)
            kind = {'.html':'text/html', '.js':'text/javascript', '.css':'text/css', '.png':'image/png'}
            route.fulfill(body=target.read_bytes(), content_type=kind.get(target.suffix, 'text/plain'))

        context.route('**/*', serve)
        page = context.new_page()
        page.on('pageerror', lambda error: errors.append(str(error)))
        page.on('console', lambda message: errors.append(message.text) if message.type == 'error' else None)
        page.goto(base)
        page.wait_for_function("typeof rummyExecute==='function'")
        page.evaluate('payload=>acceptNativePhoto(payload)', {
            'zone':'table', 'source':'startup-table', 'image':photo, 'observations':[]})
        observations = [{'candidates':[{'tile':{'color':color, 'number':10}, 'confidence':.9}]}
                        for color in ('red', 'blue', 'black')]
        page.evaluate('payload=>acceptNativePhoto(payload)', {
            'zone':'rack', 'source':'startup-rack', 'image':photo, 'observations':observations})
        page.evaluate('''()=>{
          window.startupCalls=[];
          window.startupOriginalExecute=rummyExecute;
          window.rummyExecute=async(id,raw)=>{
            const response=await startupOriginalExecute(id,raw);
            startupCalls.push({request:JSON.parse(raw),response:JSON.parse(response)});
            return response;
          };
          window.engineReady=new Promise(resolve=>window.releaseStartup=resolve);
          window.startupTask=execute('solve');
        }''')
        assert page.locator('#solve').is_disabled(), 'Solve must lock before awaiting engine readiness'
        assert page.locator('#validate').is_disabled(), 'Table validation must share the job lock'
        if case == 'edit-during-startup':
            page.locator('#rack-photo input[type=number]').first.fill('9')
        else:
            page.evaluate("()=>{window.startupTask2=execute('solve');}")
        page.evaluate('''async()=>{
          releaseStartup();
          await Promise.all([startupTask,window.startupTask2]);
        }''')
        calls = page.evaluate('startupCalls')
        rendered = page.locator('#result [data-piece-id]')
        if case == 'edit-during-startup':
            assert calls == [], 'An edit while awaiting startup must prevent the stale solver call'
            assert rendered.count() == 0, 'The earlier position must never render after an edit'
        else:
            assert len(calls) == 1, ('Two requests during startup must acquire one job', calls)
            request, response = calls[0]['request'], calls[0]['response']
            search = response['search']
            assert search['status'] == 'optimal' and search['verified']['valid'], response
            assert len(search['verified']['played']) == 3
            checked = page.evaluate('''async request=>JSON.parse(await startupOriginalExecute(
              'startup-independent-check',JSON.stringify(request)))''', {
                'operation':'check-move', 'settings':request['settings'], 'turn':request['turn'],
                'position':request['position'], 'move':search['move']})
            assert not checked.get('error') and checked['move']['valid'], checked
            assert checked['move']['after'] == search['verified']['after']
            assert checked['move']['played'] == search['verified']['played']
            assert rendered.count() == 3
            assert set(rendered.evaluate_all('(pieces)=>pieces.map(p=>p.dataset.pieceId)')) == set(checked['move']['played'])
        assert not page.locator('#solve').is_disabled() and page.locator('#cancel').is_disabled()
        page.wait_for_load_state('networkidle')
        assert not errors, errors
        results[case] = {'solver_calls':len(calls), 'rendered_occurrences':rendered.count()}
        context.close()
    browser.close()

print(json.dumps(results, sort_keys=True))
print('Delayed-readiness regression passed: immediate job lock, stale-input rejection, one solver call for repeated requests, and independent move validation. Desktop synthetic observations; bundled assets only.')
