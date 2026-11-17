"""Replay a private table/rack report through the real offline UI and Go solver.

Uses the report's edition/opening/player state and independently supplied tile
labels and expected search outcome. Images enter through the real file picker;
no observations, positions, solver responses or validation results are injected.
Optional --corrections explicitly exercises manual UI editing of hash-bound
unreadable occurrences; that result is not automatic photo recognition.
"""
import argparse
import json
from pathlib import Path

from playwright.sync_api import sync_playwright

from photo_regression import (
    asset_names, assert_expected_position, assert_reading, check_conservation,
    flatten, load_report_samples, validate_corrections,
)

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--report', required=True, type=Path)
parser.add_argument('--expectations', required=True, type=Path)
parser.add_argument('--expected-status', required=True, choices=['optimal', 'feasible_incomplete', 'no_move'])
parser.add_argument('--expected-played', required=True, type=int)
parser.add_argument('--preview-directory', type=Path)
parser.add_argument('--timeout-ms', type=int, default=180000)
parser.add_argument('--output', type=Path, help='Private result JSON destination; images are never copied')
parser.add_argument('--screenshot', type=Path, help='Optional private final-board capture destination')
parser.add_argument('--corrections', type=Path, help='External hash/rectangle-bound edits for every raw unreadable tile; uses real UI controls')
args = parser.parse_args()
if args.timeout_ms <= 0 or args.expected_played < 0:
    parser.error('timeout-ms must be positive and expected-played nonnegative')
if (args.expected_status == 'no_move') != (args.expected_played == 0):
    parser.error('no_move requires --expected-played=0; a verified move must play at least one tile')
root = args.preview_directory.resolve() if args.preview_directory else Path(__file__).resolve().parents[1] / 'phone/web'
report = json.loads(args.report.read_text())
expected = json.loads(args.expectations.read_text())
samples = load_report_samples(report, expected)
state = report.get('state') or {}
settings = state.get('settings') or {}
settings = {'edition': settings.get('edition', 4), 'opening_required': settings.get('opening_required', True)}
saved_turn = state.get('turn') or {}
assert not saved_turn.get('drawn', False), 'The phone UI cannot replay an already-drawn turn'
turn = {'opened': saved_turn.get('opened', False)}
assert settings['edition'] in (4, 6) and type(settings['edition']) is int, 'Invalid report edition'
assert type(settings['opening_required']) is bool and type(turn['opened']) is bool, 'Invalid report opening state'
allowed = asset_names(root)
requested, errors = [], []
base = 'https://rummy.test/rummy-analizer/'

with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    context = browser.new_context(offline=True, viewport={'width': 390, 'height': 844}, is_mobile=True, has_touch=True)

    def serve(route):
        request = route.request
        assert request.method == 'GET', 'Unexpected upload/request method'
        assert request.url.startswith(base), 'Unexpected external destination: ' + request.url
        name = request.url.removeprefix(base).split('?')[0] or 'index.html'
        assert name in allowed, 'Unexpected app resource: ' + name
        target = root / name
        assert target.resolve().is_relative_to(root), 'Asset escaped package directory'
        requested.append(name)
        kind = {'.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css', '.png': 'image/png'}
        route.fulfill(body=target.read_bytes(), content_type=kind.get(target.suffix, 'text/plain'))

    context.route('**/*', serve)
    page = context.new_page()
    page.set_default_timeout(args.timeout_ms)
    page.on('pageerror', lambda error: errors.append(str(error)))
    page.on('console', lambda message: errors.append(message.text) if message.type == 'error' else None)
    page.goto(base)
    if not page.evaluate('Boolean(globalThis.TesseractCore && globalThis.RUMMY_READER_MODEL)'):
        page.add_script_tag(url=base+'recognition/reader-core.js')
        page.add_script_tag(url=base+'recognition/reader-data.js')
    page.evaluate('''async()=>{
      await engineReady;
      const actual=rummyExecute;
      window.reportFlowCalls=[];
      window.rummyExecute=async(id,raw)=>{
        const answer=await actual(id,raw);
        reportFlowCalls.push({request:JSON.parse(raw),response:JSON.parse(answer)});
        return answer;
      };
    }''')
    page.locator('#edition').select_option(str(settings['edition']))
    # Set opened while the checkbox is enabled, then restore the separate rule.
    page.locator('#opening').check()
    page.locator('#opened').set_checked(turn['opened'])
    page.locator('#opening').set_checked(settings['opening_required'])
    for zone in ('table', 'rack'):
        sample = samples[zone]
        page.locator('#'+zone+'-file').set_input_files({
            'name': zone+'.'+sample['mime'].split('/')[-1],
            'mimeType': sample['mime'], 'buffer': sample['data']})
        page.wait_for_function('({zone,source})=>sessions[zone]?.source===source',
                               arg={'zone': zone, 'source': sample['source']}, timeout=args.timeout_ms)
    observed = page.evaluate('rummyObservationReport()')
    raw_unreadable = {zone: sum(not o['candidates'] for o in result['observations']) for zone, result in observed.items()}
    manual_edits = {zone: 0 for zone in observed}
    readings = observed
    if args.corrections:
        plan = validate_corrections(observed, json.loads(args.corrections.read_text()))
        assert plan, '--corrections must contain at least one unreadable occurrence'
        assert page.locator('#status.error').count() == 1 and 'could not read' in page.locator('#status').inner_text().lower(), 'Unreadable photos did not show a parse error'
        assert page.locator('.tile-error:visible').count() == sum(raw_unreadable.values()), 'Unreadable tile indicators do not match raw OCR'
        assert page.evaluate('reportFlowCalls.filter(c=>c.request.operation==="solve").length') == 0
        page.get_by_role('button', name='Find move', exact=True).click()
        assert page.locator('#status.error').count() == 1 and 'could not read' in page.locator('#status').inner_text().lower(), 'Find move did not preserve the parse error'
        assert page.evaluate('reportFlowCalls.filter(c=>c.request.operation==="solve").length') == 0, 'Unreadable input reached the solver'
        for edit in plan:
            row = page.locator('#'+edit['zone']+'-photo .tile-editor').nth(edit['index'])
            tile = edit['tile']
            row.get_by_label('Tile color', exact=True).select_option('joker' if tile.get('joker') else tile['color'])
            if not tile.get('joker'):
                row.get_by_label('Tile number', exact=True).fill(str(tile['number']))
            manual_edits[edit['zone']] += 1
        assert page.evaluate('rummyObservationReport()') == observed, 'Manual UI edits changed the original raw OCR report'
        edited = page.evaluate('''()=>Object.fromEntries(Object.entries(sessions).map(([zone,session])=>[zone,{
          source:session.source,zone,observations:session.observations.map(o=>({
            raw:o.raw,correction:o.correction,excluded:o.excluded,manual:o.manual,
            reading:session.reading(o.id)
          }))
        }]))''')
        listed = {(edit['zone'], edit['index']): edit['tile'] for edit in plan}
        readings = {}
        for zone, result in edited.items():
            assert result['source'] == observed[zone]['source'] and len(result['observations']) == len(observed[zone]['observations'])
            effective = []
            for index, item in enumerate(result['observations']):
                raw, reading = item['raw'], item['reading']
                assert raw == observed[zone]['observations'][index] and not item['excluded'] and not item['manual'], 'An original occurrence changed or was removed'
                tile = listed.get((zone, index))
                if tile is None:
                    assert item['correction'] is None, 'An unlisted observation was edited'
                else:
                    assert item['correction'] is not None and reading['tile'] == tile, 'UI correction did not reach the intended occurrence'
                    assert reading['order'] == raw['suggestedOrder'] and (zone != 'table' or reading['meld'] == raw['suggestedMeld']), 'A tile correction changed table grouping/order'
                effective.append({'candidates': [{'tile': reading['tile']}], 'suggestedMeld': reading['meld'], 'suggestedOrder': reading['order']})
            readings[zone] = {'source': result['source'], 'zone': zone, 'observations': effective}
    for zone, sample in samples.items():
        assert_reading(readings[zone], sample['truth'])
    position = page.evaluate('positionFromPhotos(sessions.table,sessions.rack)')
    assert_expected_position(position, samples)
    assert page.get_by_role('button', name='Confirm tile', exact=True).count() == 0
    if args.corrections:
        print(f"Photo correction checks passed: {sum(raw_unreadable.values())} raw unreadable tiles, {sum(manual_edits.values())} explicit manual UI edits; original raw readings preserved, all effective readings/position exact, and solve blocked before editing.", flush=True)
    page.get_by_role('button', name='Find move', exact=True).click()
    page.wait_for_function('()=>reportFlowCalls.some(c=>c.request.operation==="solve")&&!document.querySelector("#solve").disabled', timeout=args.timeout_ms)
    call = page.evaluate('reportFlowCalls.find(c=>c.request.operation==="solve")')
    request, response = call['request'], call['response']
    assert request['settings'] == settings and request['turn'] == turn, 'Report state was not applied through UI controls'
    assert request['position'] == position and not response.get('error'), response.get('error')
    search = response['search']
    assert search['status'] == args.expected_status, ('Unexpected search status', search)
    result = {
        'preview_source': page.locator('meta[name="rummy-preview-source"]').get_attribute('content') if page.locator('meta[name="rummy-preview-source"]').count() else None,
        'browser': browser.version, 'settings': settings, 'turn': turn,
        'photos': {zone: {'source': sample['source'], 'tiles': len(sample['truth']['tiles']),
                          'comparison': sample['truth'].get('comparison', 'ordered'),
                          'raw_unreadable': raw_unreadable[zone], 'manual_edits': manual_edits[zone]} for zone, sample in samples.items()},
        'recognition_mode': 'manual_corrections' if args.corrections else 'automatic',
        'raw_unreadable': sum(raw_unreadable.values()), 'manual_edits': sum(manual_edits.values()),
        'blocked_solve_before_corrections': bool(args.corrections),
        'status': search['status'], 'nodes': search['nodes'], 'reason': search.get('reason'),
        'upper_bound': search.get('upper_bound'),
    }
    if args.expected_status == 'no_move':
        assert not search.get('move') and not search.get('verified'), 'No-move result carried a proposed move'
        assert search.get('upper_bound') == 0, 'No-move result lacks an exhaustive zero bound'
        assert 'node_limit' not in search.get('reason', '') and 'canceled' not in search.get('reason', ''), 'Resource stop was mislabeled no_move'
        assert page.locator('#result [data-piece-id]').count() == 0
        result.update(played_count=0, remaining_count=len(position['rack']))
    else:
        assert search['verified']['valid'], 'Proposed move was not validated'
        played, after = search['verified']['played'], search['verified']['after']
        assert len(played) == args.expected_played, ('Wrong rack-tile objective', played)
        assert len(after['rack']) == len(position['rack'])-args.expected_played
        check_conservation(position, after, played)
        checked = page.evaluate('''async request=>JSON.parse(await rummyExecute('independent-report-check',JSON.stringify(request)))''', {
            'operation': 'check-move', 'settings': settings, 'turn': turn,
            'position': position, 'move': search['move']})
        assert not checked.get('error') and checked['move']['valid'], checked
        assert checked['move']['after'] == after and checked['move']['played'] == played
        if args.expected_status == 'optimal':
            assert search.get('upper_bound') == len(played), 'Optimal result lacks a matching primary bound'
        else:
            assert 'search incomplete' in page.locator('#status').inner_text().lower(), 'Incomplete move was labeled optimal'
            assert 'maximum rack tiles played' not in page.locator('#status').inner_text().lower()
        rendered = page.locator('#result [data-piece-id]').evaluate_all('''groups=>groups.map(g=>({
          id:g.dataset.pieceId,played:g.dataset.played,joker:g.dataset.joker,
          color:g.dataset.tileColor,number:Number(g.dataset.tileNumber),
          roleColor:g.dataset.roleColor,roleNumber:Number(g.dataset.roleNumber)
        }))''')
        wanted = []
        for piece in flatten(after):
            tile, role = piece['tile'], piece.get('as') or {}
            joker = tile.get('joker') is True
            wanted.append({'id': piece['id'], 'played': str(piece['id'] in played).lower(),
                           'joker': str(joker).lower(), 'color': '' if joker else tile['color'],
                           'number': 0 if joker else tile['number'], 'roleColor': role.get('color', ''),
                           'roleNumber': role.get('number', 0)})
        assert rendered == wanted, 'Rendered identities, values, rack markers or joker roles changed'
        rows = page.locator('#result .meld-row, #result .rack-row').evaluate_all('(rows)=>rows.map(row=>[...row.querySelectorAll("[data-piece-id]")].map(g=>g.dataset.pieceId))')
        assert rows == [[p['id'] for p in meld] for meld in after['table']] + [[p['id'] for p in after['rack']]], 'Rendered meld/rack boundaries or order changed'
        captions = [' '.join(text.split()) for text in page.locator('#result .joker-role').all_text_contents()]
        expected_captions = [f"Joker as {p['as']['color']} {p['as']['number']}" if p.get('as') else 'Joker' for p in flatten(after) if p['tile'].get('joker')]
        assert captions == expected_captions, 'Visible joker captions changed'
        result.update(played_count=len(played), remaining_count=len(after['rack']), independent_move_valid=True)
        if args.screenshot:
            page.locator('#result').screenshot(path=str(args.screenshot))
    assert not errors, errors
    result['bundled_get_assets'] = sorted(set(requested))
    result['limitations'] = 'Original external images and actual local OCR/Go; desktop Chromium only, not phone performance or representative recognition accuracy.'
    if args.corrections:
        result['limitations'] += f" This replay required {sum(manual_edits.values())} explicit manual UI corrections and is not an automatic OCR success."
    if args.output:
        args.output.write_text(json.dumps(result, indent=2)+'\n')
    browser.close()
correction_note = f"{sum(manual_edits.values())} explicit manual UI corrections" if args.corrections else 'automatic photo readings'
print(f"Original report flow passed: {sum(len(s['truth']['tiles']) for s in samples.values())} occurrences, {correction_note}, {args.expected_status}, {args.expected_played} played; report state/readings, applicable independent move/render checks and bundled-only requests.")
