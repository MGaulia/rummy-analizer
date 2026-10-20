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
`eb2e7e041e64ab18df5ec187a22dcff67cd1a298`, with HTTPS enforced. The initial
activation request from this session was denied, but Pages was subsequently
enabled. The workspace's direct public-URL request is still blocked by its
pending network configuration; live HTTP and physical Safari execution have not
been verified here. No source merge was needed; #12 remains an unmerged draft.
The current preview uses a green felt table, ivory numbered tiles, a wooden
rack, and the four tile colors. Assets include a source-version query parameter
to refresh the styling after publication. The design adapts to phone and desktop
widths. The new browser-reader branch adds local suggestions with mandatory
correction/confirmation; the solver is unchanged.

The owner authorized publishing the draft #12 browser interface on GitHub Pages.
This does not merge #12 or accept its native integration. Build from a clean,
committed checkout with the same Go/Python prerequisites as above:

```sh
bash scripts/build-safari-preview.sh /tmp/rummy-safari-preview
python3 scripts/check-safari-preview.py /tmp/rummy-safari-preview
python3 scripts/check-browser-reader.py
```

Use an empty output directory. Publish only its contents to the `gh-pages`
branch root, and configure Pages to deploy that branch. No workflow is needed.
The generated `preview-build.json` records the source commit/tree, Go version,
and deployed asset hashes. No photos, credentials, or native project are included.
The package check needs the optional desktop UI dependencies listed above and
checks manual file import/correction and solving at the nested Pages URL path.
It runs desktop Chromium at a phone viewport, not Safari or an iPhone.

In iPhone Safari, open the HTTPS Pages URL on any network. Choose an edition,
then use **Choose photo** separately for table and rack. The bundled reader
proposes tiles, colors, numbers and meld rows. Correct every uncertain reading,
tap missed tiles to add them, exclude false detections, and confirm every tile.
For table tiles check the meld number and preserve the displayed order. Confirm
each complete reading before **Check table** or **Find move**. An empty table
photo is allowed. Photos and solving stay in the browser. Reloading clears the
readings; offline reopening is not supported. The native **Take photo** buttons
are omitted from this preview. Safari's file picker may offer its own camera.

For a small first test, photograph an empty table and a separate rack with red
10, blue 10, and black 10. Choose the four-player edition and leave the opening
requirement enabled. After entering and confirming all three rack tiles, the
solver should show a 30-point opening that plays three tiles. Record the iPhone
model, iOS version, preview source shown on the page, result, and any errors.
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
reading is automatically confirmed. Rotation, perspective, shadows, overlapping
tiles, unusual fonts/backgrounds and joker designs can cause missed/wrong values.
Use a clear, upright, close photo and review completeness. Failures retain manual
entry. The unbundled source page also falls back to manual entry.

After building reader assets, run `python3 scripts/check-browser-reader.py`.
This executes the real bundled OCR on **synthetic** images of all 13 numbers,
colors, touching strips and a face-like joker, and then imports images through
the actual file picker, confirms readings and solves a legal opening. It checks
that only bundled GET assets were requested and that CSP produced no errors.
This is not real-photo accuracy or iPhone evidence. The two owner photos are
visible in chat but their original bytes are not mounted in this workspace, so
their end-to-end readings have not been evaluated here.

To export proposals for an accessible original photo without uploads:

```sh
python3 scripts/check-browser-reader.py --image /external/table.jpg --zone table --output /external/table-proposals.json
```

Use an independently annotated manifest with `scripts/evaluate-photos.py` for
accuracy metrics. Store originals, labels and proposal exports outside Git.

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
identity. Corrections are separate. Users confirm each tile, meld boundary and
within-meld order, add missed tiles, exclude false detections, and acknowledge
completeness. The engine rejects invalid positions rather than changing readings
to fit. Table-only validation does not require a rack. Search requires a second,
separately confirmed rack photo. Reusing identical photo bytes is rejected; users
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
