# iPhone-first offline integration

The phone interface uses the same Go application and rule/transition checker as
the CLI. It bundles WebAssembly bytes and the Go runtime as local script assets;
it does not run a server or fetch a model/solver. CSP disallows connections and
workers. This avoids file-URL worker/fetch restrictions in WKWebView, while Go's
periodic yielding lets cancellation and UI events run. Node tests execute the
actual Wasm engine and verify cancellation retains its best verified move.

## Reproduce portable validation

From the repository root, with supported Go, Python 3, and Node:

```sh
bash scripts/check.sh
bash scripts/build-phone.sh
node scripts/check-phone-engine.cjs
node scripts/check-photo-state.cjs
```

Optional desktop UI integration (Python Playwright, Pillow, Chromium installed):

```sh
python3 scripts/check-phone-ui.py
python3 scripts/check-phone-startup.py
```

The UI test fulfills all local assets in-process with network disabled. It uses
synthetic photo observations solely to test correction and real-engine wiring;
it does not evaluate recognition. It checks uncertainty gates, separate photo
identity, raw/corrected value isolation, exact occurrence rendering, and result
invalidation when inputs change. The Safari package additionally bundles a local
OCR reader; its separate actual-pixel checks are documented below.

## Safari preview

Open the [Safari preview](https://mgaulia.github.io/rummy-analizer/) on any network.
GitHub's Pages API confirms a successful build of `gh-pages` commit
`18a2d10976fcc6460ca2954c3a72bd7bc1d4f1cc`, with HTTPS enforced, publishing
source `88101cd91e700bd5420ad45e4e311aab95188387`. All seventeen live HTTPS assets
matched the tested package hashes. The downloaded package passed actual-photo
import/read/solve/render and compact layout checks in desktop Chromium. Melds
fit without internal scrollbars at 320, 390, 768 and 1200px; played tiles have
contrasting amber outlines on light cards. Earlier startup/OCR checks are
recorded with their exact source checkpoints in docs/implementation-plan.md.
Physical Safari execution of this revision remains unverified. No source merge
was needed; #12 and its reader follow-up
#13, artwork #14, streamlined flow #15 and adaptive photo lighting #16 remain
unmerged drafts. Merge #12 → #13 → #14 → #15 → #16.
The current preview uses a green felt table, ivory numbered tiles, a wooden
rack, and the four tile colors. Assets include a source-version query parameter
to refresh the styling after publication. The design adapts to phone and desktop
widths. The new browser-reader branch adds local suggestions with mandatory
direct edits and blocking parse errors; the solver is unchanged.

The owner authorized publishing the draft #12 browser interface on GitHub Pages.
This does not merge #12 or accept its native integration. Build from a clean,
committed checkout with the same Go/Python prerequisites as above:

```sh
bash scripts/build-safari-preview.sh /tmp/rummy-safari-preview
python3 scripts/check-safari-preview.py /tmp/rummy-safari-preview
python3 scripts/check-browser-reader.py
python3 scripts/check-browser-reader.py --preview-directory /tmp/rummy-safari-preview
python3 scripts/check-tile-art.py --preview-directory /tmp/rummy-safari-preview
```

Use an empty output directory. Publish only its contents to the `gh-pages`
branch root, and configure Pages to deploy that branch. No workflow is needed.
The generated `preview-build.json` records the source commit/tree, Go version,
and deployed asset hashes. No photos, credentials, or native project are included.
The package check needs the optional desktop UI dependencies listed above and
checks manual file import/correction and solving at the nested Pages URL path.
It runs desktop Chromium at a phone viewport, not Safari or an iPhone.

In iPhone Safari, open the HTTPS Pages URL on any network. Four players is the
default; six players remains available. Use **Add photo** separately for table
and rack. Readable values work immediately; edit a value if needed, use the red
X to remove a detection, or add a missed tile. Unreadable values and invalid
edits block search with an error. Table meld and order fields remain editable.
No confirmation clicks are required before **Check table** or **Find move**. An empty table
photo is allowed. Photos and solving stay in the browser. Reloading clears the
readings; offline reopening is not supported. The native **Take photo** buttons
are omitted from this preview. Safari's file picker may offer its own camera.

For a small first test, photograph an empty table and a separate rack with red
10, blue 10, and black 10. Leave the default edition and opening requirement.
After the photos are read, **Find move** should show a 30-point opening that plays
three tiles. If opening already happened, select **I have opened**; that control
only applies when the opening rule is enabled. There is one internal search
budget and a cancel button, with honest incomplete/proven statuses. Record the iPhone
model, iOS version, source from the downloaded report, result, and any errors.
This tests manual correction and the actual solver, not recognition accuracy.

## Local browser reader

`scripts/build-browser-reader.py` downloads the pinned Tesseract.js-core 6.1.2
runtime and an upstream tessdata_fast English model **at build time**, verifying
their checksums. The preview bundles the runtime, WebAssembly bytes, model and
Apache-2.0 notices as local script assets. There is no runtime CDN/model request,
worker, photo upload, or remote inference. CSP still blocks all connections.
The Go engine and native Apple Vision adapter are unchanged. The model increases
the initial page download; phone memory and latency still need measurement.

The reader bounds image size to 1400 pixels, finds ivory components, splits
touching strips, crops ink away from the embossed logo, runs actual OCR, proposes
colors (including cyan/teal as a blue candidate), and suggests meld rows. A
face-like joker is only a low-confidence proposal. Scores are uncalibrated; no
reading is treated as human-confirmed. Per the owner's updated preference, a
single readable proposal is used immediately; unresolved values block search.
Rotation, perspective, shadows, overlapping
tiles, unusual fonts/backgrounds and joker designs can cause missed/wrong values.
Use a clear, upright, close photo and review completeness. A failed reader shows
an error and requires another photo; it cannot become a successfully empty table.

Reader revision 2 suppresses spatially nested regions inside a detected tile
(such as the ivory area enclosed by a joker face). Separate physical tiles with
identical values stay separate. It checks full-number OCR in multiple modes and
compares it with visible tall digit components; a partial one-digit result cannot
override a clearly two-digit image because of its score. Conflicts can use local
per-digit OCR; unresolved conflicts remain blank for manual correction. Crops
cover a wider ink area while excluding the lower embossed logo. This improves
specific regression cases, not a measured general accuracy guarantee.

After building reader assets, run `python3 scripts/check-browser-reader.py`.
This executes the real bundled OCR on **synthetic** images of all 13 numbers and
colors at normal, 70% and 60% brightness, touching strips and a face-like joker,
and then imports images through
the actual file picker and immediately solves a legal opening. It checks
that only bundled GET assets were requested and that CSP produced no errors.
These generated fixtures are not real-photo accuracy or iPhone evidence. A
subsequent owner-shared reading report supplied the two original photos for local
replay: reader revision 3 proposes all 14 manually labeled tiles correctly,
including the rack's red 11, with exact counts and suggested table melds.

Reader revision 4 addresses the later dim iPhone photos: ivory detection uses
75% of the image's 98th-percentile mean RGB luminance, bounded to 110–165, instead
of a fixed 165 cutoff. Diagnostic metadata records the reference and threshold.
A narrow saturated blue/teal rule recovers dim colored ink that the former
absolute filters dropped or proposed as black; warm-black joker handling stays
unchanged. The later two originals now propose all 22 labeled tiles, while the
earlier 14 remain correct. These four-photo desktop regressions are not
representative recognition evaluation or physical-iPhone acceptance. Very dark,
mixed-lighting, low-contrast and differently framed scenes remain unmeasured.

To export proposals for an accessible original photo without uploads:

```sh
python3 scripts/check-browser-reader.py --image /external/table.jpg --zone table --output /external/table-proposals.json
```

Use an independently annotated manifest with `scripts/evaluate-photos.py` for
accuracy metrics. Store originals, labels and proposal exports outside Git.

If a photo still produces duplicates or wrong readings in Safari, expand
**Reading report** at the bottom and choose **Save report**. The
download contains the selected original photos, raw proposals, model/build
revision, browser details, rule settings and latest search/error outcome;
corrections are not mislabeled as recognition.
Nothing is uploaded. The user can choose whether to share that JSON file for
debugging. This explicit export is the only added persistence; reloading still
clears the active session. It is a browser feature, hidden in the native adapter.
Rerun the actual images from a report locally:

```sh
python3 scripts/check-browser-reader.py --report /external/rummy-reading-report.json --output /external/rechecked.json
```

The checker verifies embedded photo hashes and never follows remote image URLs.
Add `--expectations /external/expected.json` to assert an exact private regression.
The JSON maps each photo key to `source` (SHA256), `zone`, and an ordered `tiles`
array containing `tile`, `meld` and `order` per occurrence. Label these from the
photos independently of proposals; keep the report and labels outside Git.
The truncation regression deliberately injects a high-scoring partial output in
one OCR mode while other modes still run the real engine; this fault injection
tests the selection guard and is separate from actual-OCR fixture checks. Neither
that test nor the synthetic images establish general photo accuracy.

The confirmed table and remaining rack now use local artwork rendered from the
owner-selected [Printables model](https://www.printables.com/model/130082-multicolor-rummikub-tiles).
Physical occurrences, played outlines and explicit joker roles remain exact.
See [artwork source, license and reproduction](../docs/tile-art.md).
`python3 scripts/check-tile-art.py` checks all numbered sprites and joker roles
on a real Go/Wasm diagram; `--preview-directory` checks a generated/live package.
`python3 scripts/check-table-view.py` checks responsive meld cards, physical IDs,
roles, rack counts and compact meld wrapping without internal scrollbars. The private report regression below
imports original photo bytes and checks no opening, then an independently
verified four-tile optimum after opening (it is specific to that labeled case):

```sh
python3 scripts/check-reading-flow.py --report /external/report.json --expectations /external/expected.json --expected-played 4 --output /external/flow-results.json
```

The expectation format is the same hash-bound ordered tile/meld JSON described
above. Use `--expected-played 1` for the later 21-table-tile/red-1-rack report.
That case cannot make a 30-point opening; when already opened or opening is
disabled, it has a verified optimum that empties the rack. Add
`--preview-directory` to verify the exact generated/downloaded build.

## Build the iOS integration

Requires a Mac with Xcode 15+ and iOS 17+ SDK. iOS 17 is an initial integration
target, not verified handset compatibility or the final product minimum.

1. Run `bash scripts/build-phone.sh` with Go and Python installed on that Mac.
   The generated `phone/web/engine` directory is ignored; regenerate it from
   reviewed source. The script copies the matching Go runtime and BSD license;
   the app exposes the license locally under Open-source licenses.
2. Open `phone/ios/RummyAssistant.xcodeproj` in Xcode. The blue `web` folder
   reference must include all web assets, including the generated engine.
3. Select the `RummyAssistant` scheme. For a simulator, run without a signing
   team. For a physical phone, select your existing development team and your
   phone. No signing account, paid service, or deployment is created here.
4. Or compile for a simulator from the repository root:

   ```sh
   xcodebuild -project phone/ios/RummyAssistant.xcodeproj \
     -scheme RummyAssistant -sdk iphonesimulator -configuration Debug \
     -derivedDataPath phone/ios/build CODE_SIGNING_ALLOWED=NO build
   ```

These native build steps have NOT executed in the Linux environment. A successful
Wasm build or parsed project file does not prove Swift compilation or WebKit/iOS
compatibility. The app has real camera/import, local Vision proposal code, and
engine integration source; it is not accepted as a completed phone app yet.

## Recognition and correction

The native adapter normalizes and downsamples an image to at most 2000 pixels on
its longest side, proposes tile rectangles using Vision, reads number candidates
using installed on-device text recognition, and proposes glyph colors from
local pixels. Apple Vision/UIKit/WebKit/CryptoKit are system frameworks; there
are no downloaded model weights or third-party runtime licenses. This is a
candidate pipeline, not a calibrated or measured recognition model. Unrecognized
joker faces stay uncertain and must be corrected; text-only joker detection is
not full joker recognition. Rectangle/grouping/color heuristics require real
representative-photo evaluation before acceptance.

Each proposal retains its raw text, candidates, image rectangle, and source
identity. Edits are separate. Readable proposals are used automatically, while
unknown or ambiguous values and invalid edits block search. Users can edit tiles,
melds and order, add missed tiles, or remove detections without confirmation steps.
The engine rejects invalid positions rather than changing readings to fit.
Table-only validation does not require a rack. Search requires a second,
separate rack photo. Reusing identical photo bytes is rejected; users
must also check the photographs do not contain overlapping physical tiles.
Photos stay in memory in the local view and are not committed, persisted, or
uploaded. Import/camera permissions use the operating system's normal controls.

## Remaining acceptance

- Compile/run the native target on macOS, then a real iPhone; fix native issues
  and verify camera/import, WebKit/Wasm cancellation and memory behavior.
- Obtain consented representative table/rack photographs and ground truth,
  including both joker designs/editions, lighting, angles, fonts, and occlusions.
  Keep these outside the repository. Evaluate detection, OCR, colors, jokers,
  group boundaries, correction usability, and latency/memory on the target phone.
- Use measurements to decide whether the Vision pipeline is sufficient or a
  compact specialized detector/classifier is needed. Do not invent accuracy.
- With all assets local, disable networking on the phone and test the entire
  two-photo/correction/validate/solve/render/cancel workflow. Record actual device,
  OS, performance, expected/resulting tiles, and any unsupported rule cases.
- Complete official joker-rule coverage before claiming complete optimal search
  for tables containing jokers. Current search reports that limit explicitly.

## Reproducible local recognition evaluation

`scripts/evaluate-photos.py` matches raw proposals to independently annotated
rectangles by IoU >= 0.5 (maximum-cardinality spatial matching, independent of
labels). It reports detection precision/recall, exact top-1 tile accuracy,
end-to-end tile recall, unresolved matches, and exact table grouping. Real and
synthetic samples are never pooled. It checks image SHA-256, prediction source,
zone, explicit consent, and declared provenance. It cannot independently prove
those provenance/consent declarations. No real dataset has been evaluated here.

Keep the dataset outside this checkout. A version-1 manifest has `samples`, each
with `id`, `image` (relative to manifest), `sha256`, `zone` (`table`/`rack`),
`origin` (`real`/`synthetic`), `consented: true`, and `tiles`. Each ground-truth
tile has normalized top-left `rect: {x,y,width,height}`, a confirmed `tile`
(`{color,number}` or `{joker:true}`), and table `meld` label. Use the normalized
JPEG the native adapter actually analyzes, not the original oriented photo.

Predictions are keyed by sample ID and contain `source` (normalized image hash),
`zone`, and `observations` (`rect`, raw `candidates` in proposal order, optional
`suggestedMeld`). On a development device, the WebKit inspector's
`rummyObservationReport()` returns raw observations and recognizer metadata;
it omits image bytes and corrections. Only with the owner's consent, retain
local normalized images/ground truth for evaluation. Do not commit/upload them.
Record the reported OS/pipeline/Vision revisions alongside the dataset report.

```sh
python3 scripts/evaluate-photos.py /external/consented/manifest.json \
  /external/consented/raw-predictions.json
python3 -m unittest discover -s scripts -p 'test_photo_evaluation.py' -v
```

The evaluator's two synthetic unit fixtures verify metric arithmetic and input
integrity, not recognition. Actual phone latency, memory, accuracy, occlusion
coverage, and offline end-to-end behavior remain required measurements.

Current rulebook-backed engine improvements include chained release/reuse and
initial-role ambiguity handling. A verified move reaching the exhaustive relaxed
partition bound proves optimality, and a zero bound proves no move. Other
unresolved temporary-loose-tile cases stay incomplete. See
[proof boundaries](../docs/joker-transitions.md). The owner confirmed no Mac build
machine and no real-photo dataset are available yet; native/recognition
acceptance remains blocked by those actual prerequisites, not network access.
