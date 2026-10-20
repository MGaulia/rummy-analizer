"""Bundle a pinned OCR runtime/model as scripts. No runtime downloads or uploads.

The Apache-2.0 runtime receives explicit bundled WebAssembly bytes. The English model is
from the pinned upstream tessdata_fast revision. No npm install is required.
"""
import base64
import hashlib
import io
import json
import os
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
notices = ('Tesseract.js-core 6.1.2 — Apache-2.0\n' + core_url + '\n\n' + core_license +
           '\n\nTesseract tessdata_fast English model — Apache-2.0\n' + model_base + '\n\n' + license_text)
(target / 'READER-LICENSES.txt').write_text(notices)
(target / 'reader-data.js').write_text(
    'globalThis.RUMMY_READER_WASM="' + base64.b64encode(wasm).decode() + '";\n' +
    'globalThis.RUMMY_READER_MODEL="' + base64.b64encode(model).decode() + '";\n' +
    'globalThis.RUMMY_READER_LICENSES=' + json.dumps(notices) + ';\n')
print('Bundled checksum-verified local OCR runtime, model, and licenses.')
