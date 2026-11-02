'use strict';
const {PhotoSession, positionFromPhotos, validTile} = RummyPhotos;
const sessions = {table:null, rack:null};
const images = {table:null, rack:null};
const recognizers = {table:null, rack:null};
const photoRequests = {table:0,rack:0};
const $ = id => document.getElementById(id);
const SEARCH_NODES = 200000;
let currentJob=null, nextJob=0, revision=0, lastOutcome=null, pendingPhotoZone=null;
// Raw reader output stays separate from edits and automatic use of proposals.
window.rummyObservationReport = () => Object.fromEntries(['table','rack'].filter(zone => sessions[zone]).map(zone => [zone, {
  source:sessions[zone].source, zone, recognizer:recognizers[zone],
  observations:sessions[zone].observations.filter(o => !o.manual).map(o => JSON.parse(JSON.stringify(o.raw)))
}]));
$('licenses').textContent=(globalThis.RUMMY_GO_LICENSE||'License text is missing from this build.')+
  (globalThis.RUMMY_READER_LICENSES?'\n\n'+globalThis.RUMMY_READER_LICENSES:'');
if(window.webkit?.messageHandlers?.photo)$('reading-help').hidden=true;
$('save-report').onclick=()=>{
  const photos=rummyObservationReport();
  for(const zone of Object.keys(photos))photos[zone].image=images[zone].src;
  if(!Object.keys(photos).length){status('Choose a photo before saving a reading report.',true);return;}
  const report={version:1,kind:'rummy-photo-reading-report',
    preview_source:document.querySelector('meta[name="rummy-preview-source"]')?.content||null,
    browser:navigator.userAgent,photos,
    state:{settings:settings(),turn:{opened:$('opened').checked},last_outcome:lastOutcome}};
  const url=URL.createObjectURL(new Blob([JSON.stringify(report)],{type:'application/json'}));
  const link=node('a');link.href=url;link.download='rummy-reading-report.json';
  document.body.append(link);link.click();link.remove();setTimeout(()=>URL.revokeObjectURL(url),60000);
  status('Reading report saved on your device.');
};
function status(text,error=false){$('status').textContent=text;$('status').classList.toggle('error',error);}
function invalidate(){revision++;lastOutcome=null;$('result').replaceChildren();if(currentJob&&typeof rummyCancel==='function')rummyCancel(currentJob);}
function node(tag,text,className){const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(className)e.className=className;return e;}
function settings(){return {edition:Number($('edition').value),opening_required:$('opening').checked};}
function readingStatus(){
  let count=0;
  for(const zone of ['table','rack'])if(sessions[zone]){
    try{count+=sessions[zone].pieces().length;}catch(error){status(error.message,true);return;}
  }
  status(sessions.table&&sessions.rack?`${count} tiles ready. Find your move.`:sessions.table?'Table ready. Add your rack.':sessions.rack?'Rack ready. Add the table.':'Add your table and rack photos.');
}
async function photo(zone,camera=false){
  if(window.webkit?.messageHandlers?.photo){
    pendingPhotoZone=zone;window.webkit.messageHandlers.photo.postMessage({zone,camera});return;
  }
  $(zone+'-file').click();
}
document.querySelectorAll('[data-photo]').forEach(b=>b.onclick=()=>photo(b.dataset.photo));
document.querySelectorAll('[data-camera]').forEach(b=>b.onclick=()=>photo(b.dataset.camera,true));
for(const zone of ['table','rack'])$(zone+'-file').onchange=async event=>{
  let request;
  const cancelled=()=>request!==undefined&&photoRequests[zone]!==request;
  try{
    const file=event.target.files[0];if(!file)return;event.target.value='';
    request=++photoRequests[zone];
    sessions[zone]=null;images[zone]=null;invalidate();
    $('save-report').disabled=!sessions.table&&!sessions.rack;
    $(zone+'-photo').replaceChildren(node('p','Reading photo…','notice'));
    const bytes=await file.arrayBuffer();
    const hash=[...new Uint8Array(await crypto.subtle.digest('SHA-256',bytes))].map(n=>n.toString(16).padStart(2,'0')).join('');
    const image=await new Promise((resolve,reject)=>{const r=new FileReader();r.onload=()=>resolve(r.result);r.onerror=reject;r.readAsDataURL(file)});
    if(cancelled())return;
    checkPhotoSource(zone,hash);
    const decoded=new Image();decoded.src=image;await decoded.decode();
    let result;
    try{result=await RummyReader.read(decoded,{cancelled,progress:message=>{if(!cancelled())status(message)}});}
    catch(error){if(cancelled())return;result={observations:[],readError:String(error?.message||error||'Try another photo.')};}
    if(!cancelled())installPhoto({zone,source:hash,...result},decoded);
  }catch(error){if(cancelled())return;$(zone+'-photo').replaceChildren(node('p',error.message,'reading-error'));status(error.message,true);}
};
function checkPhotoSource(zone, source){
  if(!['table','rack'].includes(zone))throw new Error('Invalid photo zone.');
  if(sessions[zone==='table'?'rack':'table']?.source===source)throw new Error('Use separate table and rack photos.');
}
function installPhoto(payload, image){
  const {zone,source}=payload;
  checkPhotoSource(zone,source);
  recognizers[zone]=payload.recognizer||null;
  sessions[zone]=new PhotoSession(zone,source,payload.observations||[],{error:payload.readError||''});
  images[zone]=image;
  $('save-report').disabled=false;
  invalidate();
  drawPhoto(zone);
  readingStatus();
}
window.acceptNativePhoto=async(payload,isCurrent=()=>true)=>{
  try{
    if(!isCurrent())return;
    if(payload.error){
      const zone=payload.zone||pendingPhotoZone;
      if(sessions[zone]!==undefined){sessions[zone]=null;images[zone]=null;invalidate();}
      throw new Error(payload.error);
    }
    const {zone,source}=payload;
    checkPhotoSource(zone,source);
    const image=new Image();image.src=payload.image;await image.decode();
    if(!isCurrent())return;
    installPhoto(payload,image);
  }catch(error){if(isCurrent())status(error.message,true);}
};
function drawPhoto(zone){
  const session=sessions[zone],image=images[zone],parent=$(zone+'-photo');parent.replaceChildren();
  const preview=image.cloneNode();preview.className='photo-preview';preview.alt=zone+' photo';
  if(session.readError){parent.append(preview,node('p','Could not read this photo. Try another photo.','reading-error'));return;}
  preview.onclick=event=>{
    const bounds=preview.getBoundingClientRect(),x=(event.clientX-bounds.left)/bounds.width,y=(event.clientY-bounds.top)/bounds.height;
    session.add({x:Math.max(0,Math.min(.9,x-.05)),y:Math.max(0,Math.min(.86,y-.07)),width:.1,height:.14});
    invalidate();drawPhoto(zone);readingStatus();
  };
  const active=session.observations.filter(o=>!o.excluded);
  const summary=node('div',undefined,'photo-summary');summary.append(node('span',`${active.length} tiles`));
  const add=node('button','+ Add tile','add-tile secondary');add.onclick=()=>{
    session.add({x:.45,y:.43,width:.1,height:.14});invalidate();drawPhoto(zone);readingStatus();
  };summary.append(add);parent.append(preview,summary);
  const list=node('div',undefined,'tile-list');parent.append(list);
  for(const [index,observation] of session.observations.entries()){
    if(observation.excluded)continue;
    const row=node('div',undefined,'tile-editor');
    const remove=node('button','×','remove-tile');remove.setAttribute('aria-label','Remove tile');
    remove.onclick=()=>{session.exclude(observation.id,true);invalidate();drawPhoto(zone);readingStatus();};
    const crop=node('canvas');crop.width=120;crop.height=150;
    const rect=observation.raw.rect||{x:0,y:0,width:1,height:1};
    crop.getContext('2d').drawImage(image,rect.x*image.naturalWidth,rect.y*image.naturalHeight,rect.width*image.naturalWidth,rect.height*image.naturalHeight,0,0,120,150);
    const reading=session.reading(observation.id);
    const fields=node('div',undefined,'tile-fields'),suggested=reading.tile||{};
    const color=node('select');color.setAttribute('aria-label','Tile color');
    for(const c of ['', 'black','blue','orange','red','joker']){const opt=node('option',c||'Color');opt.value=c;color.append(opt);}
    color.value=suggested.joker?'joker':suggested.color||'';
    const number=node('input');number.type='number';number.min='1';number.max='13';number.value=suggested.number||'';number.setAttribute('aria-label','Tile number');number.disabled=color.value==='joker';
    const meld=node('input');meld.type='number';meld.min='1';meld.value=reading.meld??'';meld.setAttribute('aria-label','Table meld number');
    const order=node('input');order.type='number';order.min='1';order.value=reading.order??index+1;order.setAttribute('aria-label','Tile order within meld');
    for(const [label,control] of [['Color',color],['Number',number],...(zone==='table'?[['Meld',meld],['Order',order]]:[])]){const l=node('label',label);l.append(control);fields.append(l);}
    const error=node('span','Could not read this tile.','tile-error');error.hidden=validTile(suggested);fields.append(error);
    for(const control of [color,number,meld,order])control.oninput=()=>{
      number.disabled=color.value==='joker';
      const tile=color.value==='joker'?{joker:true}:{color:color.value,number:Number(number.value)};
      session.edit(observation.id,tile,Number(meld.value),Number(order.value));
      error.textContent='Enter a color and number from 1 to 13.';error.hidden=validTile(tile);
      invalidate();readingStatus();
    };
    row.append(remove,crop,fields);list.append(row);
  }
}
async function execute(operation){
  if(currentJob)return;
  const chosen=settings(),turn={opened:$('opened').checked};
  try{
    const position=positionFromPhotos(sessions.table,sessions.rack,operation==='solve');
    const id=String(++nextJob),startRevision=revision;currentJob=id;
    $('cancel').disabled=false;$('solve').disabled=true;$('validate').disabled=true;status(operation==='solve'?'Finding your move…':'Checking table…');
    await engineReady;
    if(startRevision!==revision){readingStatus();return;}
    const response=JSON.parse(await rummyExecute(id,JSON.stringify({operation,settings:chosen,turn,position,max_nodes:SEARCH_NODES})));
    if(startRevision!==revision){readingStatus();return;}
    if(response.error)throw new Error(response.position?.issues?.map(i=>i.message).join('; ')||response.error);
    lastOutcome={operation,settings:chosen,turn,status:response.search?.status||'valid',reason:response.search?.reason||'',nodes:response.search?.nodes};
    $('result').replaceChildren();
    if(response.search){
      const search=response.search,reason=search.reason||'';
      const messages={optimal:'Verified move · maximum rack tiles played',feasible_incomplete:'Verified move · search incomplete; a better move may exist',no_move:'No legal move on this turn.',incomplete:'Search incomplete · no verified move found',invalid_input:'Check your tiles.'};
      let message=messages[search.status]||'Search result needs review.';
      if(search.verified)message+=` · ${search.verified.played.length} rack tiles`;
      if(search.status==='no_move'&&chosen.opening_required&&!turn.opened)message='No 30-point opening is possible from this rack. If you opened on a previous turn, select “I have opened”.';
      if(search.status==='incomplete'||search.status==='feasible_incomplete'){
        if(reason.includes('canceled'))message+='. Search cancelled.';
        if(reason.includes('node_limit'))message+='. Search limit reached.';
        if(reason.includes('table_joker_transition'))message+='. Some joker rearrangements remain unverified.';
        if(Number.isInteger(search.upper_bound))message+=` At most ${search.upper_bound} rack tiles could be played.`;
      }
      status(message);
      if(search.verified?.valid)$('result').append(RummyTableView.render(search.verified.after,search.verified.played,{label:'Final table'}));
    }else{
      status('Table is valid.');
      $('result').append(RummyTableView.render({table:response.position.resolved_table||[],rack:position.rack},[],{label:'Current table',showRack:false}));
    }
  }catch(error){lastOutcome={operation,settings:chosen,turn,error:error.message};status(error.message,true);}
  finally{currentJob=null;$('cancel').disabled=true;$('solve').disabled=false;$('validate').disabled=false;}
}
function openingControls(){
  $('opened').disabled=!$('opening').checked;
  if($('opened-option'))$('opened-option').hidden=!$('opening').checked;
}
$('solve').onclick=()=>execute('solve');$('validate').onclick=()=>execute('validate');
$('cancel').onclick=()=>{if(currentJob){if(typeof rummyCancel==='function')rummyCancel(currentJob);else revision++;status('Stopping search…');}};
for(const id of ['edition','opening','opened'])$(id).onchange=()=>{openingControls();invalidate();readingStatus();};
openingControls();engineReady.then(readingStatus).catch(error=>status(error.message,true));
