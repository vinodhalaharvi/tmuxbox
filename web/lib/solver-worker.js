// Runs the Go solver (compiled to WebAssembly) off the main thread.
importScripts('wasm_exec.js');
const go = new Go();
const ready = fetch('bean.wasm').then(r => r.arrayBuffer()).then(b => WebAssembly.instantiate(b, go.importObject))
  .then(({instance}) => { go.run(instance); });
onmessage = e => {
  ready.then(() => {
    const t0 = performance.now();
    const json = beanSolve(e.data.aspects, e.data.weights);
    postMessage({id: e.data.id, json, ms: performance.now() - t0});
  }).catch(err => postMessage({id: e.data.id, error: String(err)}));
};
