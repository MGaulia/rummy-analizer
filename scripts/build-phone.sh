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
printf 'Bundled offline Go engine and matching runtime in phone/web/engine\n'
