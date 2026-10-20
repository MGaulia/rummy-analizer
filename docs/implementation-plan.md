# Implementation status

## Browser reading follow-up (2026-10-19, owner local date)

- Owner reported that Safari did not read their tiles and supplied two inline
  photos. The previous published preview intentionally had no automatic browser
  recognizer. Visual transcription: rack red 11, black 9, blue/teal 2 and 3,
  orange 8; table orange/black/red 12, black 8–joker–10, blue/black/orange 11.
  These are manual observations, not program output. The inline originals are
  not mounted under `/workspace/attachments` or otherwise accessible as files;
  real-pixel evaluation on those exact examples remains blocked by original-file
  access. Do not turn this transcription into a mocked recognition result.
- Implemented a separate browser-reader branch on top of unmerged draft #12.
  Real local Tesseract.js-core 6.1.2 and a pinned tessdata_fast model are bundled
  as scripts with verified source checksums and Apache-2.0 notices. No workers,
  runtime fetches, photo uploads, cloud inference or remote solver. Inputs are
  bounded to 1400px; ivory segmentation, touching-strip seams, ink crops, OCR,
  cyan/teal color proposals, low-confidence possible-joker detection and geometric
  meld suggestions feed the existing raw/corrected confirmation boundary.
- Actual-OCR tests initially exposed missing 5/9 readings under single-line
  segmentation; a raw-line fallback fixes the synthetic regression. The new
  integration check also exposed an existing CSP violation in the embedded SVG
  style element. Browser rendering now removes the generated style before parsing
  and uses the local stylesheet; standalone Go SVG output is unchanged. The
  strict CSP is preserved. Synthetic reader, actual file import, confirmation,
  real-engine opening and no-network/no-CSP-error checks passed before committing.
- The owner also requested real official tiles for visualization. Inspected the
  official site's home page, media searches and older asset records. Found product
  photographs, joker logos and rulebook illustrations, but not a complete clean
  individual tile-face atlas. Requested the intended official page URL using an
  asynchronous question; original tile-art substitution remains pending that
  asset/source. Do not silently replace it with generated/CSS tiles. Continue
  browser reading and publication independently.
- Owner supplied the desired source:
  <https://www.printables.com/model/130082-multicolor-rummikub-tiles>. This
  supersedes the request for an unspecified official tile-art page. The actual
  request returned proxy CONNECT HTTP 403. Saved page/file/API domains
  `printables.com`, `www.printables.com`, `media.printables.com`,
  `files.printables.com`, `api.printables.com` alongside the existing three
  allowed hosts. The draft save was confirmed; it requires user save/publication.
  No token or existing binding was replaced. Blender and OpenSCAD are installed,
  so local model inspection/rendering can proceed once model assets/license are
  accessible. Do not invent a license or describe the current SVG as real art.
- [Draft #13](https://github.com/MGaulia/rummy-analizer/pull/13) is the focused
  reader change, based on `codex/iphone-photo-correction` (#12). Merge order is
  #12 then #13; both stay unmerged. Exact tested/published source:
  `d4aa14e9d1fa2ca8b3bca291423ba7c293f6a82a`; source tree
  `6c86700bcd265070cd145f7788d960a84f2551d3`. Go 1.27.2 linux/amd64,
  Node v24.19.0, Python 3.12.14, desktop Chromium 154.0.8037.57. With the existing
  Go PATH/cache activation, all of these commands passed:

  ```sh
  bash scripts/check.sh
  bash scripts/build-safari-preview.sh /workspace/rummy-review/safari-preview-local-reader
  node scripts/check-phone-engine.cjs
  node scripts/check-photo-state.cjs
  python3 scripts/check-phone-ui.py
  python3 scripts/check-safari-preview.py /workspace/rummy-review/safari-preview-local-reader
  python3 scripts/check-browser-reader.py
  python3 scripts/check-browser-reader.py --preview-directory /workspace/rummy-review/safari-preview-local-reader
  ```

- The final command also checks the actual published package, including its
  script-loaded OCR/runtime/model bytes, with no mocked recognizer. Its generated
  images remain explicitly synthetic. Original photo bytes, real-photo accuracy,
  physical-iPhone execution, memory and latency remain unverified. Recognition
  proposals never count as confirmed physical tile identities without review.
- All 14 remote static files exactly match the checked package. Pages commit
  `75c5e96f0190da8d49e58d5dd8752d98ed02f831` fast-forwards the previous
  `eb2e7e041e64ab18df5ec187a22dcff67cd1a298`. GitHub's API reports that exact
  commit `built` with no error. Preview:
  <https://mgaulia.github.io/rummy-analizer/?v=d4aa14e9d1fa>.
  Direct public-URL access from this workspace remains blocked by its pending
  network policy; no live Safari or original-photo success is inferred from the
  deployment result. Subsequent changes here only add the package-check option
  and update documentation; the published web/runtime source remains the exact
  checked source above.

## Current state after network publication (2026-10-19)

- Published policy is enforced for api.github.com and rummikub.com; actual API
  reads/writes and the official PDF download now work. No network action remains.
- At the owner's request, #9, #10, and #11 were merged in order using merge
  commits. Main is `23eee9e492acd0cbcc97901143d40ab5a1bdee3d`, with tested tree
  `6ef761a1976be9a65a80e039beff184573019d16`. Fresh repository checks with fuzz,
  Wasm build, and Node engine checks passed before merging; the final main tree
  exactly matched that tested tree. All three PR heads are ancestors of main.
  [Draft #12](https://github.com/MGaulia/rummy-analizer/pull/12) now targets main
  and remains unmerged. The owner authorized publishing its Safari preview on
  GitHub Pages independently of merging or native acceptance. The preview is
  generated by `scripts/build-safari-preview.sh`; see `phone/README.md` for
  manual tile entry and physical-phone test instructions.
- Official PDF D-2600-1236-0041 (120819), pages 3-4, visually inspected. All four
  release illustrations (including all three group-replacement alternatives)
  are encoded for both editions. The opening policy matches page 1. Source hash
  and exact semantics are in docs/joker-transitions.md.
- Engine: initial joker ambiguity, retained-joker replacement, chained release/
  reuse, bounded checkpoint graph, and rigorous relaxed-partition proof bounds.
  The zero bound proves absence; a verified move matching the bound proves the
  maximum rack-tile count. Unresolved temporary-loose-tile transitions remain
  explicitly incomplete; the valid-table checkpoint format is sufficient, not
  proven complete for every possible manipulation. No false no-move claim.
- 576 numbered oracle cases plus 31 nonempty one-table-joker rack subsets using
  independent wildcard substitution/partitioning. Four-joker expanded-edition,
  table-only joker reuse, invalid certificates, identity, and cancellation tests
  also pass. See the commands and exact tested checkpoints below.
- Native iPhone source, real photo accuracy, and device acceptance remain UNRUN.
  Owner confirmed **no Mac build machine** and **no real dataset** are available.
  The owner now has an iPhone available for the manual Safari preview test.
  Those prerequisites block compiling/validating the native integration and
  selecting/evaluating the recognizer. #12 remains a draft.
- Local bundle now ships Go's BSD license text from the matching toolchain and
  displays it offline. Phone status text explains limits without internal codes.

## Safari preview publication checkpoint (2026-10-19)

- Owner authorized a Pages preview while keeping #12 unmerged. Tested source:
  `7f9d357f974ec18d40d0357f8ebc8d3c501a212c`; clean source tree:
  `f23a8aa23b62e12723de6772acab5a29d5701c5a`.
- Toolchain: Go 1.27.2 linux/amd64, Node v24.19.0, Python 3.12.14,
  desktop Chromium 154.0.8037.57. Commands run from the repository root with
  `/workspace/.rummy-tools/go/bin` first in PATH and the existing workspace
  GOCACHE/GOPATH (`GOWORK=off`):

  ```sh
  bash scripts/check.sh
  bash scripts/build-safari-preview.sh /workspace/rummy-review/safari-preview-publish
  node scripts/check-phone-engine.cjs
  node scripts/check-photo-state.cjs
  python3 scripts/check-phone-ui.py
  python3 scripts/check-safari-preview.py /workspace/rummy-review/safari-preview-publish
  ```

- All commands passed. Repository tests/vet/races/build/CLI smoke passed. The
  final package check imported two distinct synthetic PNG files through the
  file inputs, manually entered a red/blue/black 10 rack, confirmed an empty
  table, and checked the actual Go/Wasm 30-point opening and exact three-tile
  rendering at a nested Pages path. No upload or remote solver requests occurred.
  These are desktop Chromium checks, not Safari or recognition accuracy evidence.
- Published only ten static package files to `gh-pages` at
  `383f66c17baad35a9d31bd4c8a0225f706fa882d`. No source, native project, photos,
  credentials, or CI workflow is included. `preview-build.json` records the
  exact source commit/tree, Go version, and individual asset SHA256 hashes.
- Blocker: `POST /repos/MGaulia/rummy-analizer/pages` with `build_type=legacy`
  and source `gh-pages:/` returned HTTP 403, `Resource not accessible by
  integration`. The integration can push code but cannot enable Pages. Owner
  must select **Settings → Pages → Deploy from a branch → gh-pages → / (root)**
  and save. Expected URL: `https://mgaulia.github.io/rummy-analizer/`; deployment
  and public HTTP/Safari execution are not yet confirmed. No live-site claim.
- Preview instructions explicitly say manual tile entry, local photos/solver,
  reload loses readings, and no supported offline reopening. Native camera
  buttons are omitted. #12 remains a draft; native/recognition blockers remain.

Deployment follow-up (owner local date 2026-10-19): GitHub's Pages API now reports
`status=built`, source `gh-pages:/`, public URL
`https://mgaulia.github.io/rummy-analizer/`, and HTTPS enforced. The latest build
completed without an error and its commit is exactly the published
`383f66c17baad35a9d31bd4c8a0225f706fa882d`; the source metadata still matches
`7f9d357f974ec18d40d0357f8ebc8d3c501a212c`. Thus Pages activation is resolved.
After the owner reported a generic environment save failure, configuration reads
confirmed a saved `RUMMY_PAGES_TOKEN` binding and the intended three hostnames.
The token variable is not yet present in this workspace, and runtime policy/
secret observations are unknown. The direct public-URL request returned proxy
CONNECT HTTP 403, so no live HTTP/browser success is inferred from the API build.
A minimal network-only draft retry was saved successfully, preserving the secret
binding; publication remains required for workspace configuration. No duplicate
credential was requested and no secret value was read or printed. The owner can
test the Pages link in Safari independently of this workspace update.

## Rummikub-style UI republication (2026-10-19, owner local date)

- Implemented the requested game styling: green felt, ivory tiles in the four
  colors, wooden rack, numbered step tiles, touch-friendly controls, gold move
  button, and responsive table/rack photo panels. Shortened browser-preview
  guidance into a visible manual-entry notice and expandable instructions.
  No engine, photo processing, rule, or search semantics changed. No external
  fonts, images, dependencies, or runtime network calls were added.
- Exact tested/published source: `68ad21da591290b4a2d5d3640e7bbeb8317b431e`;
  tree `6138ccdaa3ec061c1ddd292bdbcd1750037a61a2`. Go 1.27.2 linux/amd64,
  Node v24.19.0, Python 3.12.14, desktop Chromium 154.0.8037.57. With the existing
  Go PATH/cache activation, these all passed:

  ```sh
  bash scripts/check.sh
  bash scripts/build-safari-preview.sh /workspace/rummy-review/safari-preview-game-publish
  node scripts/check-phone-engine.cjs
  node scripts/check-photo-state.cjs
  python3 scripts/check-phone-ui.py
  python3 scripts/check-safari-preview.py /workspace/rummy-review/safari-preview-game-publish
  ```

- Inspected desktop/mobile screenshots and checked empty-page layout at 320,
  390, 768, and 1280 pixels: no page-width overflow. Fixed a missing arrow glyph
  using CSS and added source-version query parameters to CSS/JS assets for Safari
  cache refresh. The final generated-package test exercised these versioned URLs,
  manual file import/correction, and actual solve/render with no uploads.
- Fast-forward publication preserved Pages history. Static deployment commit:
  `eb2e7e041e64ab18df5ec187a22dcff67cd1a298`, parent
  `383f66c17baad35a9d31bd4c8a0225f706fa882d`. GitHub Pages reports this exact
  commit as `built` without an error; its deployment run
  <https://github.com/MGaulia/rummy-analizer/actions/runs/39345720538> completed
  successfully. HTTPS remains enforced. No CI workflow was added to the source.
- Public URL: <https://mgaulia.github.io/rummy-analizer/?v=68ad21da5912>.
  Workspace CONNECT requests to the public host still return proxy HTTP 403,
  so verification is via the deployment API plus local browser package checks.
  Actual iPhone/Safari execution and recognition/native acceptance remain unrun.
  #12 remains a draft against main and is not merged.

The remaining acceptance work is not a completed photo-to-move phone product.
Before continuing, obtain a Mac/Xcode/iPhone with authorized build/signing access
and consented representative table/rack photos with ground truth. Run the native
build, evaluate local recognition, then perform the full network-disabled device
workflow. Preserve explicit joker search limitations while those are unresolved.

## Earlier checkpoints and historical failures

The sections below preserve earlier observations. Former API/PDF denials and
"no PR created" entries are historical; the current PR/access state is above.

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

## Verification after published access and rulebook review

Source/branch checkpoints (Go 1.27.2 linux/amd64):

| Unit | Commit | Source tree |
| --- | --- | --- |
| Rulebook-backed engine (#10) | d2ebc415831a0ed5a7cee6707d5985d83b1c48d0 | 88349feb862fd75e33952b1a6be13326a52d147d |
| Integrated CLI/Wasm (#11) | 39d440d8d47f9798170e6b8e96300496a600cb78 | 6ef761a1976be9a65a80e039beff184573019d16 |
| Integrated phone code (#12) | f5b69a5267a3dd52961d4101dd6e16603363a002 | 7cf2a46289ccb6a9c120234fb4d05d0c0443e680 |
| Offline notices/status UI (#12) | 546e7764286ac22a917150c5f143e43d762584e4 | 818489d1fb78dfd992ebc14501001cf1cee967eb |

Actual checks:
- `bash scripts/check.sh` passed on the combined clean f5b69a5 commit: uncached
  tests, formatting/whitespace, vet, races, executable build, and original/new
  CLI behavior. Engine/interface scripts also passed separately on the source
  incorporated in their listed commits. Existing generated bundle/bytecode
  artifacts appeared untracked on the older branches and were not committed.
- `bash scripts/build-phone.sh` and `node scripts/check-phone-engine.cjs` passed:
  actual Go/Wasm solve, SVG, cancellation/retained-best replay, and tight table-
  joker bound through JSON. The later bundle includes/tests the Go BSD notice.
- `node scripts/check-photo-state.cjs`, two evaluator arithmetic/integrity tests
  (`python3 -m unittest discover -s scripts -p test_photo_evaluation.py -v`), and
  `python3 scripts/check-phone-ui.py` passed again on 546e776. That UI check uses
  synthetic observations, actual Go/Wasm, no external requests, and checks raw
  report isolation, confirmation/invalidation, rendering, plain status and notice.
- Existing parser/inventory fuzz runs are recorded above; these components were
  unchanged in the joker follow-up. No new fuzz or native/device result is claimed.
- The unbound-role and retained-second-joker regressions were observed failing
  before their fixes. The chained-reuse regression failed before `reused` stages
  existed; the independent witness finder now returns a verified staged path.
  The relaxed-bound examples failed under the former blanket-incomplete policy
  and now prove the requested optimum/absence. Four-joker search initially hit
  its work limit; rejecting impossible template lengths and skipping non-improving
  witness work lets that regression complete under the same bound.
- `go test ./internal/game -run '^$' -bench BenchmarkSearchRearrangement -benchmem
  -benchtime=1s`: 93,552 ns/op, 141,833 B/op, 2,094 allocs/op on AMD EPYC 9V74.
  This host differs from the earlier AMD EPYC 7763; this is not a controlled
  speedup comparison or phone measurement.

The owner confirmed there is no Mac build machine and no real-photo dataset.
Native compilation, recognition evaluation, simulator/device runs, and offline
phone acceptance remain unrun. Temporary-loose-tile certificate completeness
also remains an explicit engine limitation; unresolved bounds stay incomplete.
Open PR bodies are updated with these limits and dependencies. #12 stays draft.
Main and #9 were rechecked unchanged; no PR was merged, force-pushed, or deployed.
