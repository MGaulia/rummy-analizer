#!/usr/bin/env python3
"""Evaluate raw local photo proposals against a consented, external dataset.

Never uploads images. Output separates real from synthetic samples. This tool
cannot certify that a manifest's claimed provenance or consent is truthful.
"""
import argparse
import hashlib
import json
from pathlib import Path


def iou(a, b):
    overlap_w = max(0, min(a['x']+a['width'], b['x']+b['width']) - max(a['x'], b['x']))
    overlap_h = max(0, min(a['y']+a['height'], b['y']+b['height']) - max(a['y'], b['y']))
    overlap = overlap_w * overlap_h
    area = a['width']*a['height'] + b['width']*b['height'] - overlap
    return overlap/area if area > 0 else 0


def match_rectangles(truth, proposed, threshold=.5):
    # Maximum-cardinality spatial matching via augmenting paths, independent of
    # recognized labels. Label accuracy cannot improve which detections match.
    owners = {}
    def assign(t, visited):
        candidates = sorted(range(len(proposed)), key=lambda p: -iou(truth[t]['rect'], proposed[p]['rect']))
        for p in candidates:
            if p in visited or iou(truth[t]['rect'], proposed[p]['rect']) < threshold:
                continue
            visited.add(p)
            if p not in owners or assign(owners[p], visited):
                owners[p] = t
                return True
        return False
    for t in range(len(truth)):
        assign(t, set())
    return [(t, p) for p, t in owners.items()]


def valid_rect(rect):
    return (set(rect) == {'x','y','width','height'} and all(isinstance(v,(int,float)) for v in rect.values())
            and rect['x'] >= 0 and rect['y'] >= 0 and rect['width'] > 0 and rect['height'] > 0
            and rect['x']+rect['width'] <= 1.00001 and rect['y']+rect['height'] <= 1.00001)


def evaluate(manifest, predictions, directory):
    if manifest.get('version') != 1 or not manifest.get('samples'):
        raise ValueError('version 1 and at least one explicitly labeled sample are required')
    summaries = {}
    seen = set()
    for sample in manifest['samples']:
        ident = sample['id']
        if ident in seen:
            raise ValueError('duplicate sample ID')
        seen.add(ident)
        if sample.get('origin') not in ('real','synthetic') or sample.get('zone') not in ('table','rack'):
            raise ValueError('each sample requires origin real/synthetic and zone table/rack')
        if sample.get('consented') is not True:
            raise ValueError('samples must be explicitly consented for local evaluation')
        image_path = directory / sample['image']
        digest = hashlib.sha256(image_path.read_bytes()).hexdigest()
        if digest != sample['sha256']:
            raise ValueError('image checksum mismatch: '+ident)
        observed = predictions[ident]
        if observed['source'] != digest or observed['zone'] != sample['zone']:
            raise ValueError('prediction source/zone mismatch: '+ident)
        truth, proposed = sample['tiles'], observed['observations']
        if not all(valid_rect(t['rect']) for t in truth+proposed):
            raise ValueError('invalid normalized rectangle: '+ident)
        pairs = match_rectangles(truth,proposed)
        result = summaries.setdefault(sample['origin'],dict(samples=0,truth_tiles=0,proposals=0,matched=0,top1_correct=0,unresolved=0,table_samples=0,table_grouping_exact=0))
        result['samples'] += 1; result['truth_tiles'] += len(truth); result['proposals'] += len(proposed); result['matched'] += len(pairs)
        for t,p in pairs:
            candidates = proposed[p].get('candidates',[])
            if not candidates: result['unresolved'] += 1
            elif candidates[0]['tile'] == truth[t]['tile']: result['top1_correct'] += 1
        if sample['zone']=='table':
            result['table_samples'] += 1
            exact = len(pairs)==len(truth)==len(proposed)
            if exact:
                exact = all('meld' in t for t in truth) and all('suggestedMeld' in p for p in proposed)
            if exact:
                exact = all((truth[t]['meld']==truth[u]['meld']) == (proposed[p]['suggestedMeld']==proposed[q]['suggestedMeld']) for t,p in pairs for u,q in pairs)
            result['table_grouping_exact'] += int(exact)
    for result in summaries.values():
        result['detection_recall'] = result['matched']/result['truth_tiles'] if result['truth_tiles'] else None
        result['detection_precision'] = result['matched']/result['proposals'] if result['proposals'] else None
        result['matched_top1_accuracy'] = result['top1_correct']/result['matched'] if result['matched'] else None
        result['end_to_end_tile_recall'] = result['top1_correct']/result['truth_tiles'] if result['truth_tiles'] else None
    return {'iou_threshold':.5,'results_by_declared_origin':summaries,'real_photo_evaluated':'real' in summaries,
            'limitations':'Manifest provenance/consent is declared, not independently certified. No device latency, offline execution, or model accuracy beyond these supplied samples is established.'}


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('manifest',type=Path);parser.add_argument('predictions',type=Path)
    args=parser.parse_args()
    try:
        print(json.dumps(evaluate(json.loads(args.manifest.read_text()),json.loads(args.predictions.read_text()),args.manifest.parent),indent=2))
    except (ValueError,KeyError,OSError,TypeError) as error:
        parser.exit(1,f'Invalid evaluation dataset: {error}\n')
