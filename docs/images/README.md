# README screenshots

- `ui-overview.png`: the browser's opening settings and separate table/rack inputs.
- `ui-move.png`: a verified four-tile move, joker role and one remaining rack tile.

Captured in Chromium 154.0.8037.57 from application source
`02c7801adda6c9eae8c62fbe6a520f7a23a3ed02`, using the normal static preview build.
Input images were generated from the bundled tile atlas on a drawn background;
no personal photos or reading reports are included. The actual local reader
proposed all eight demo tiles correctly. Values were reviewed through the normal
UI, and the actual Go/Wasm solver returned an optimal four-tile move. A separate
`check-move` request validated it; all eight physical occurrences were rendered.
These screenshots illustrate usage, not representative photo accuracy or a
physical-phone acceptance test. The overview is cropped above the editor rows;
the result is captured at a 900px viewport using the app's normal responsive UI.

The depicted tile artwork is adapted from **Multicolor Rummikub tiles** by
[KlaasYK](https://www.printables.com/model/130082-multicolor-rummikub-tiles).
The screenshots are distributed under
[CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
Changes include rendering the original tile meshes in ivory and four colors,
compositing them into demo inputs, and capturing them in the assistant UI.
See [tile-art source and reproduction](../tile-art.md) for the mesh, atlas and
license details. This notice applies to these images, not the separate app code.
