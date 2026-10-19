#!/usr/bin/env bash
set -euo pipefail

fuzz=false
if [[ $# -eq 1 && $1 == --fuzz ]]; then
  fuzz=true
elif [[ $# -ne 0 ]]; then
  printf 'Usage: bash scripts/check.sh [--fuzz]\n' >&2
  exit 2
fi

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
for tool in go gofmt git; do
  command -v "$tool" >/dev/null || { printf 'Missing tool: %s\n' "$tool" >&2; exit 1; }
done
export GOWORK=off
printf 'Commit: %s\n' "$(git rev-parse HEAD)"
go version
printf 'Working-tree status (empty means clean):\n'
git status --short

printf 'Checking formatting and whitespace...\n'
unformatted="$(gofmt -l .)"
if [[ -n "$unformatted" ]]; then
  printf 'Run gofmt on:\n%s\n' "$unformatted" >&2
  exit 1
fi
git diff --check
git diff --cached --check
git show --format= --check HEAD

printf 'Running tests, vet, and race checks...\n'
go test -mod=readonly -count=1 ./...
go vet -mod=readonly ./...
go test -mod=readonly -race -count=1 ./...

tmp="$(mktemp -d "${TMPDIR:-/tmp}/rummy-check.XXXXXX")"
trap 'rm -rf -- "$tmp"' EXIT
go build -mod=readonly -o "$tmp/rummy-analizer" .

expect_output() {
  if ! grep -Fq -- "$2" "$1"; then
    printf 'Missing expected output: %s\n' "$2" >&2
    cat "$1" >&2
    exit 1
  fi
}
expect_empty() {
  if [[ -s "$1" ]]; then
    printf 'Expected an empty stream: %s\n' "$1" >&2
    cat "$1" >&2
    exit 1
  fi
}

printf 'Checking executable behavior...\n'
"$tmp/rummy-analizer" -h >"$tmp/out" 2>"$tmp/err"
expect_output "$tmp/out" 'Usage:'
expect_empty "$tmp/err"
# Exercise recovery, piped input, and a final line without a newline.
printf 'invalid\nblue 3\nblue 1\nblue 2' | "$tmp/rummy-analizer" >"$tmp/out" 2>"$tmp/err"
expect_output "$tmp/out" '[blue 3, blue 1, blue 2]'
expect_output "$tmp/out" 'Can form group/run: true'
expect_empty "$tmp/err"
printf 'myblue 1\nmyblue 2\nmyblue 3' | "$tmp/rummy-analizer" -color-aliases myblue=blue >"$tmp/out" 2>"$tmp/err"
expect_output "$tmp/out" '[blue 1, blue 2, blue 3]'
expect_output "$tmp/out" 'Can form group/run: true'
expect_empty "$tmp/err"
printf 'blue 1\njoker\nblue 3' | "$tmp/rummy-analizer" >"$tmp/out" 2>"$tmp/err"
expect_output "$tmp/out" 'Can form group/run: true'
expect_empty "$tmp/err"
for argument in -unknown extra; do
  status=0
  "$tmp/rummy-analizer" "$argument" </dev/null >"$tmp/out" 2>"$tmp/err" || status=$?
  if [[ $status -ne 1 || ! -s "$tmp/err" ]]; then
    printf 'Argument %s: expected exit 1 and stderr, got exit %s\n' "$argument" "$status" >&2
    exit 1
  fi
  expect_empty "$tmp/out"
done

printf 'Checking structured position commands and exact rendering...\n'
"$tmp/rummy-analizer" solve --edition=4 --opened <testdata/positions/rearrangement.json >"$tmp/out" 2>"$tmp/err"
expect_output "$tmp/out" '"status": "optimal"'
expect_output "$tmp/out" 'data-piece-id='
expect_output "$tmp/out" 'Joker as'
expect_empty "$tmp/err"
"$tmp/rummy-analizer" validate --edition=6 --format=svg <testdata/positions/rearrangement.json >"$tmp/out" 2>"$tmp/err"
expect_output "$tmp/out" '<svg'
expect_output "$tmp/out" 'Remaining rack'
expect_empty "$tmp/err"

if [[ $fuzz == true ]]; then
  printf 'Running bounded inventory and parser fuzz checks...\n'
  go test -mod=readonly ./internal/game -run '^$' -fuzz '^FuzzInventoryConsumption$' -fuzztime=10s -parallel=2
  go test -mod=readonly ./internal/cli -run '^$' -fuzz '^FuzzParseTile$' -fuzztime=10s -parallel=2
fi
printf 'All requested local checks passed.\n'
