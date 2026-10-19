# Implementation status

## Observed starting state (2026-10-19)

- Live main: a0caeb94ae18c67fb331bfb8ff561fc516012937.
- PR #9 branch: cc8ce4f58fc9f9194b7f46ced698e5f0c1354816,
  codex/editions-jokers-table. Git refs show no newer feature branches/PR heads.
- Main baseline: `bash scripts/check.sh` passed on Go 1.27.2 linux/amd64,
  GCC 14.2.0. PR #9 baseline and fuzz validation in progress.
- Git reads work. GitHub REST/GraphQL and official booklet download currently
  return proxy 403; network additions saved to configuration draft. Open-PR
  metadata and illustrated rulebook review await working access.

## Remaining milestones

1. Reuse/review #9's edition inventory, joker meld/layout, and identified positions.
   Add any rulebook-backed gaps; preserve its code/tests instead of duplicating.
2. Implement move validation: physical conservation, opening/draw state, exact
   rack accounting, and explicit joker release/reuse evidence. Cover rejected
   transitions and caller immutability.
3. Implement bounded cancellable search over all meld partitions, maximizing
   played rack occurrences; independently compare tiny cases exhaustively and
   verify each result. Report honest scope/completeness for joker transitions.
4. Add explicit settings and structured CLI commands, exact SVG rendering,
   integration fixtures, and Go-core offline mobile bridge feasibility.
5. Add separate photo observation/correction boundary; real recognizer requires
   representative licensed/consented photos and evaluated local assets. Do not
   label a manual correction boundary as recognition.
6. Build a functioning phone interface and engine integration, then evaluate real
   recognition and offline physical-device workflow. No device is attached now.

## Review and dependencies

Implementation starts on codex/offline-assistant-moves based on #9. Keep #9
unmerged. Engine PR depends on #9, application/mobile work depends on engine.
Owner merges in dependency order; if squashing, rebase dependent branches onto
actual merged commits and retarget main, checking the cumulative diff. Do not
assume every merge strategy automatically cleans dependent diffs.

Track exact commits/trees, checks, outcomes, PR URLs, and blockers here as work
lands. A saved network draft is not evidence runtime policy changed. A branch
push is not an opened PR. Desktop/mobile/browser/device results stay separate.

## Engine slice delivered on the working branch

- Reused #9 unchanged. Its full baseline and 10-second inventory/parser fuzz
  targets passed on Go 1.27.2; 216,417 and 184,660 executions respectively.
- Added physical transition validation and explicit joker release certificates,
  opening/draw checks, and cancellable bounded physical-partition search.
- `bash scripts/check.sh` passed with these changes (tests, vet, races, build,
  executable checks). Independent oracle comparison: 512 rack-only subsets and
  64 positions with mandatory table tiles. Replacement/restructuring tests pass.
- `go test ./internal/game -run '^$' -bench BenchmarkSearchRearrangement -benchmem
  -benchtime=1s`: 553,655 ns/op, 816,147 B/op, 7,176 allocs/op on this Linux
  AMD EPYC host. This is a tiny desktop fixture, not a phone latency claim.
- Full joker-transition profile remains incomplete pending booklet review;
  engine documents the sufficient witness scope and does not return false proofs.
- Target platform clarified by owner: iPhone first. No Xcode/macOS/iPhone here.
