'use strict';
const {PhotoSession, positionFromPhotos} = RummyPhotos;
const sessions = {table:null, rack:null};
const images = {table:null, rack:null};
const recognizers = {table:null, rack:null};
// Developer-only local diagnostics: raw observations, never photos or corrections.
window.rummyObservationReport = () => Object.fromEntries(['table','rack'].filter(zone => sessions[zone]).map(zone => [zone, {
  source:sessions[zone].source, zone, recognizer:recognizers[zone],
  observations:sessions[zone].observations.map(o => JSON.parse(JSON.stringify(o.raw)))
}]));
let currentJob = null;
let nextJob = 0;
let revision = 0;
const $ = id => document.getElementById(id);
function status(text, error=false) { $('status').textContent=text; $('status').classList.toggle('error',error); }
function invalidate() { revision++; $('result').replaceChildren(); if(currentJob) rummyCancel(currentJob); }
function node(tag, text, className) { const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(className)e.className=className;return e; }
function settings() {
  if (!$('edition').value) throw new Error('Choose the 4-player or 6-player tile pool');
  return {edition:Number($('edition').value),opening_required:$('opening').checked};
}
async function photo(zone,camera=false) {
  if(window.webkit?.messageHandlers?.photo){window.webkit.messageHandlers.photo.postMessage({zone,camera});return;}
  if(camera) { status('Camera recognition is available in the iPhone app. Choose a local photo here.',true);return; }
  $(zone+'-file').click();
}
document.querySelectorAll('[data-photo]').forEach(b=>b.onclick=()=>photo(b.dataset.photo));
document.querySelectorAll('[data-camera]').forEach(b=>b.onclick=()=>photo(b.dataset.camera,true));
for(const zone of ['table','rack']){
  $(zone+'-file').onchange=async event=>{
    try{
      const file=event.target.files[0];if(!file)return;
      const bytes=await file.arrayBuffer();
      const hash=[...new Uint8Array(await crypto.subtle.digest('SHA-256',bytes))].map(n=>n.toString(16).padStart(2,'0')).join('');
      const image=await new Promise((resolve,reject)=>{const r=new FileReader();r.onload=()=>resolve(r.result);r.onerror=reject;r.readAsDataURL(file)});
      await acceptNativePhoto({zone,source:hash,image,observations:[],notice:'Browser import has no recognizer. Tap every tile in the photo and enter its value; iPhone recognition uses local Apple Vision.'});
    }catch(error){status(String(error),true);}
  };
}
window.acceptNativePhoto=async payload=>{
  try{
    if(payload.error)throw new Error(payload.error);
    const {zone,source}=payload;
    if(!['table','rack'].includes(zone))throw new Error('Invalid photo zone');
    const other=zone==='table'?'rack':'table';
    if(sessions[other]?.source===source)throw new Error('Use separate table and rack photos');
    const image=new Image();image.src=payload.image;await image.decode();
    recognizers[zone]=payload.recognizer||null;
    sessions[zone]=new PhotoSession(zone,source,payload.observations||[]);images[zone]=image;
    invalidate();drawPhoto(zone);
    status(payload.notice||'Review every tile, its color, and the table meld boundaries. Tap the photo to add missed tiles.');
  }catch(error){status(String(error),true);}
};
function drawPhoto(zone){
  const session=sessions[zone], image=images[zone], parent=$(zone+'-photo');parent.replaceChildren();
  const preview=image.cloneNode();preview.className='photo-preview';preview.alt=zone+' photo. Tap to add a missed tile.';
  preview.onclick=event=>{
    const rect=preview.getBoundingClientRect();const x=(event.clientX-rect.left)/rect.width,y=(event.clientY-rect.top)/rect.height;
    session.add({x:Math.max(0,Math.min(.9,x-.05)),y:Math.max(0,Math.min(.86,y-.07)),width:.1,height:.14});invalidate();drawPhoto(zone);
  };
  parent.append(preview,node('p','Tap a missed tile to add it. Exclude false detections. Confirm each value; for table tiles, assign the meld number and keep their displayed order.','notice'));
  for(const observation of session.observations){
    const row=node('div',undefined,'tile-editor');
    const crop=node('canvas');crop.width=120;crop.height=150;
    const rect=observation.raw.rect||{x:0,y:0,width:1,height:1};
    const ctx=crop.getContext('2d');ctx.drawImage(image,rect.x*image.naturalWidth,rect.y*image.naturalHeight,rect.width*image.naturalWidth,rect.height*image.naturalHeight,0,0,120,150);
    const detail=node('div');
    const guesses=(observation.raw.candidates||[]).map(c=>`${c.tile.joker?'joker':c.tile.color+' '+c.tile.number} (${Math.round(c.confidence*100)}%)`).join(', ');
    detail.append(node('p',`Read: ${observation.raw.text||'uncertain'} · ${guesses||'no proposed value'}`,'raw-reading'));
    const fields=node('div',undefined,'tile-fields');
    const suggested=observation.correction?.tile||observation.raw.candidates?.[0]?.tile||{};
    const color=node('select');color.setAttribute('aria-label','Tile color');
    for(const c of ['', 'black','blue','orange','red','joker']){const opt=node('option',c||'Choose');opt.value=c;color.append(opt)}
    color.value=suggested.joker?'joker':(suggested.color||'');
    const number=node('input');number.type='number';number.min='1';number.max='13';number.value=suggested.number||'';number.setAttribute('aria-label','Tile number');
    const meld=node('input');meld.type='number';meld.min='1';meld.value=observation.correction?.meld||observation.raw.suggestedMeld||'';meld.setAttribute('aria-label','Table meld number');
    const order=node('input');order.type='number';order.min='1';order.value=observation.correction?.order||observation.raw.suggestedOrder||session.observations.indexOf(observation)+1;order.setAttribute('aria-label','Tile order within meld');
    for(const [label,control] of [['Color',color],['Number',number],...(zone==='table'?[['Meld',meld],['Order',order]]:[])]){const l=node('label',label);l.append(control);fields.append(l)}
    const save=node('button',observation.correction?'Confirmed ✓':'Confirm tile');
    save.onclick=()=>{try{
      session.correct(observation.id,color.value==='joker'?{joker:true}:{color:color.value,number:Number(number.value)},Number(meld.value),Number(order.value));
      invalidate();drawPhoto(zone);
    }catch(error){status(String(error),true)}};
    // Edited controls invalidate their previous confirmation immediately.
    for(const control of [color,number,meld,order])control.oninput=()=>{observation.correction=null;session.confirmed=false;invalidate();save.textContent='Confirm tile';save.classList.remove('confirmed')};
    if(observation.correction)save.classList.add('confirmed');
    const exclude=node('button',observation.excluded?'Restore detection':'Exclude detection','secondary');
    exclude.onclick=()=>{session.exclude(observation.id,!observation.excluded);invalidate();drawPhoto(zone)};
    fields.append(save,exclude);detail.append(fields);row.append(crop,detail);row.style.opacity=observation.excluded?'.45':'1';parent.append(row);
  }
  const confirm=node('button',session.confirmed?'Photo confirmed ✓':'Confirm complete '+zone+' reading', 'confirm-photo');
  confirm.onclick=()=>{try{session.confirm(true);invalidate();drawPhoto(zone);status('Confirmed '+zone+' reading. Check that every physical tile is included exactly once.')}catch(error){status(String(error),true)}};
  parent.append(node('p','Before confirming: check that all tiles are visible, no tile was missed, and table and rack photos do not overlap. An empty table is allowed.','notice'),confirm);
}
async function execute(operation){
  if(currentJob)return;
  try{
    const chosen=settings();const position=positionFromPhotos(sessions.table,sessions.rack,operation==='solve');
    await engineReady;const id=String(++nextJob);currentJob=id;const startRevision=revision;
    $('cancel').disabled=false;$('solve').disabled=true;$('validate').disabled=true;status(operation==='solve'?'Searching locally…':'Checking table locally…');
    const response=JSON.parse(await rummyExecute(id,JSON.stringify({operation,settings:chosen,turn:{opened:$('opened').checked,drawn:$('drawn').checked},position,max_nodes:Number($('budget').value)})));
    if(startRevision!==revision){status('Input changed; discarded the earlier result.');return;}
    if(response.error){const issues=response.position?.issues?.map(i=>i.message).join('; ');throw new Error(issues||response.error)}
    const messages={optimal:'Verified move · maximum rack tiles played',feasible_incomplete:'Verified move · search incomplete; a better move may exist',no_move:'Complete search · no legal move on this turn',incomplete:'Search incomplete · no verified move found',invalid_input:'Invalid input'};
    if(response.search){status(messages[response.search.status]+(response.search.reason?` (${response.search.reason})`:''))}else{status('Confirmed table is valid.')}
    $('result').replaceChildren();
    if(response.svg){$('result').append(node('p','Scroll sideways to see the full table and rack.','notice'));const parsed=new DOMParser().parseFromString(response.svg,'image/svg+xml');$('result').append(document.importNode(parsed.documentElement,true));}
  }catch(error){status(String(error),true)}finally{currentJob=null;$('cancel').disabled=true;$('solve').disabled=false;$('validate').disabled=false;}
}
$('solve').onclick=()=>execute('solve');$('validate').onclick=()=>execute('validate');$('cancel').onclick=()=>{if(currentJob){rummyCancel(currentJob);status('Cancelling; keeping the best verified move found so far…')}};
for(const id of ['edition','opening','opened','drawn','budget'])$(id).onchange=invalidate;
engineReady.then(()=>status('Local engine ready. Choose an edition, then import the table photo.')).catch(error=>status(String(error),true));
