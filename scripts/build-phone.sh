#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
export GOWORK=off
mkdir -p phone/web/engine
GOOS=js GOARCH=wasm go build -mod=readonly -trimpath -o phone/web/engine/rummy.wasm ./cmd/phonewasm
runtime="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [[ ! -f "$runtime" ]]; then runtime="$(go env GOROOT)/misc/wasm/wasm_exec.js"; fi
cp "$runtime" phone/web/engine/wasm_exec.js
python3 - <<'PYCODE'
import base64
from pathlib import Path
root = Path('phone/web/engine')
encoded = base64.b64encode((root / 'rummy.wasm').read_bytes()).decode('ascii')
(root / 'wasm_data.js').write_text('globalThis.RUMMY_WASM_BASE64="' + encoded + '";\n')
PYCODE
printf 'Bundled offline Go engine and matching runtime in phone/web/engine\n'
