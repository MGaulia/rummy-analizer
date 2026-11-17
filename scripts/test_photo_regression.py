"""Check comparison/provenance semantics, not recognition accuracy."""
import base64
from copy import deepcopy
import hashlib
import json
from pathlib import Path
import tempfile
import unittest

from photo_regression import asset_names, assert_reading, load_report_samples, validate_corrections


def piece(number, meld=1, order=1, color='blue'):
    return {'tile': {'color': color, 'number': number}, 'meld': meld, 'order': order}


def reading(pieces, zone='table', source='same-source'):
    return {'source': source, 'zone': zone, 'observations': [
        {'candidates': [{'tile': p['tile']}], 'suggestedMeld': p['meld'], 'suggestedOrder': p['order']}
        for p in pieces
    ]}


def correction_fixture():
    observed, photos = {}, {}
    for zone in ('table', 'rack'):
        observed[zone] = reading([piece(1), piece(2)], zone, source=zone+'-source')
        for i, observation in enumerate(observed[zone]['observations']):
            observation['rect'] = {'x': i*.2, 'y': .1, 'width': .1, 'height': .2}
        observed[zone]['observations'][1]['candidates'] = []
        photos[zone] = {'source': zone+'-source', 'corrections': [{
            'index': 1, 'rect': dict(observed[zone]['observations'][1]['rect']),
            'tile': {'color': 'blue', 'number': 2}}]}
    return observed, {'version': 1, 'photos': photos}


class PhotoComparisonTests(unittest.TestCase):
    def test_default_keeps_exact_order_and_meld_numbers(self):
        expected = [piece(1, order=1), piece(2, order=2), piece(3, order=3)]
        truth = {'source': 'same-source', 'zone': 'table', 'tiles': expected}
        assert_reading(reading(expected), truth)
        for changed in [list(reversed(expected)), [piece(p['tile']['number'], 9, p['order']) for p in expected]]:
            with self.assertRaises(AssertionError):
                assert_reading(reading(changed), truth)

    def test_melds_ignores_only_global_group_numbering(self):
        expected = [piece(n, meld, n) for meld in (1, 2) for n in (1, 2, 3)]
        actual = [piece(n, meld, n) for meld in (80, 40) for n in (1, 2, 3)]
        truth = {'source': 'same-source', 'zone': 'table', 'comparison': 'melds', 'tiles': expected}
        assert_reading(reading(actual), truth)
        # Keep the same tile multiset but change displayed order or boundaries.
        for changed in [
            [piece(n, meld, 4-n) for meld in (80, 40) for n in (1, 2, 3)],
            [piece(n, 80, i+1) for i, n in enumerate((1, 2, 3, 1, 2, 3))],
            actual[:3],
            [dict(p, order=p['order']+1) for p in actual],
        ]:
            with self.assertRaises(AssertionError):
                assert_reading(reading(changed), truth)

    def test_melds_preserves_identical_group_multiplicity(self):
        expected = [piece(n, meld, i+1) for meld, values in [(1, (1, 2, 3)), (2, (1, 2, 3)), (3, (4, 5, 6))] for i, n in enumerate(values)]
        actual = [piece(n, meld, i+1) for meld, values in [(10, (1, 2, 3)), (20, (4, 5, 6)), (30, (4, 5, 6))] for i, n in enumerate(values)]
        with self.assertRaises(AssertionError):
            assert_reading(reading(actual), {'source': 'same-source', 'zone': 'table', 'comparison': 'melds', 'tiles': expected})

    def test_rack_multiset_preserves_duplicate_counts(self):
        expected = [piece(1), piece(1), piece(2)]
        truth = {'source': 'same-source', 'zone': 'rack', 'comparison': 'tiles', 'tiles': expected}
        assert_reading(reading([piece(2, 99, 8), piece(1, 30, 5), piece(1, 30, 6)], zone='rack'), truth)
        for changed in [[piece(1), piece(2), piece(2)], [piece(1), piece(2)]]:
            with self.assertRaises(AssertionError):
                assert_reading(reading(changed, zone='rack'), truth)

    def test_modes_do_not_disable_source_or_zone_checks(self):
        for zone, comparison in [('table', 'melds'), ('rack', 'tiles')]:
            truth = {'source': 'same-source', 'zone': zone, 'comparison': comparison, 'tiles': [piece(1)]}
            with self.assertRaises(AssertionError):
                assert_reading(reading([piece(1)], zone=zone, source='different'), truth)
            with self.assertRaises(AssertionError):
                assert_reading(reading([piece(1)], zone='rack' if zone == 'table' else 'table'), truth)
        for zone, comparison in [('table', 'tiles'), ('rack', 'melds'), ('table', 'unknown')]:
            with self.assertRaises(AssertionError):
                assert_reading(reading([piece(1)], zone=zone), {'source': 'same-source', 'zone': zone, 'comparison': comparison, 'tiles': [piece(1)]})

    def test_duplicate_order_and_unreadable_tile_are_rejected(self):
        truth = {'source': 'same-source', 'zone': 'table', 'comparison': 'melds', 'tiles': [piece(1, order=1), piece(2, order=2)]}
        with self.assertRaises(AssertionError):
            assert_reading(reading([piece(1), piece(2)]), truth)
        result = reading(truth['tiles'])
        result['observations'][1]['candidates'] = []
        with self.assertRaises(AssertionError):
            assert_reading(result, truth)

    def test_report_loading_requires_local_hash_bound_distinct_inputs(self):
        photos, truth = {}, {}
        for zone in ('table', 'rack'):
            data = (zone + ' synthetic provenance fixture, not a photo').encode()
            source = hashlib.sha256(data).hexdigest()
            photos[zone] = {'source': source, 'zone': zone, 'image': 'data:image/png;base64,'+base64.b64encode(data).decode()}
            truth[zone] = {'source': source, 'zone': zone, 'tiles': []}
        report = {'kind': 'rummy-photo-reading-report', 'version': 1, 'photos': photos}
        samples = load_report_samples(report, truth)
        self.assertEqual(set(samples), {'table', 'rack'})
        for uri in ['https://example.invalid/private-photo.png', 'file:///private-photo.png']:
            changed = deepcopy(report)
            changed['photos']['rack']['image'] = uri
            with self.assertRaises(AssertionError):
                load_report_samples(changed, truth)
        changed = deepcopy(report)
        changed['photos']['rack']['source'] = 'incorrect'
        with self.assertRaises(AssertionError):
            load_report_samples(changed, truth)
        changed = deepcopy(report)
        changed['photos']['rack'] = dict(changed['photos']['table'], zone='rack')
        same_source_truth = deepcopy(truth)
        same_source_truth['rack']['source'] = same_source_truth['table']['source']
        with self.assertRaises(AssertionError):
            load_report_samples(changed, same_source_truth)

    def test_corrections_validate_a_plan_without_changing_raw_readings(self):
        observed, corrections = correction_fixture()
        before = deepcopy(observed)
        plan = validate_corrections(observed, corrections)
        self.assertEqual(plan, [{'zone': zone, 'index': 1, 'tile': {'color': 'blue', 'number': 2}} for zone in ('table', 'rack')])
        self.assertEqual(observed, before)

    def test_corrections_cannot_override_accepted_values_or_other_sources(self):
        observed, corrections = correction_fixture()
        changed = deepcopy(corrections)
        changed['photos']['table']['corrections'][0].update(index=0, rect=observed['table']['observations'][0]['rect'])
        with self.assertRaisesRegex(AssertionError, 'accepted raw reading'):
            validate_corrections(observed, changed)
        changed = deepcopy(corrections)
        changed['photos']['rack']['source'] = 'different-photo'
        with self.assertRaisesRegex(AssertionError, 'source/zone'):
            validate_corrections(observed, changed)

    def test_corrections_bind_every_unknown_to_one_exact_rectangle(self):
        observed, corrections = correction_fixture()
        changed = deepcopy(corrections)
        changed['photos']['table']['corrections'] = []
        with self.assertRaisesRegex(AssertionError, 'Every raw unreadable'):
            validate_corrections(observed, changed)
        changed = deepcopy(corrections)
        changed['photos']['table']['corrections'] *= 2
        with self.assertRaisesRegex(AssertionError, 'Duplicate'):
            validate_corrections(observed, changed)
        changed = deepcopy(corrections)
        changed['photos']['table']['corrections'][0]['rect']['x'] += .001
        with self.assertRaisesRegex(AssertionError, 'rectangle'):
            validate_corrections(observed, changed)
        for invalid in ({'color': 'blue', 'number': 0}, {'color': 'blue', 'number': True}, {'color': 'purple', 'number': 2}, {'joker': 1}):
            changed = deepcopy(corrections)
            changed['photos']['table']['corrections'][0]['tile'] = invalid
            with self.assertRaisesRegex(AssertionError, 'Invalid correction tile'):
                validate_corrections(observed, changed)


    def test_asset_allowlist_accepts_only_local_resources(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            index = root / 'index.html'
            index.write_text('<script src="tile-geometry.js?v=source"></script><a href="https://example.invalid/license">License</a>')
            self.assertIn('tile-geometry.js', asset_names(root))
            self.assertIn('recognition/vision-core.js', asset_names(root))
            index.write_text('<script src="https://example.invalid/remote-reader.js"></script>')
            with self.assertRaises(AssertionError):
                asset_names(root)
            metadata = root / 'preview-build.json'
            metadata.write_text(json.dumps({'files_sha256': {'index.html': 'unused', '../private-photo.png': 'unused'}}))
            with self.assertRaises(AssertionError):
                asset_names(root)
            metadata.write_text(json.dumps({'files_sha256': {'index.html': 'unused', 'https://example.invalid/reader.js': 'unused'}}))
            with self.assertRaises(AssertionError):
                asset_names(root)


if __name__ == '__main__':
    unittest.main()
