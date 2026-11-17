"""Bundle pinned local vision/OCR runtimes and model. No inference downloads or uploads.

The Apache-2.0 runtime receives explicit bundled WebAssembly bytes. The English model is
from the pinned upstream tessdata_fast revision. No npm install is required.
"""
import base64
import hashlib
import io
import json
import os
import re
from pathlib import Path
import tarfile
import urllib.request

root = Path(__file__).resolve().parents[1]
cache = Path(os.environ.get('RUMMY_READER_CACHE', root.parent / '.rummy-cache/browser-reader'))
cache.mkdir(parents=True, exist_ok=True)
target = root / 'phone/web/recognition'
target.mkdir(parents=True, exist_ok=True)


def asset(name, url, expected, algorithm='sha256'):
    path = cache / name
    data = path.read_bytes() if path.exists() else urllib.request.urlopen(url, timeout=60).read()
    digest = hashlib.new(algorithm, data).digest()
    actual = base64.b64encode(digest).decode() if algorithm == 'sha512' else digest.hex()
    if actual != expected:
        raise SystemExit('Reader asset checksum mismatch: ' + name)
    if not path.exists():
        path.write_bytes(data)
    return data


core_url = 'https://registry.npmjs.org/tesseract.js-core/-/tesseract.js-core-6.1.2.tgz'
core = asset('tesseract.js-core-6.1.2.tgz', core_url,
             'pv4GjmramjdObhDyR1q85Td8X60Puu/lGQn7Kw2id05LLgHhAcWgnz6xSdMCSxBMWjQDmMyDXPTC2aqADdpiow==', 'sha512')
model_base = 'https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/87416418657359cb625c412a48b6e1d6d41c29bd/'
model = asset('eng.traineddata', model_base + 'eng.traineddata',
              '7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2')
license_text = asset('TESSDATA-LICENSE', model_base + 'LICENSE',
                     'cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30').decode()
with tarfile.open(fileobj=io.BytesIO(core)) as archive:
    (target / 'reader-core.js').write_bytes(archive.extractfile('package/tesseract-core-lstm.js').read())
    wasm = archive.extractfile('package/tesseract-core-lstm.wasm').read()
    core_license = archive.extractfile('package/LICENSE').read().decode()
# The pinned vision runtime contains its Wasm bytes. Replace its generated
# dynamic wrappers with ordinary closures and bypass fetch for the embedded
# data URI; the app retains its no-connect/no-unsafe-eval CSP.
vision_url = 'https://registry.npmjs.org/@techstark/opencv-js/-/opencv-js-4.12.0-release.1.tgz'
vision = asset('opencv-js-4.12.0-release.1.tgz', vision_url,
               'e5745fae26adb6ba79e90ef3bf3b639b5b696187f7230c1b346785bc4a759a85')
with tarfile.open(fileobj=io.BytesIO(vision)) as archive:
    vision_js = archive.extractfile('package/dist/opencv.js').read().decode()
    vision_license = archive.extractfile('package/LICENSE').read().decode()
profile = (root / 'scripts/vision-csp.js').read_text()
for name, next_function in [
    ('createNamedFunction', 'extendError'),
    ('makeDynCaller', None),
    ('craftInvokerFunction', 'heap32VectorToArray'),
    ('__emval_get_method_caller', '__emval_get_property'),
]:
    replacement = re.search(r'^function ' + name + r'\([^\n]*[\s\S]*?(?=^function |\Z)', profile, re.MULTILINE)
    if replacement is None:
        raise SystemExit('Missing CSP binding replacement: ' + name)
    start = vision_js.index('function ' + name + '(')
    end = vision_js.index('function ' + next_function + '(', start) if next_function else vision_js.index('var fp;', start)
    vision_js = vision_js[:start] + replacement.group().rstrip() + vision_js[end:]
# Export the factory instead of creating its heap at page load. The preview
# downloads the script with its other assets; only local instantiation is lazy.
if vision_js.count('return cv(Module);') != 1:
    raise SystemExit('Pinned vision runtime factory profile changed')
vision_js = vision_js.replace('return cv(Module);', 'return cv;')
vision_js = vision_js.replace('root.cv = factory()', 'root.RummyVisionRuntime = factory()')
fetch_condition = 'if(!wasmBinary&&(ENVIRONMENT_IS_WEB||ENVIRONMENT_IS_WORKER)&&typeof fetch==="function")'
if vision_js.count(fetch_condition) != 1 or 'new Function(' in vision_js or 'new_(Function' in vision_js:
    raise SystemExit('Pinned vision runtime no longer matches the audited CSP profile')
vision_js = vision_js.replace(fetch_condition,
    'if(!isDataURI(wasmBinaryFile)&&!wasmBinary&&(ENVIRONMENT_IS_WEB||ENVIRONMENT_IS_WORKER)&&typeof fetch==="function")')
(target / 'vision-core.js').write_text(vision_js)
notices = ('Tesseract.js-core 6.1.2 — Apache-2.0\n' + core_url + '\n\n' + core_license +
           '\n\nTesseract tessdata_fast English model — Apache-2.0\n' + model_base + '\n\n' + license_text +
           '\n\nOpenCV.js 4.12.0-release.1 — Apache-2.0\n' + vision_url +
           '\nAdapted generated wrappers and embedded-data loading for the existing CSP.\n\n' + vision_license)
(target / 'READER-LICENSES.txt').write_text(notices)
(target / 'reader-data.js').write_text(
    'globalThis.RUMMY_READER_WASM="' + base64.b64encode(wasm).decode() + '";\n' +
    'globalThis.RUMMY_READER_MODEL="' + base64.b64encode(model).decode() + '";\n' +
    'globalThis.RUMMY_READER_LICENSES=' + json.dumps(notices) + ';\n')
print('Bundled checksum-verified local vision/OCR runtimes, model, and licenses.')
