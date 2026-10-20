# Offline Rummikub assistant

Implement the personal assistant described in the owner's CODEX_HANDOFF.md
(2026-10-19). The final workflow is two separate local photos (table, then rack),
visible uncertainty and corrections, confirmed-position validation, legal move
search, and exact final-table/remaining-rack rendering on a phone. No runtime
server, inference API, photo upload, or remote solver is permitted.

- Edition 4: two numbered copies and two jokers (106 tiles). Edition 6: three
  numbered copies and four jokers (160 tiles); both have four colors.
- New position commands require explicit `--edition=4|6`.
- Opening is a rack-only 30 points and does not manipulate the existing table.
  `--opening-required=false` skips this phase, independent of player `--opened`.
  Drawing ends the turn. Opening state never changes table validity.
- Maximize rack tiles played; allow unlimited legal table rearrangement. Stable
  enumeration is the initial engineering tie-break, with no strategic claim.
- Conserve identified physical table tiles; consume exact rack occurrences.
  Jokers retain identity independently of their roles. Verify every returned move.
- Distinguish optimal, feasible/incomplete, proven no move, incomplete/no result,
  invalid or unresolved input, and unsupported rule coverage. A resource limit is
  never proof of no move. Search must support cancellation and bounded resources.
- Render every final table and rack occurrence, rack additions, and joker roles.
- Preserve Go core and `go run .`. Structured JSON is a development interface,
  not a substitute for recognition or the phone app.

Rules reference: booklet D-2600-1236-0041 (120819), linked in docs/rules.md.
PR #9's selected profile includes joker replacement and illustrated releases by
restructuring. The actual booklet was downloaded and its illustrations visually
checked after network publication; see docs/joker-transitions.md for the source
hash, regressions, chained certificates and remaining proof boundaries. Do not
silently mix profiles or invent another house rule.

Acceptance needs real representative photos and a physical-phone end-to-end run
with networking disabled and all assets local. Synthetic tests, desktop runs,
WebAssembly builds, and mock predictions alone do not establish that acceptance.
Do not commit or upload private photos. Model/runtime licenses, reproducible
training/export, actual device/OS performance, and ambiguity correction are part
of the remaining acceptance work. Device/model/latency targets are not fixed.

## Integration status and acceptance prerequisites

The owner selected iPhone first. The initial source integration targets iOS 17+
using native camera/import/Vision proposals and the bundled Go/Wasm engine; this
is not a measured final minimum device/OS or an accepted model choice. The local
interface and actual Go bridge are tested on desktop; native compilation, real
recognition, physical-device memory/latency, and the end-to-end offline workflow
are unverified. The owner confirmed there is no Mac build machine or real-photo
dataset yet. PRs #9–#11 are merged into main; #12 targets main and remains a draft.
The owner authorized a static GitHub Pages Safari preview and then requested
automatic browser tile reading. The browser-reader branch depends on #12 and
bundles local Tesseract OCR, tile/ink segmentation, color proposals and suggested
meld rows. Model/runtime downloads happen at build time; CSP still prohibits
runtime connections, uploads and remote inference. All proposals require explicit
confirmation. The preview does not persist photos/readings or support offline
reopening. Synthetic actual-OCR checks do not establish real-photo accuracy,
native acceptance, or measured iPhone memory/latency. The owner's two inline
examples are visually transcribed but unavailable as original files to the runner.

The owner requested a Rummikub-style interface: green felt, ivory numbered
tiles with red/blue/black/orange accents, and wooden rack details. Decorative
tiles are separate from the confirmed table/rack rendering. Phone layouts stack
the two photo panels; wider layouts show them side by side. Local processing,
explicit reading confirmation, search coverage messages, and edition/opening
state remain part of the interface contract.
The owner additionally requested actual tile artwork for the confirmed table and
identified <https://www.printables.com/model/130082-multicolor-rummikub-tiles> as
the desired source. That host is blocked by the current environment network
policy; page/file/API hosts were saved in the draft and require publication.
Inspect the model files and license, then render a local tile atlas when those
assets are accessible. Existing exact diagrams remain in place meanwhile. Do not
claim CSS boxes or generated replacement art are the requested model assets, and
do not substitute the user's photos for them.
