/* Script-loaded bytes avoid file-URL fetch/worker restrictions in WKWebView. */
window.engineReady = (async () => {
  if (!globalThis.RUMMY_WASM_BASE64) throw new Error('Build the bundled Go engine first');
  const bytes = Uint8Array.from(atob(RUMMY_WASM_BASE64), ch => ch.charCodeAt(0));
  delete globalThis.RUMMY_WASM_BASE64;
  const go = new Go();
  const result = await WebAssembly.instantiate(bytes, go.importObject);
  go.run(result.instance);
})();
