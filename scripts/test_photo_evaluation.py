"""Tests of evaluation arithmetic only; no recognition measurements."""
import hashlib
import importlib.util
from pathlib import Path
import tempfile
import unittest

spec=importlib.util.spec_from_file_location('evaluate_photos',Path(__file__).with_name('evaluate-photos.py'))
evaluate_photos=importlib.util.module_from_spec(spec)
spec.loader.exec_module(evaluate_photos)

class EvaluationTests(unittest.TestCase):
    def test_augmenting_match_and_no_label_leakage(self):
        # A overlaps both proposals, B only the first: greedy one-pass loses B.
        truth=[{'rect':{'x':.2,'y':0,'width':.6,'height':1}}, {'rect':{'x':0,'y':0,'width':.6,'height':1}}]
        proposals=[{'rect':{'x':.1,'y':0,'width':.6,'height':1}}, {'rect':{'x':.4,'y':0,'width':.6,'height':1}}]
        self.assertEqual(len(evaluate_photos.match_rectangles(truth,proposals)),2)
    def test_provenance_checks_and_metrics(self):
        with tempfile.TemporaryDirectory() as temp:
            directory=Path(temp);image=directory/'local-photo';image.write_bytes(b'evaluation fixture, not an actual photo')
            source=hashlib.sha256(image.read_bytes()).hexdigest()
            rect={'x':0,'y':0,'width':.2,'height':.5};tile={'color':'blue','number':1}
            manifest={'version':1,'samples':[{'id':'a','image':image.name,'sha256':source,'zone':'rack','origin':'synthetic','consented':True,'tiles':[{'rect':rect,'tile':tile}]}]}
            predictions={'a':{'source':source,'zone':'rack','observations':[{'rect':rect,'candidates':[{'tile':tile}]}]}}
            result=evaluate_photos.evaluate(manifest,predictions,directory)
            self.assertFalse(result['real_photo_evaluated'])
            self.assertEqual(result['results_by_declared_origin']['synthetic']['end_to_end_tile_recall'],1)
            predictions['a']['observations'][0]['candidates']=[]
            result=evaluate_photos.evaluate(manifest,predictions,directory)
            self.assertEqual(result['results_by_declared_origin']['synthetic']['end_to_end_tile_recall'],0)
            manifest['samples'][0]['sha256']='wrong'
            with self.assertRaises(ValueError): evaluate_photos.evaluate(manifest,predictions,directory)

if __name__=='__main__':unittest.main()
