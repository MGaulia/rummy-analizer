# Implementation status

## Observed starting state (2026-10-19)

- Live main: a0caeb94ae18c67fb331bfb8ff561fc516012937.
- PR #9 branch: cc8ce4f58fc9f9194b7f46ced698e5f0c1354816,
  codex/editions-jokers-table. Git refs show no newer feature branches/PR heads.
- Main baseline: `bash scripts/check.sh` passed on Go 1.27.2 linux/amd64,
  GCC 14.2.0. PR #9 baseline and fuzz validation in progress.
- Git reads work. GitHub REST/GraphQL and official booklet download currently
  return proxy 403; network additions saved to configuration draft. Open-PR
  metadata was subsequently read through the public GitHub web pages; illustrated
  rulebook review still awaits working access.

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

## Application and portable phone-engine slice

- Engine branch pushed: codex/offline-assistant-moves at
  1724da9128e19079b7be450afcec59762e5aecc6; PR opening remains blocked by API
  policy. Intended base: codex/editions-jokers-table (#9).
- Dependent branch: codex/offline-assistant-interface. Adds strict bounded JSON,
  validate/solve/check-move commands, required edition and default opening flags,
  deterministic exact SVGs, and the shared application boundary.
- Renderer data/escaping/uniqueness tests passed. The rearrangement fixture SVG
  was rendered in Chromium and visually inspected: numbered groups, explicit
  joker role, all four rack additions, and an empty remaining rack were correct.
- Actual Go/Wasm bundle ran in Node: legal joker opening, exact rendering,
  asynchronous cancellation, and independent validation of retained best result
  passed. The worker yields periodically for cancellation events. No mock engine.
- `bash scripts/check.sh` passed including tests/vet/races/build/legacy CLI.
  Structured-command executable checks were then incorporated and passed.
- Go 1.27.2 linux/amd64 remains the toolchain; this is desktop validation only.

## iPhone correction and native adapter slice

- Application/bridge branch pushed at 59c17d6e76d1743feb6c887442cd4e5daf9687b3.
  Intended PR base: codex/offline-assistant-moves. API access still blocks PRs.
- Added a functional local web interface using actual bundled Go/Wasm, separate
  table/rack sessions, raw observations/candidates, explicit per-tile value,
  meld/order correction, exclusion/addition, completeness confirmation, settings,
  cancellation, and exact rendered solutions. Browser import is manual annotation.
- Desktop offline UI integration passed with synthetic observations and the
  actual engine; no requests left the in-process asset origin. This is explicitly
  NOT recognition accuracy evidence. The 390px rendering was visually inspected;
  tile confirmation contrast and final diagram sizing were improved.
- iOS project/source adds real camera/PHPicker import, orientation normalization,
  bounded local Vision rectangles/text proposals, glyph-color suggestions, and
  a WKWebView bridge with bundled bytes. No native Swift build/device execution
  is possible here: no Swift/Xcode, macOS SDK, simulator, or attached iPhone.
- Recognition is an unmeasured candidate pipeline, including limited text-only
  joker proposals. Real consented photos/ground truth and iPhone evaluation are
  required before choosing/tuning the final vision model or claiming accuracy.
- During integration review, reproduced a release-certificate bug: reordering
  the same physical meld was incorrectly accepted as a new set. Added regression,
  observed failure against the earlier implementation, then passed after checking
  physical membership independently of ordering/roles.
- See phone/README.md for reproducible bundle, desktop checks, and unexecuted
  macOS/iPhone build/acceptance steps. The Xcode project is not a verified app.

## Validation and concrete remaining requirements

Latest combined local checks passed:
- `bash scripts/check.sh --fuzz`: tests, vet, races, build, legacy and structured
  executable checks. Inventory fuzz: 199,846 executions; parser fuzz: 228,795.
- `bash scripts/build-phone.sh`; `node scripts/check-phone-engine.cjs`;
  `node scripts/check-photo-state.cjs`; `python3 scripts/check-phone-ui.py`.
- `python3 -m unittest discover -s scripts -p test_photo_evaluation.py -v`:
  two evaluator-arithmetic/input-integrity tests passed (not recognition).
- Node v24.19.0, Python 3.12.14, Chromium 154.0.8037.57, Go 1.27.2 linux/amd64.
- Local recognition evaluation now checks source checksums and separates declared
  real/synthetic samples, but no real sample or physical iPhone was available.

Remaining requirements that this Linux workspace cannot supply:
1. Apply network settings for api.github.com and rummikub.com. Public GitHub web
   pages independently confirmed #9 is the only open PR and remains unmerged.
   API access is required to open PRs. Inspect the official booklet, finish complete joker
   operational rules/search coverage, and open the prepared dependency PRs.
2. A Mac with Xcode/iOS SDK (initial integration target iOS 17) to compile and
   run the native project. No Swift build success or simulator result is claimed.
3. Consented representative real table/rack photos with ground truth, including
   actual joker faces, for local recognition evaluation/model selection. Current
   Vision/color/grouping proposals are unmeasured and not accepted recognition.
4. An iPhone and existing signing access for physical-device camera/import,
   correction, cancellation, memory/latency, and network-disabled end-to-end
   acceptance. No minimum handset or performance claim is established.

Until those are supplied, the completed work is engine/CLI/rendering and tested
portable correction integration plus native integration source, not the complete
photo-to-move phone product. Do not relabel any of these blockers as completion.

## Final review checkpoints

All branches are pushed without modifying main or merging any GitHub PR.

| Review unit | Latest engine/application checkpoint | Intended base |
| --- | --- | --- |
| Physical moves/search | b8396fcae760836cae6385fdafc2a70b73cfd77e | #9 / codex/editions-jokers-table |
| CLI/render/Wasm | 45bc1d010b98184e16b227f9fcc681f67ca33214 | codex/offline-assistant-moves |
| Photo correction/iPhone integration candidate | 03cd6a06355d1658d1435e8265ce975c4e69daf8 | codex/offline-assistant-interface |

Source trees: engine `9533b422bf58d0154a937f73d299315b8e5e43f5`, interface
`e2a8162875d49e6ab095d388186bc4bf8604ab04`, combined
`f8335961ec67407f805c9aeb0290470192ccd10e`.

The combined full repository check, bundle build, actual Node/Wasm tests,
photo-state tests, two evaluator tests, and desktop offline UI integration ran
successfully on clean commit 33bdf0c3f6543d4b333f1d9b8773e2ca71d9ec01 (tree
bd19247e25aaf303c8d8bdc64c813df647af44c5). The dependency integration merge
fd6cbb61e35b5ac4e945c3ea69f74c1ff086eb4b has that identical tree. Engine and
interface checkpoints also passed their complete repository scripts separately.
The subsequent 03cd6a0 change only isolates manually added observations from raw
recognizer reports; photo-state and offline UI checks passed again on those
exact files before committing. Go code and bundled engine inputs did not change.

Fresh native Swift, simulator, real-photo recognition, and physical-device tests
remain UNRUN. No official-PDF visual inspection occurred because its destination
is blocked. All passed checks above are local, not hosted CI or phone results.

PR creation was attempted with the completed body and explicit base/head:
`gh pr create --repo MGaulia/rummy-analizer --head codex/offline-assistant-moves
--base codex/editions-jokers-table ... --body-file /workspace/rummy-review/moves.md`.
It failed: `Post https://api.github.com/graphql: Forbidden`. Runtime policy
still reports enforced `allowed_hosts: []`; the saved API/rulebook draft additions
have not taken effect. No new PR was created, and there is no PR URL to report.
Complete proposed descriptions are retained locally in `/workspace/rummy-review/`
(`moves.md`, `interface.md`, `iphone.md`). Open the phone integration PR as a draft
until its native build and actual acceptance checks run. No token request is
needed: injected Git access already supports the successful branch pushes.

Continue with the explicit missing requirements above. Refresh live main/PRs
before resuming, preserve these branches, and keep every PR unmerged for owner
review. Main was rechecked at a0caeb94ae18c67fb331bfb8ff561fc516012937; #9 still
points to cc8ce4f58fc9f9194b7f46ced698e5f0c1354816 and remains the only open PR.
