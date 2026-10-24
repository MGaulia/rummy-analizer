# Offline Rummikub assistant

Implement the personal assistant described in the owner's CODEX_HANDOFF.md
(2026-10-19). The final workflow is two separate local photos (table, then rack),
visible uncertainty and corrections, position validation, legal move
search, and exact final-table/remaining-rack rendering on a phone. No runtime
server, inference API, photo upload, or remote solver is permitted.

The root README presents two user entry points: the browser UI with separate
table/rack photos, and the Go CLI with JSON positions or single-meld text input.
It includes app screenshots using generated demo inputs, correction/opening
instructions, and independently checked CLI examples. Internal APIs, native
build steps and migration history stay in developer references. Demo screenshots
illustrate usage and do not establish real-photo accuracy or phone acceptance.

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
are unverified. There is no Mac build machine; a representative recognition
dataset beyond the four shared photos remains unavailable. PRs #9–#16 are
merged into main. The owner authorized merging the remaining stack on
2026-10-20; #12 → #13 → #14 → #15 → #16 landed in that order, each directly into
main. Merged source and fresh verification are recorded in the implementation
plan; merging does not establish native/device acceptance.
The owner approved merging independent maintenance PRs #19–#21 into main on
2026-10-24. Their GitHub records provide the current merge status.
Their contracts preserve search enumeration and work accounting, physical identities,
raw/correction isolation and current photo-error behavior. They remove unused
search allocations, avoid repeated photo decoding and preserve cancellation
ownership when a request ID is reused. Combined verification and merge guidance
are recorded in the implementation plan; native/device acceptance is unchanged.
The owner authorized a static GitHub Pages Safari preview and then requested
automatic browser tile reading. The integrated browser reader
bundles local Tesseract OCR, tile/ink segmentation, color proposals and suggested
meld rows. Model/runtime downloads happen at build time; CSP still prohibits
runtime connections, uploads and remote inference. Per the owner's 2026-10-20
feedback, one readable proposal is used automatically. There are no per-tile or
photo confirmation gates; uncertain readings, invalid drafts and read failures
block search with an error. Edits apply immediately, and a red X removes a
detection. Raw proposals stay separate from edits in diagnostics.
The preview does not automatically persist photos/readings or
support offline reopening. An explicit local reading-report download includes
selected originals, raw proposals and build/browser information for reproducible
debugging; it never uploads or shares the report automatically. Synthetic actual-OCR checks do not establish real-photo accuracy,
native acceptance, or measured iPhone memory/latency. The owner subsequently
shared a reading report containing the two original examples. Reader revision 3
passes exact local replay of their 14 manually labeled tiles, counts and suggested
melds. A later iPhone report added two dimmer originals with 22 labeled tiles.
Reader revision 4 adapts face brightness to the image's 98th-percentile luminance
and recovers saturated dim blue/teal ink, retaining the established warm-black
joker handling. All 36 occurrences across these four photos pass exact desktop
replay, including duplicate counts and suggested melds. This bounded regression
does not establish representative dataset accuracy or physical-iPhone acceptance
of the updated reader. Dominant ink filtering prevents
cross-color background slivers from contaminating number crops; same-color
artifacts and other capture conditions remain unmeasured.

The owner requested a Rummikub-style interface: green felt, ivory numbered
tiles with red/blue/black/orange accents, and wooden rack details. Decorative
tiles are separate from the confirmed table/rack rendering. Phone layouts stack
the two photo panels; wider layouts show them side by side. Local processing,
blocking parse errors, search coverage messages, and edition/opening state remain
part of the interface contract. The phone page defaults to four players and uses
one internal 200,000-node search setting with cancellation. The user's report
reaches the same proven optimum at both former budgets; no general or phone
speed claim follows. Opening is enabled by default, and “I have opened” is only
available while that rule is enabled; disabling it preserves separate player
history. The drawn-turn control is omitted from the
phone page; drawing still ends the turn in the Go rules/CLI. An unavailable
30-point opening gets a specific result message. Preview branding, tutorials
and search-effort controls are removed. Responsive meld cards and a wrapping
remaining rack replace the fixed-width browser diagram; occurrence IDs, played
markers, order and joker roles remain exact.
Meld cards use a light neutral background and amber played-tile outlines.
Cards fit their contents, tiles sit close together, and long melds wrap inside
one labeled card without an internal scrollbar. Joker-role captions remain
readable below their tiles; the board heading is “Final table”.
The owner additionally requested actual tile artwork for the confirmed table and
identified <https://www.printables.com/model/130082-multicolor-rummikub-tiles> as
the desired source. Published network access now permits downloading the actual
model files and inspecting their CC BY-SA 4.0 license. A locally rendered atlas
now uses the author's actual numeral, joker and tile-body meshes. Browser/app
diagrams decorate validated Go positions with these bundled images, preserving
each occurrence, played marker and explicit joker role. See docs/tile-art.md for
source/license evidence, pinned mesh hash and reproducible Blender rendering.
No user photos are used as artwork. The CLI SVG remains self-contained.
