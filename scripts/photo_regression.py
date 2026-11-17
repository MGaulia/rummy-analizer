"""Hash-bound private photo regression assertions, independent of game code.

The default comparison is exact and ordered. A truth entry can select
comparison="melds" for table group-number independence or comparison="tiles" for
a rack multiset. Meld order and duplicate counts remain checked.
Nothing here reads remote URLs or stores the supplied image bytes.
"""
import base64
from collections import Counter, defaultdict
import hashlib
from html.parser import HTMLParser
import json
import math
from pathlib import Path
from urllib.parse import unquote, urlsplit


def tile_key(tile):
    return json.dumps(tile, sort_keys=True, separators=(',', ':'))


def occurrences(observations):
    return [
        {'tile': item['candidates'][0]['tile'] if item.get('candidates') else None,
         'meld': item.get('suggestedMeld'), 'order': item.get('suggestedOrder')}
        for item in observations
    ]


def meld_key(pieces):
    groups = defaultdict(list)
    for piece in pieces:
        meld, order = piece['meld'], piece['order']
        assert type(meld) is int and meld > 0, 'Invalid suggested meld number'
        assert type(order) is int and order > 0, 'Invalid suggested tile order'
        groups[meld].append((order, tile_key(piece['tile'])))
    for group in groups.values():
        assert len({order for order, _ in group}) == len(group), 'Duplicate tile order in a meld'
    # A Counter preserves identical meld multiplicities; a set would lose them.
    return Counter(tuple(sorted(group)) for group in groups.values())


def assert_reading(result, truth):
    assert result['source'] == truth['source'] and result['zone'] == truth['zone'], 'Regression source/zone mismatch'
    actual = occurrences(result['observations'])
    expected = truth['tiles']
    assert len(actual) == len(expected), ('Photo tile count changed', {'expected': len(expected), 'actual': len(actual)})
    comparison = truth.get('comparison', 'ordered')
    if comparison == 'ordered':
        matched = actual == expected
    elif comparison == 'melds':
        assert truth['zone'] == 'table', 'Meld comparison is only valid for a table photo'
        matched = meld_key(actual) == meld_key(expected)
    elif comparison == 'tiles':
        assert truth['zone'] == 'rack', 'Tile multiset comparison is only valid for a rack photo'
        matched = Counter(tile_key(p['tile']) for p in actual) == Counter(tile_key(p['tile']) for p in expected)
    else:
        raise AssertionError('Unknown photo comparison: ' + str(comparison))
    assert matched, ('Exact photo regression failed', {'comparison': comparison, 'expected': expected, 'actual': actual})


def load_report_samples(report, expected):
    assert report.get('kind') == 'rummy-photo-reading-report' and report.get('version') == 1, 'Unknown report format'
    assert set(report['photos']) == set(expected), 'Expected photo set changed'
    samples = {}
    for ident, photo in report['photos'].items():
        truth = expected[ident]
        zone = photo['zone']
        assert zone in ('table', 'rack') and zone not in samples, 'One distinct photo per zone is required'
        uri = photo['image']
        assert uri.startswith('data:image/') and ';base64,' in uri, 'Report must contain original local image bytes'
        mime, encoded = uri.split(';base64,', 1)
        data = base64.b64decode(encoded, validate=True)
        source = hashlib.sha256(data).hexdigest()
        assert source == photo['source'] == truth['source'] and truth['zone'] == zone, 'Photo source/zone mismatch'
        samples[zone] = {'source': source, 'data': data, 'mime': mime.removeprefix('data:'), 'truth': truth}
    assert set(samples) == {'table', 'rack'}, 'Both table and rack photos are required'
    assert samples['table']['source'] != samples['rack']['source'], 'Table and rack must be separate photos'
    return samples


def expected_melds(truth):
    groups = defaultdict(list)
    for piece in truth['tiles']:
        groups[piece['meld']].append(piece)
    return [[p['tile'] for p in sorted(groups[key], key=lambda p: p['order'])] for key in sorted(groups)]


def assert_expected_position(position, samples):
    table = [[piece['tile'] for piece in meld] for meld in position['table']]
    expected_table = expected_melds(samples['table']['truth'])
    if samples['table']['truth'].get('comparison', 'ordered') == 'melds':
        assert Counter(tile_key(meld) for meld in table) == Counter(tile_key(meld) for meld in expected_table), 'Position changed table melds/order/multiplicity'
    else:
        assert table == expected_table, 'Position changed table melds/order'
    rack = [piece['tile'] for piece in position['rack']]
    expected_rack = [piece['tile'] for piece in samples['rack']['truth']['tiles']]
    if samples['rack']['truth'].get('comparison', 'ordered') == 'tiles':
        assert Counter(map(tile_key, rack)) == Counter(map(tile_key, expected_rack)), 'Position changed rack multiplicity'
    else:
        assert rack == expected_rack, 'Position changed rack order/multiplicity'


def validate_corrections(observed, corrections):
    """Validate an external, hash-bound plan for only unreadable occurrences.

    Format: {"version": 1, "photos": {zone: {"source": sha256,
    "corrections": [{"index": zero_based_index, "rect": exact_raw_rect,
                     "tile": physical_tile}]}}}. Both photo zones are required.
    This returns an edit plan; it never modifies observations or app state.
    """
    assert isinstance(corrections, dict) and set(corrections) == {'version', 'photos'} and type(corrections['version']) is int and corrections['version'] == 1, 'Unknown correction format'
    assert set(observed) == set(corrections['photos']) == {'table', 'rack'}, 'Corrections require both photo zones'
    plan = []
    for zone, result in observed.items():
        photo = corrections['photos'][zone]
        assert set(photo) == {'source', 'corrections'}, 'Unknown correction photo fields'
        assert photo['source'] == result['source'] and result['zone'] == zone, 'Correction source/zone mismatch'
        assert isinstance(photo['corrections'], list), 'Corrections must be a list'
        observations = result['observations']
        assert all(isinstance(o.get('candidates'), list) for o in observations), 'Raw candidates must be explicit lists'
        unreadable = {i for i, o in enumerate(observations) if not o['candidates']}
        listed = set()
        for edit in photo['corrections']:
            assert set(edit) == {'index', 'rect', 'tile'}, 'Unknown correction fields'
            index = edit['index']
            assert type(index) is int and 0 <= index < len(observations), 'Invalid correction index'
            assert index not in listed, 'Duplicate correction index'
            listed.add(index)
            observation = observations[index]
            assert observation['candidates'] == [], 'A correction cannot override an accepted raw reading'
            rect = edit['rect']
            assert isinstance(rect, dict) and set(rect) == {'x', 'y', 'width', 'height'} and all(type(v) in (int, float) and math.isfinite(v) for v in rect.values()), 'Invalid correction rectangle'
            assert edit['rect'] == observation.get('rect'), 'Correction rectangle mismatch'
            tile = edit['tile']
            valid = tile == {'joker': True} and type(tile.get('joker')) is bool
            if isinstance(tile, dict) and set(tile) == {'color', 'number'}:
                valid = tile['color'] in ('black', 'blue', 'orange', 'red') and type(tile['number']) is int and 1 <= tile['number'] <= 13
            assert valid, 'Invalid correction tile'
            plan.append({'zone': zone, 'index': index, 'tile': tile})
        assert listed == unreadable, 'Every raw unreadable observation requires exactly one correction'
    return plan


def flatten(position):
    return [piece for meld in position['table'] for piece in meld] + position['rack']


def check_conservation(before, after, played):
    original = {piece['id']: piece for piece in flatten(before)}
    final = flatten(after)
    assert len(original) == len(flatten(before)), 'Duplicate input physical ID'
    assert len(final) == len(original) and {p['id'] for p in final} == set(original), 'A physical occurrence was lost or duplicated'
    assert all(p['tile'] == original[p['id']]['tile'] for p in final), 'Physical tile value changed'
    table_ids = {p['id'] for meld in after['table'] for p in meld}
    assert {p['id'] for meld in before['table'] for p in meld} <= table_ids, 'Existing table tile returned to rack'
    expected_played = {p['id'] for p in before['rack']} & table_ids
    assert set(played) == expected_played and len(played) == len(expected_played), 'Played IDs do not match actual rack additions'
    assert [p['id'] for p in after['rack']] == [p['id'] for p in before['rack'] if p['id'] not in expected_played]


def asset_names(root):
    """Permit only the explicit local app resources, never arbitrary report files."""
    metadata = root / 'preview-build.json'
    if metadata.is_file():
        names = set(json.loads(metadata.read_text())['files_sha256'])
    else:
        names = {'index.html', 'recognition/reader-core.js', 'recognition/reader-data.js',
                 'recognition/vision-core.js'}

        class Resources(HTMLParser):
            def handle_starttag(self, tag, attrs):
                attrs = dict(attrs)
                attribute = 'src' if tag == 'script' else 'href'
                if tag not in ('script', 'link', 'image') or attribute not in attrs:
                    return
                uri = urlsplit(attrs[attribute])
                if uri.scheme == 'data' or not uri.path:
                    return
                assert not uri.scheme and not uri.netloc, 'App resource must be local'
                names.add(unquote(uri.path))

        Resources().feed((root / 'index.html').read_text())
    assert 'index.html' in names, 'App manifest lacks index.html'
    assert all(not urlsplit(name).scheme and not urlsplit(name).netloc and not Path(name).is_absolute() and '..' not in Path(name).parts for name in names), 'Asset escaped package directory'
    return names
