"""Check actual model sprites on exact Go diagrams, offline in desktop Chromium.

Uses a synthetic position, not recognition or physical-phone acceptance.
"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image
from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--preview-directory', type=Path)
parser.add_argument('--screenshot', type=Path)
args = parser.parse_args()
root = args.preview_directory.resolve() if args.preview_directory else Path(__file__).resolve().parents[1] / 'phone/web'
atlas = root / 'tile-art/tiles-atlas.png'
assert hashlib.sha256(atlas.read_bytes()).hexdigest() == '37dfd3f03bd4d6281c4f359c10f2cdd06adbe716d274bfed1011496d332323a6'
manifest = json.loads((root / 'tile-art/tiles-atlas.json').read_text())
assert manifest['license'] == 'CC-BY-SA-4.0' and manifest['creator'] == 'KlaasYK'
assert len(manifest['sprites']) == 56
with Image.open(atlas) as image:
    assert image.mode == 'RGBA' and image.size == (2688, 1152)
    assert image.getchannel('A').getextrema() == (0, 255)

def numbered(ident, color, number):
    return {'id': ident, 'tile': {'color': color, 'number': number}}

def joker(ident, color=None, number=None):
    result = {'id': ident, 'tile': {'joker': True}}
    if color:
        result['as'] = {'color': color, 'number': number}
    return result

position = {'table': [
    [numbered('b1','blue',1), numbered('b2','blue',2), joker('j-blue','blue',3)],
    [numbered('r12','red',12), numbered('b12','blue',12), joker('j-black','black',12)],
    [numbered('o11','orange',11), joker('j-orange','orange',12), numbered('o13','orange',13)],
], 'rack': [numbered(f'{c}-{n}',c,n) for c in ['red','blue','black','orange'] for n in range(1,14)] +
    [numbered('duplicate-"<&','red',11), joker('rack-joker')]}
expected_ids = [p['id'] for meld in position['table'] for p in meld] + [p['id'] for p in position['rack']]
errors, requested = [], []
with sync_playwright() as p:
    browser = p.chromium.launch(executable_path='/usr/bin/chromium', headless=True, args=['--no-sandbox'])
    context = browser.new_context(offline=True, viewport={'width':1200,'height':1000})
    base = 'https://rummy.test/rummy-analizer/'
    def serve(route):
        assert route.request.method == 'GET' and route.request.url.startswith(base), route.request.url
        name = route.request.url.removeprefix(base).split('?')[0] or 'index.html'
        target = root / name
        assert target.resolve().is_relative_to(root)
        requested.append(name)
        kind = {'.html':'text/html','.js':'text/javascript','.css':'text/css','.png':'image/png'}
        route.fulfill(body=target.read_bytes(), content_type=kind.get(target.suffix,'text/plain'))
    context.route('**/*', serve)
    page = context.new_page()
    page.on('pageerror', lambda error: errors.append(str(error)))
    page.on('console', lambda message: errors.append(message.text) if message.type == 'error' else None)
    page.goto(base)
    result = page.evaluate('''async position => {
      await engineReady;
      const response=JSON.parse(await rummyExecute('art-test',JSON.stringify({operation:'validate',settings:{edition:6},position})));
      if(response.error)throw new Error(JSON.stringify(response));
      const svg=new DOMParser().parseFromString(response.svg.replace(/<style>[\\s\\S]*?<\\/style>/g,''),'image/svg+xml').documentElement;
      const before={viewBox:svg.getAttribute('viewBox'),labels:[...svg.querySelectorAll('[data-piece-id]>title')].map(t=>t.textContent)};
      RummyTileArt.decorate(svg);RummyTileArt.decorate(svg);
      document.querySelector('#result').replaceChildren(document.importNode(svg,true));
      const shown=document.querySelector('#result>svg');
      return {before,viewBox:shown.getAttribute('viewBox'),pieces:[...shown.querySelectorAll('[data-piece-id]')].map(g=>({
        id:g.dataset.pieceId,title:g.querySelector('title').textContent,label:g.getAttribute('aria-label'),
        sprites:g.querySelectorAll('.tile-art-sprite').length,cell:g.querySelector('.tile-art-sprite').getAttribute('viewBox'),
        fillRects:[...g.querySelectorAll('rect')].filter(r=>r.getAttribute('fill')!=='none').length,
        caption:g.querySelector('.tile-art-caption')?.textContent||'',
        width:g.querySelector('.tile-art-sprite').width.baseVal.value,
        height:g.querySelector('.tile-art-sprite').height.baseVal.value,
        minWidth:getComputedStyle(g.querySelector('.tile-art-sprite')).minWidth
      }))};
    }''', position)
    assert result['viewBox'] == result['before']['viewBox']
    pieces = result['pieces']
    assert [g['id'] for g in pieces] == expected_ids
    assert [g['title'] for g in pieces] == result['before']['labels']
    assert all(g['sprites'] == 1 and g['fillRects'] == 0 and g['label'] == g['title'] for g in pieces)
    assert all(0 < g['width'] < 70 and g['minWidth'] != '1020px' for g in pieces), 'CSS must not expand nested sprite SVGs'
    assert all(g['height'] == 76 for g in pieces), 'Jokers and numbered physical tiles must have the same size'
    by_id = {g['id']:g for g in pieces}
    for row,color in enumerate(['red','blue','black','orange']):
        for number in range(1,14):
            assert by_id[f'{color}-{number}']['cell'] == f'{(number-1)*192} {row*288} 192 288'
    assert by_id['duplicate-"<&']['cell'] == by_id['red-11']['cell']
    for ident,caption in [('j-blue','Joker as blue 3'),('j-black','Joker as black 12'),('j-orange','Joker as orange 12'),('rack-joker','Joker')]:
        assert by_id[ident]['caption'] == caption
        assert by_id[ident]['cell'] == '2496 576 192 288', 'Joker must retain its physical artwork'
    # Decode the same local image used by the SVG, then wait for a painted frame.
    page.evaluate('''async()=>{const im=new Image();im.src=document.querySelector('.tile-art-sprite image').getAttribute('href');await im.decode();await new Promise(requestAnimationFrame)}''')
    if args.screenshot:
        page.locator('#result>svg').screenshot(path=str(args.screenshot))
    assert 'tile-art/tiles-atlas.png' in requested
    assert not errors, errors
    browser.close()
print('Actual model art: all52 numbered sprites, 4 physical jokers with explicit roles, duplicate occurrences, exact order/accessibility and idempotent decoration passed; all assets local. Desktop only.')
