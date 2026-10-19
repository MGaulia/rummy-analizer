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
invalidation when inputs change. Browser-only imports require manual annotation;
no browser recognizer is claimed.

## Safari preview

The owner authorized publishing the draft #12 browser interface on GitHub Pages.
This does not merge #12 or accept its native integration. Build from a clean,
committed checkout with the same Go/Python prerequisites as above:

```sh
bash scripts/build-safari-preview.sh /tmp/rummy-safari-preview
python3 scripts/check-safari-preview.py /tmp/rummy-safari-preview
```

Use an empty output directory. Publish only its contents to the `gh-pages`
branch root, and configure Pages to deploy that branch. No workflow is needed.
The generated `preview-build.json` records the source commit/tree, Go version,
and deployed asset hashes. No photos, credentials, or native project are included.
The package check needs the optional desktop UI dependencies listed above and
checks manual file import/correction and solving at the nested Pages URL path.
It runs desktop Chromium at a phone viewport, not Safari or an iPhone.

In iPhone Safari, open the HTTPS Pages URL on any network. Choose an edition,
then use **Choose photo** separately for table and rack. Browser imports have no
recognizer: tap each tile to add it, enter color/number or joker, and confirm it.
For table tiles assign a meld number and preserve the displayed order. Confirm
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
