"""Replay an external two-photo regression through the actual local UI/solver.

This case expects no rack-only opening, then an optimal move after opening.
Original photos stay outside Git and are imported as file-picker bytes; no
observations, solver replies or validation results are injected. Desktop only.
"""
import argparse
import json
from pathlib import Path

from playwright.sync_api import sync_playwright

from photo_regression import (
    asset_names, assert_expected_position, assert_reading, check_conservation,
    flatten, load_report_samples,
)

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--report', required=True, type=Path)
parser.add_argument('--expectations', required=True, type=Path)
parser.add_argument('--preview-directory', type=Path)
parser.add_argument('--expected-played', type=int, default=4)
parser.add_argument('--output', type=Path)
parser.add_argument('--screenshot', type=Path)
args = parser.parse_args()
if args.expected_played < 1:
    parser.error('--expected-played must be positive')
root = args.preview_directory.resolve() if args.preview_directory else Path(__file__).resolve().parents[1] / 'phone/web'
report = json.loads(args.report.read_text())
expected = json.loads(args.expectations.read_text())
samples = load_report_samples(report, expected)
allowed = asset_names(root)


requested, errors, outcomes = [], [], []
with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    context = browser.new_context(offline=True, viewport={'width':390,'height':844}, is_mobile=True, has_touch=True)
    base = 'https://rummy.test/rummy-analizer/'

    def serve(route):
        assert route.request.method == 'GET', 'Unexpected upload/request method'
        assert route.request.url.startswith(base), 'Unexpected external destination: ' + route.request.url
        name = route.request.url.removeprefix(base).split('?')[0] or 'index.html'
        assert name in allowed, 'Unexpected app resource: ' + name
        target = root / name
        assert target.resolve().is_relative_to(root), 'Asset escaped package directory'
        requested.append(name)
        kind = {'.html':'text/html','.js':'text/javascript','.css':'text/css','.png':'image/png'}
        route.fulfill(body=target.read_bytes(), content_type=kind.get(target.suffix,'text/plain'))

    context.route('**/*', serve)
    page = context.new_page()
    page.on('pageerror', lambda error: errors.append(str(error)))
    page.on('console', lambda message: errors.append(message.text) if message.type == 'error' else None)
    page.goto(base)
    if not page.evaluate('Boolean(globalThis.TesseractCore && globalThis.RUMMY_READER_MODEL)'):
        page.add_script_tag(url=base+'recognition/reader-core.js')
        page.add_script_tag(url=base+'recognition/reader-data.js')
    page.evaluate('''async()=>{
      await engineReady;
      const original=rummyExecute;
      window.readingFlowCalls=[];
      window.rummyExecute=async(id,raw)=>{
        const response=await original(id,raw);
        readingFlowCalls.push({request:JSON.parse(raw),response:JSON.parse(response)});
        return response;
      };
    }''')
    assert page.locator('#edition').input_value() == '4'
    assert page.locator('#opening').is_checked() and not page.locator('#opened').is_checked()
    for zone in ('table', 'rack'):
        sample = samples[zone]
        page.locator('#'+zone+'-file').set_input_files({'name':zone+'.'+sample['mime'].split('/')[-1], 'mimeType':sample['mime'], 'buffer':sample['data']})
        page.wait_for_function('({zone,source})=>sessions[zone]?.source===source', arg={'zone':zone,'source':sample['source']})
    observed = page.evaluate('rummyObservationReport()')
    for zone, sample in samples.items():
        assert_reading(observed[zone], sample['truth'])
    position = page.evaluate('positionFromPhotos(sessions.table,sessions.rack)')
    assert_expected_position(position, samples)

    def solve(label, expect_move):
        count = page.evaluate('readingFlowCalls.length')
        page.get_by_role('button', name='Find move', exact=True).click()
        page.wait_for_function('count=>readingFlowCalls.length>count&&!document.querySelector("#solve").disabled', arg=count)
        call = page.evaluate('readingFlowCalls.at(-1)')
        request, response = call['request'], call['response']
        assert request['operation'] == 'solve' and request['position'] == position
        assert not response.get('error'), response.get('error')
        search = response['search']
        outcome = {'case':label,'settings':request['settings'],'turn':request['turn'],
                   'status':search['status'],'nodes':search['nodes'],'upper_bound':search.get('upper_bound')}
        if not expect_move:
            assert search['status'] == 'no_move' and search.get('upper_bound') == 0, search
            assert '30-point opening' in page.locator('#status').inner_text()
            assert page.locator('#result [data-piece-id]').count() == 0
        else:
            assert search['status'] == 'optimal' and search['verified']['valid'], search
            played, after = search['verified']['played'], search['verified']['after']
            assert len(played) == args.expected_played, played
            assert len(after['rack']) == len(position['rack']) - args.expected_played
            check_conservation(position, after, played)
            checked = page.evaluate('''async request=>JSON.parse(await rummyExecute('independent-reading-check',JSON.stringify(request)))''',
                {'operation':'check-move','settings':request['settings'],'turn':request['turn'], 'position':position,'move':search['move']})
            assert not checked.get('error') and checked['move']['valid'], checked
            assert checked['move']['after'] == after and checked['move']['played'] == played
            rendered = page.locator('#result [data-piece-id]').evaluate_all('(groups)=>groups.map(g=>({id:g.dataset.pieceId,played:g.dataset.played}))')
            assert len(rendered) == len(flatten(position)) and {p['id'] for p in rendered} == {p['id'] for p in flatten(position)}
            assert {p['id'] for p in rendered if p['played']=='true'} == set(played)
            outcome.update(played_count=len(played),remaining_count=len(after['rack']),independent_move_valid=True)
        outcomes.append(outcome)

    solve('opening-required-unopened', False)
    page.locator('#opened').check()
    solve('already-opened', True)
    page.locator('#opening').uncheck()
    assert page.locator('#opened').is_disabled() and page.locator('#opened').is_checked()
    solve('opening-disabled', True)
    page.wait_for_function('()=>[...document.querySelectorAll("#result image")].length>0')
    if args.screenshot:
        page.locator('#result').screenshot(path=str(args.screenshot))
    page.locator('#opening').check()
    assert page.locator('#opened').is_enabled() and page.locator('#opened').is_checked(), 'Player history must survive a rule toggle'
    assert not errors, errors
    result = {'preview_source':page.locator('meta[name="rummy-preview-source"]').get_attribute('content') if page.locator('meta[name="rummy-preview-source"]').count() else None,
              'browser':browser.version,'photos':{zone:{'source':sample['source'],'tiles':len(sample['truth']['tiles'])} for zone,sample in samples.items()},
              'outcomes':outcomes,'bundled_get_assets':sorted(set(requested)),
              'limitations':'Actual external photo bytes and local Go/Wasm; desktop Chromium, not physical-iPhone acceptance or general recognition accuracy.'}
    if args.output:
        args.output.write_text(json.dumps(result,indent=2)+'\n')
    browser.close()
print('Actual report UI flow passed: all original photo readings/occurrences, explicit no-opening result, optimal moves after opening/when disabled, and independent move validation; no confirmation clicks, uploads or remote solver.')
