#!/usr/bin/env bash
# Build an explicit, static-only browser preview; never publish repository files.
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
if [[ $# -ne 1 ]]; then
  printf 'Usage: bash scripts/build-safari-preview.sh EMPTY_OUTPUT_DIRECTORY\n' >&2
  exit 2
fi
bash scripts/build-phone.sh
python3 - "$1" <<'PYCODE'
import hashlib
import html
import json
from pathlib import Path
import shutil
import subprocess
import sys

destination = Path(sys.argv[1]).resolve()
destination.mkdir(parents=True, exist_ok=True)
if any(destination.iterdir()):
    raise SystemExit('Preview output must be an empty directory.')
source = Path('phone/web')
commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip()
tree = subprocess.check_output(['git', 'rev-parse', 'HEAD^{tree}'], text=True).strip()
if subprocess.check_output(['git', 'status', '--porcelain', '--untracked-files=no']):
    raise SystemExit('Commit tracked source changes before building a published preview.')
assets = ['index.html', 'style.css', 'engine.js', 'photo-state.js', 'app.js',
          'engine/wasm_exec.js', 'engine/wasm_data.js', 'engine/GO-LICENSE.txt']
for name in assets:
    target = destination / name
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source / name, target)
index = destination / 'index.html'
page = index.read_text()
notice = '''<section aria-label="Safari preview instructions"><h2>Safari test preview</h2>
<p>Choose separate table and rack photos. Tap each tile in each photo, enter its
color and number (or joker), and confirm it. For table tiles, enter the meld
number and its displayed order. Confirm each complete reading, then check the
table or find a move. An empty table photo is allowed.</p>
<p>Tile entry is manual in this browser preview. Automatic recognition requires
the native iPhone app. Photos and solving stay on this device. Reloading clears
your readings; reopening offline is not supported by this preview.</p>
<p class="notice">Preview source: <code>''' + html.escape(commit[:12]) + '''</code>.
Native build and recognition acceptance remain pending.</p></section>'''
page = page.replace('</header>', '</header>\n' + notice, 1)
page = page.replace('PERSONAL · OFFLINE', 'PERSONAL · SAFARI PREVIEW')
for zone in ['table', 'rack']:
    page = page.replace('<button data-camera="' + zone + '" class="secondary">Take photo</button>', '')
page = page.replace('Recognition suggestions require your confirmation.',
                    'Browser tile entry is manual. Confirm every reading.')
index.write_text(page)
(destination / '.nojekyll').touch()
metadata = {
    'source_commit': commit,
    'source_tree': tree,
    'go_version': subprocess.check_output(['go', 'version'], text=True).strip(),
    'files_sha256': {name: hashlib.sha256((destination / name).read_bytes()).hexdigest()
                     for name in assets},
}
(destination / 'preview-build.json').write_text(json.dumps(metadata, indent=2) + '\n')
print('Safari preview built at ' + str(destination))
PYCODE
