# Local model-based tile artwork

The confirmed table and remaining rack use actual geometry from
[Multicolor Rummikub tiles by KlaasYK](https://www.printables.com/model/130082-multicolor-rummikub-tiles),
retrieved 2026-10-20. Number shapes and smiley jokers come from the author's mesh.
The adaptation reorients the model, assigns ivory plastic and four ink colors,
and renders a transparent atlas. Source and adapted artwork are
[CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
Credits and changes are visible in the app's **Tile artwork credits** and bundled
in `phone/web/tile-art/ATTRIBUTION.txt`; this license applies to the artwork,
separately from application source and dependencies.

The author's Printables model PDF, page 3, explicitly states “This work is
licensed under a Creative Commons (4.0 International License) Attribution-ShareAlike”
and links to that license. Downloaded PDF SHA256:
`482725bd2bc274c43b4f729a3a62c050eceb4df6c5774299f42b39dd70aa4969`.
The PDF and download evidence remain under
`/workspace/rummy-review/printables-130082/evidence/`. The 3MF LicenseTerms field
was empty; the author-provided PDF supplies the explicit license.

The source OBJ contains fourteen actual rounded bodies and complementary inlay
meshes, arranged as 1–13 and a joker. It references a material file not supplied
by the author; the renderer assigns materials. No photographs or font-generated
replacement numbers are included.

## Reproduce the artwork

Rendering is an optional build-time task with Blender 4.3.2, Cycles CPU. Phone
execution uses the committed PNG and needs neither Blender nor network access.
Download the original mesh, then run the pinned renderer:

```sh
mkdir -p /tmp/rummy-art-source
curl --fail --location 'https://files.printables.com/media/prints/130082/stls/1246447_3ba28a3e-b369-4652-887f-75c7ba08219f/rummikub_joker.obj' --output /tmp/rummy-art-source/rummikub_joker.obj
blender --background --threads 8 --python-exit-code 1 --python scripts/render-tile-art.py -- --source /tmp/rummy-art-source/rummikub_joker.obj --output /tmp/rummy-art-render
sha256sum /tmp/rummy-art-render/tiles-atlas.png
```

The script rejects source bytes other than SHA256
`cd92b514f2f435f1877b64e70e71c14716319e3bebdfa73c654f81cec905d1bc`.
It uses 96 samples and fixed seed 130082, explicit geometry/materials/lights,
and strips Blender's timestamp/path PNG text metadata without changing pixels.
Two separate renders on Blender 4.3.2 produced identical normalized PNG bytes:
`37dfd3f03bd4d6281c4f359c10f2cdd06adbe716d274bfed1011496d332323a6`.
Different Blender versions are not claimed to reproduce identical pixels.

The atlas is 2688×1152 RGBA, 3,511,134 bytes, with 192×288 cells: columns 1–13
and joker; rows red, blue, black and orange. `tiles-atlas.json` documents all
56 rectangles and attribution. The app uses a black physical joker image and
an explicit visible role caption. Every occurrence keeps its unique ID, title,
accessibility label and ordering; played tiles keep a transparent green outline.
The Go SVG remains self-contained for CLI export. The browser decorates its
validated groups with local artwork using explicit physical tile/role metadata.

## Verify integration

```sh
bash scripts/build-phone.sh
python3 scripts/check-tile-art.py --screenshot /tmp/rummy-art-diagram.png
python3 scripts/check-phone-ui.py
```

The artwork check runs actual Go/Wasm validation and draws a synthetic six-player
position containing all 52 numbered faces, four physical jokers, and duplicate
occurrences. It checks exact IDs/order/roles, source-art hash, all sprite crops,
accessible labels, idempotence and local-only image loading. The UI test runs
an actual opening search and checks artwork plus all three played outlines.
These desktop checks do not establish iPhone compatibility, recognition quality,
or phone memory/latency.
