/* Local bundled assets only. No remote endpoint is used by this worker. */
importScripts('engine/wasm_exec.js');
const go = new Go();
const ready = fetch('engine/rummy.wasm').then(r => {
  if (!r.ok) throw new Error('Bundled engine is missing');
  return r.arrayBuffer();
}).then(bytes => WebAssembly.instantiate(bytes, go.importObject)).then(result => {
  go.run(result.instance);
});
onmessage = async ({data}) => {
  try {
    await ready;
    if (data.cancel) { rummyCancel(String(data.id)); return; }
    const response = JSON.parse(await rummyExecute(String(data.id), JSON.stringify(data.request)));
    postMessage({id: data.id, response});
  } catch (error) { postMessage({id: data.id, response: {error: String(error)}}); }
};
ready.then(() => postMessage({ready: true})).catch(error => postMessage({fatal: String(error)}));
