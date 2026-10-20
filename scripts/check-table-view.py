"""Exercise responsive model-art boards with actual Go/Wasm output.

Synthetic positions check rendering and layout; this is not photo recognition
evaluation or physical-iPhone acceptance.
"""
import argparse
import re
from pathlib import Path

from playwright.sync_api import sync_playwright

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--preview-directory', type=Path)
parser.add_argument('--screenshot', type=Path)
args = parser.parse_args()
root = args.preview_directory.resolve() if args.preview_directory else Path(__file__).resolve().parents[1] / 'phone/web'

def numbered(ident, color, number):
    return {'id':ident, 'tile':{'color':color, 'number':number}}

def joker(ident, color=None, number=None):
    piece = {'id':ident, 'tile':{'joker':True}}
    if color:
        piece['as'] = {'color':color, 'number':number}
    return piece

position = {'table':[
    [numbered('b1','blue',1), numbered('b2','blue',2), joker('j-blue','blue',3)],
    [numbered('r12','red',12), numbered('b12','blue',12), joker('j-black','black',12)],
    [numbered('o11','orange',11), joker('j-orange','orange',12), numbered('o13','orange',13)]
], 'rack':[numbered(f'{c}-{n}',c,n) for c in ['red','blue','black','orange'] for n in range(1,14)] +
    [numbered('duplicate-"<&','red',11),joker('rack-joker')]}
expected_ids = [p['id'] for meld in position['table'] for p in meld] + [p['id'] for p in position['rack']]
requested, errors = [], []

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
        route.fulfill(body=target.read_bytes(),content_type=kind.get(target.suffix,'text/plain'))
    context.route('**/*',serve)
    page = context.new_page()
    page.on('pageerror',lambda error:errors.append(str(error)))
    page.on('console',lambda message:errors.append(message.text) if message.type=='error' else None)
    page.goto(base)
    if not page.evaluate('Boolean(globalThis.RummyTableView)'):
        page.add_script_tag(url=base+'table-view.js')
    render = '''async position=>{
      await engineReady;
      const response=JSON.parse(await rummyExecute('responsive-validate',JSON.stringify({operation:'validate',settings:{edition:6},position})));
      if(response.error)throw new Error(JSON.stringify(response));
      const resolved={table:response.position.resolved_table||[],rack:position.rack};
      document.querySelector('#result').replaceChildren(RummyTableView.render(resolved,[],{label:'Table'}));
      return [...document.querySelectorAll('#result [data-piece-id]')].map(g=>({id:g.dataset.pieceId,played:g.dataset.played,
        label:g.getAttribute('aria-label'),sprites:g.querySelectorAll('.tile-art-sprite').length,
        cell:g.querySelector('.tile-art-sprite').getAttribute('viewBox'),caption:g.closest('.art-tile').querySelector('.tile-art-caption')?.textContent||''}));
    }'''
    for width in [320,390,768,1200]:
        page.set_viewport_size({'width':width,'height':1000})
        actual = page.evaluate(render,position)
        assert [g['id'] for g in actual] == expected_ids
        assert all(g['played']=='false' and g['sprites']==1 and g['label'] for g in actual)
        by_id = {g['id']:g for g in actual}
        for row,color in enumerate(['red','blue','black','orange']):
            for n in range(1,14):
                assert by_id[f'{color}-{n}']['cell'] == f'{(n-1)*192} {row*288} 192 288'
        for ident,caption in [('j-blue','Joker as blue 3'),('j-black','Joker as black 12'),('j-orange','Joker as orange 12'),('rack-joker','Joker')]:
            assert by_id[ident]['caption']==caption
            assert by_id[ident]['cell']=='2496 576 192 288'
        assert by_id['duplicate-"<&']['cell']==by_id['red-11']['cell']
        assert page.locator('.board-rack .board-count').text_content()=='54 tiles'
        assert page.locator('.board-table .meld-card').count()==3
        assert page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'), 'Board caused page-wide overflow'

    # A long run stays in one labeled card and wraps in order. It must never
    # hide tiles behind a per-meld scrollbar or force the phone page sideways.
    long_position = {'table':position['table']+[[numbered(f'long-red-{n}','red',n) for n in range(1,14)]], 'rack':position['rack']}
    for width in [320,390,768,1200]:
        page.set_viewport_size({'width':width,'height':1000})
        page.evaluate(render,long_position)
        actual_ids = page.locator('.meld-card').last.locator('[data-piece-id]').evaluate_all('(pieces)=>pieces.map(p=>p.dataset.pieceId)')
        assert actual_ids == [f'long-red-{n}' for n in range(1,14)], 'Wrapping must preserve all thirteen tiles in order'
        assert page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'), 'Long meld caused page-wide overflow'
        rows = page.locator('.meld-row').evaluate_all('''(rows)=>rows.map(row=>({
          scroll:row.scrollWidth,client:row.clientWidth,
          pieces:[...row.querySelectorAll('.art-tile')].map(tile=>{const r=tile.getBoundingClientRect(),p=row.getBoundingClientRect();return {left:r.left-p.left,right:r.right-p.left}})
        }))''')
        assert all(r['scroll'] <= r['client']+1 for r in rows), ('Meld tiles must fit without an internal scrollbar',width,rows)
        assert all(p['left']>=-1 and p['right']<=r['client']+1 for r in rows for p in r['pieces']), ('A tile is clipped at its meld boundary',width,rows)
        if args.screenshot and width in (320,390):
            page.locator('#result').screenshot(path=str(args.screenshot.with_name(args.screenshot.stem+f'-long-{width}'+args.screenshot.suffix)))

    # A four-tile group must fit as a compact row, including the screenshot's
    # formerly clipped fourth tile. Do not spread a short meld across its card.
    four = {'table':[[numbered('four-'+c,c,11) for c in ['red','blue','black','orange']]],'rack':[]}
    for width in [320,390,768,1200]:
        page.set_viewport_size({'width':width,'height':1000})
        actual = page.evaluate(render,four)
        assert [g['id'] for g in actual] == ['four-'+c for c in ['red','blue','black','orange']]
        geometry = page.locator('.meld-row').evaluate('''row=>({scroll:row.scrollWidth,client:row.clientWidth,
          tiles:[...row.querySelectorAll('.art-tile')].map(t=>{const r=t.getBoundingClientRect();return {x:r.x,y:r.y,width:r.width}})})''')
        assert geometry['scroll'] <= geometry['client']+1, ('Four-tile meld must not scroll',width,geometry)
        tiles = geometry['tiles']
        assert all(abs(t['y']-tiles[0]['y'])<1 for t in tiles), ('Four tiles should fit on one row',width,geometry)
        assert all(tiles[i]['width']-.5 <= tiles[i+1]['x']-tiles[i]['x'] <= 56 for i in range(3)), ('Tiles overlap or their spacing is too wide',width,geometry)
        if args.screenshot and width in (390,1200):
            page.locator('#result').screenshot(path=str(args.screenshot.with_name(args.screenshot.stem+f'-four-{width}'+args.screenshot.suffix)))

    # An actual opening solve supplies played IDs; independently replay the move
    # before checking their visible played outlines and the empty remaining rack.
    opening = {'table':[], 'rack':[numbered('played-red','red',10),numbered('played-blue','blue',10),joker('played-joker')]}
    result = page.evaluate('''async position=>{
      const settings={edition:4};
      const solved=JSON.parse(await rummyExecute('responsive-solve',JSON.stringify({operation:'solve',settings,position})));
      const checked=JSON.parse(await rummyExecute('responsive-check',JSON.stringify({operation:'check-move',settings,position,move:solved.search.move})));
      if(!checked.move?.valid)throw new Error('Independent move validation failed');
      document.querySelector('#result').replaceChildren(RummyTableView.render(checked.move.after,checked.move.played));
      return {status:solved.search.status,played:checked.move.played};
    }''',opening)
    assert result['status']=='optimal' and len(result['played'])==3
    assert page.locator('[data-played="true"]').count()==3
    assert page.locator('.tile-art-played').count()==3
    marker = page.locator('.tile-art-played').first.evaluate('''node=>({stroke:getComputedStyle(node).stroke,
      background:getComputedStyle(node.closest('.meld-card')).backgroundColor})''')
    def rgb(value):
        channels = [float(n) for n in re.findall(r'[\d.]+',value)]
        assert len(channels) in (3,4) and (len(channels)==3 or channels[3]==1), ('Played marker needs a solid contrasting card',marker)
        return channels[:3]
    def luminance(channels):
        c=[v/255 for v in channels]
        c=[v/12.92 if v<=.04045 else ((v+.055)/1.055)**2.4 for v in c]
        return sum(v*w for v,w in zip(c,[.2126,.7152,.0722]))
    light,dark=sorted([luminance(rgb(marker['stroke'])),luminance(rgb(marker['background']))],reverse=True)
    assert (light+.05)/(dark+.05)>=3, ('Played marker must contrast with the meld surface',marker)
    assert page.locator('.board-rack .board-count').text_content()=='0 tiles'
    assert page.locator('.board-rack .board-empty').text_content()=='Rack is empty'
    page.evaluate('''async()=>{const im=new Image();im.src=document.querySelector('#result .tile-art-sprite image').getAttribute('href');await im.decode();await new Promise(requestAnimationFrame)}''')
    if args.screenshot:
        page.locator('#result').screenshot(path=str(args.screenshot))
    page.evaluate('''()=>document.querySelector('#result').replaceChildren(RummyTableView.render({table:[],rack:[]},[],{label:'Current table',showRack:false}))''')
    assert page.locator('.board-rack').count()==0
    assert page.locator('.board-table .board-empty').text_content()=='Table is empty'
    assert 'tile-art/tiles-atlas.png' in requested and not errors, errors
    browser.close()
print('Responsive table: actual Go validation, all 52 numbered faces, 63 occurrences, exact joker roles/rack counts, compact four-tile and wrapped 13-tile melds without scrollbars at 320/390/768/1200px, and independently checked opening/played outlines passed. Desktop only; all assets local.')
