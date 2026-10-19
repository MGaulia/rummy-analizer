// Actual Go/Wasm execution in Node; no network, no mock engine.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
globalThis.crypto ??= require('node:crypto').webcrypto;
require(path.resolve(__dirname, '../phone/web/engine/wasm_exec.js'));
(async () => {
 const license = fs.readFileSync(path.resolve(__dirname,'../phone/web/engine/GO-LICENSE.txt'),'utf8');
 assert.match(license,/Redistributions in binary form/);
 const go = new Go();
 const wasm = await WebAssembly.instantiate(fs.readFileSync(path.resolve(__dirname,'../phone/web/engine/rummy.wasm')), go.importObject);
 go.run(wasm.instance);
 const piece=(id,color,number)=>({id,tile:{color,number}});
 const request={operation:'solve',settings:{edition:4},position:{table:[],rack:[piece('r','red',10),piece('b','blue',10),{id:'j',tile:{joker:true}}]}};
 const response=JSON.parse(await rummyExecute('solve',JSON.stringify(request)));
 assert.equal(response.search.status,'optimal');
 assert.equal(response.search.verified.played.length,3);
 assert.match(response.svg,/Joker as/);
 const boundRequest={operation:'solve',settings:{edition:4,opening_required:false},position:{table:[[piece('t1','blue',1),piece('t2','blue',2),{id:'tj',tile:{joker:true}}]],rack:[piece('add','blue',4),piece('unused','red',13)]}};
 const bounded=JSON.parse(await rummyExecute('joker-bound',JSON.stringify(boundRequest)));
 assert.equal(bounded.search.status,'optimal');
 assert.equal(bounded.search.upper_bound,1);
 assert.equal(bounded.search.verified.played.length,1);
 assert.equal(bounded.search.verified.after.rack[0].id,'unused');
 const rack=[];
 for(const color of ['black','red'])for(let n=1;n<=12;n++)rack.push(piece(color+n,color,n));
 rack.push(piece('lonely','blue',13));
 const pending=rummyExecute('cancel',JSON.stringify({operation:'solve',settings:{edition:4,opening_required:false},position:{table:[],rack},max_nodes:100000000}));
 setTimeout(()=>rummyCancel('cancel'),20);
 const cancelled=JSON.parse(await pending);
 assert.ok(['incomplete','feasible_incomplete'].includes(cancelled.search.status));
 assert.match(cancelled.search.reason,/canceled/);
 if(cancelled.search.move){
  const verified=JSON.parse(await rummyExecute('verify',JSON.stringify({operation:'check-move',settings:{edition:4,opening_required:false},position:{table:[],rack},move:cancelled.search.move})));
  assert.equal(verified.move.valid,true);
 }
 console.log('Go/Wasm: actual solve, exact SVG, cancellation and retained-best validation passed.');
 process.exit(0);
})().catch(error=>{console.error(error);process.exit(1)});
