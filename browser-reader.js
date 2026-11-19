/* Local tile proposals. No fetch, worker, remote inference, or game-rule guessing.
 * Unreadable proposals stay empty; users can edit values and meld boundaries.
 */
(function (root) {
  'use strict';
  const MAX_DIMENSION = 1400;
  const pause = () => new Promise(resolve => setTimeout(resolve, 0));
  let initialization;

  function colorOf(r, g, b) {
    const high = Math.max(r, g, b), low = Math.min(r, g, b), delta = high - low;
    const saturation = high ? delta/high : 0;
    if (!delta) return high < 125 ? 'black' : null;
    let hue;
    if (high === r) hue = ((g - b) / delta) % 6;
    else if (high === g) hue = (b - r) / delta + 2;
    else hue = (r - g) / delta + 4;
    hue = (hue * 60 + 360) % 360;
    // Dim teal retains a strong hue even below the absolute black-ink cutoff.
    // Keep the established warm/neutral black rule: warm lighting can tint a
    // printed black joker brown without changing its physical ink color.
    if (hue >= 145 && hue <= 265 && high >= 45 && delta >= 20 && saturation >= .35) return 'blue';
    if (high < 125 && delta < 65) return 'black';
    if (delta < 65 || high < 95) return null;
    if (hue < 13 || hue > 335) return 'red';
    if (hue >= 13 && hue < 65) return 'orange';
    // The blue ink on some physical sets is cyan/teal. This is only a proposal.
    if (hue >= 145 && hue <= 265) return 'blue';
    return null;
  }

  function components(mask, width, height, minimum) {
    const visited = new Uint8Array(mask.length), queue = new Int32Array(mask.length), found = [];
    for (let start = 0; start < mask.length; start++) {
      if (!mask[start] || visited[start]) continue;
      let head = 0, tail = 1, x0 = width, y0 = height, x1 = 0, y1 = 0;
      queue[0] = start; visited[start] = 1;
      while (head < tail) {
        const index = queue[head++], x = index % width, y = Math.floor(index / width);
        x0 = Math.min(x0, x); x1 = Math.max(x1, x); y0 = Math.min(y0, y); y1 = Math.max(y1, y);
        const neighbors = [x > 0 ? index - 1 : -1, x + 1 < width ? index + 1 : -1,
          y > 0 ? index - width : -1, y + 1 < height ? index + width : -1];
        for (const next of neighbors) if (next >= 0 && mask[next] && !visited[next]) {
          visited[next] = 1; queue[tail++] = next;
        }
      }
      if (tail >= minimum) found.push({x:x0, y:y0, width:x1-x0+1, height:y1-y0+1, area:tail});
    }
    return found;
  }

  function percentile(histogram, count, fraction) {
    let remaining = Math.ceil(count*fraction);
    for (let value=0; value<256; value++) {
      remaining -= histogram[value];
      if (remaining <= 0) return value;
    }
    return 255;
  }

  function faceInk(canvas, cv) {
    const context=canvas.getContext('2d',{willReadFrequently:true});
    const image=context.getImageData(0,0,canvas.width,canvas.height);
    const {width,height,data}=image;
    const histograms=[new Uint32Array(256),new Uint32Array(256),new Uint32Array(256)];
    const margin=Math.max(2,Math.round(Math.min(width,height)*.07));
    let samples=0;
    for(let y=margin;y<height-margin;y++)for(let x=margin;x<width-margin;x++){
      const i=(y*width+x)*4;
      for(let channel=0;channel<3;channel++)histograms[channel][data[i+channel]]++;
      samples++;
    }
    const background=histograms.map(h=>Math.max(1,percentile(h,samples,.75)));
    const mask=new Uint8Array(width*height),labels={black:1,blue:2,orange:3,red:4};
    let gray,threshold,foreground;
    const normalized=new Float32Array(width*height*3);
    try {
    gray=new cv.Mat(height,width,cv.CV_8UC1);threshold=new cv.Mat();
    const grayValues=gray.data;
    for(let y=0;y<height;y++)for(let x=0;x<width;x++){
      const i=y*width+x;
      for(let c=0;c<3;c++)normalized[i*3+c]=Math.min(255,data[i*4+c]*230/background[c]);
      grayValues[i]=Math.min(normalized[i*3],normalized[i*3+1],normalized[i*3+2]);
    }
    cv.adaptiveThreshold(gray,threshold,255,cv.ADAPTIVE_THRESH_GAUSSIAN_C,cv.THRESH_BINARY_INV,35,10);
    foreground=threshold.data.slice();
    } finally {threshold?.delete();gray?.delete();}
    for(let y=margin;y<height-margin;y++)for(let x=margin;x<width-margin;x++){
      const i=y*width+x;
      if(!foreground[i])continue;
      const r=normalized[i*3],g=normalized[i*3+1],b=normalized[i*3+2];
      const high=Math.max(r,g,b),low=Math.min(r,g,b);
      const contrast=230-low;
      if(contrast<25)continue;
      const delta=high-low;
      const color=delta/contrast<.35?'black':colorOf(
        (r-low)*255/delta,(g-low)*255/delta,(b-low)*255/delta);
      if(color)mask[y*width+x]=labels[color];
    }
    const parts=components(mask,width,height,Math.max(3,width*height*.00025));
    const inside=parts.filter(p=>{
      const touchesLeft=p.x<=margin+1,touchesRight=p.x+p.width>=width-margin-1;
      const touchesTop=p.y<=margin+1,touchesBottom=p.y+p.height>=height-margin-1;
      return p.width/p.height>.1 && p.height/p.width>.1 && !(touchesLeft&&touchesRight || touchesTop&&touchesBottom ||
        p.width>width*.82 || p.height>height*.82);
    });
    // Boundary shadows and rack rails are not printed numerals. Work only with
    // surviving components before selecting a color or a number's orientation.
    const allowed=new Uint8Array(mask.length);
    for(const part of inside)for(let y=part.y;y<part.y+part.height;y++)for(let x=part.x;x<part.x+part.width;x++){
      const i=y*width+x;if(mask[i])allowed[i]=mask[i];
    }
    return {mask:allowed,parts:inside,width,height,labels,background};
  }

  function illuminationInk(raw, width, height, cv, size) {
    let gray, closed, kernel, difference;
    try {
      gray=new cv.Mat(height,width,cv.CV_8UC1);closed=new cv.Mat();difference=new cv.Mat();
      kernel=cv.getStructuringElement(cv.MORPH_ELLIPSE,new cv.Size(size,size));
      const pixels=gray.data;
      for(let i=0;i<width*height;i++)pixels[i]=Math.min(raw[i*4],raw[i*4+1],raw[i*4+2]);
      cv.morphologyEx(gray,closed,cv.MORPH_CLOSE,kernel);cv.subtract(closed,gray,difference);
      return difference.data.slice();
    } finally {difference?.delete();kernel?.delete();closed?.delete();gray?.delete();}
  }

  function turnCanvas(canvas) {
    const result=document.createElement('canvas');result.width=canvas.width;result.height=canvas.height;
    const context=result.getContext('2d');context.translate(result.width,result.height);context.rotate(Math.PI);context.drawImage(canvas,0,0);
    return result;
  }

  function glyph(canvas, cv) {
    let ink=faceInk(canvas, cv);
    if(!ink.parts.length)return null;
    const strongest=[...ink.parts].sort((a,b)=>b.area-a.area)[0];
    const main=ink.parts.filter(p=>p.height>=strongest.height*.4 && p.area>=strongest.area*.25);
    const mean=main.reduce((sum,p)=>sum+(p.y+p.height/2)*p.area,0)/main.reduce((sum,p)=>sum+p.area,0);
    let rotation=0;
    if(mean>ink.height*.53){canvas=turnCanvas(canvas);ink=faceInk(canvas, cv);rotation=180;}
    const {width,height,labels}=ink;
    let parts=ink.parts;
    if(!parts.length)return null;
    const tallest=Math.max(...parts.map(p=>p.height));
    const significant=parts.filter(p=>p.height>=tallest*.48);
    const counts={black:0,blue:0,orange:0,red:0},colorLabels=['','black','blue','orange','red'];
    for(const part of significant)for(let y=part.y;y<part.y+part.height;y++)for(let x=part.x;x<part.x+part.width;x++){
      const label=ink.mask[y*width+x];
      if(label)counts[colorLabels[label]]++;
    }
    const ranked=Object.entries(counts).sort((a,b)=>b[1]-a[1]);
    const color=ranked[0][1]>=12 && ranked[0][1]>ranked[1][1]*1.4?ranked[0][0]:null;
    const mask=Uint8Array.from(ink.mask,value=>color?Number(value===labels[color]):Number(value!==0));
    parts=components(mask,width,height,Math.max(3,width*height*.00025));
    if(!parts.length)return null;
    const largest=[...parts].sort((a,b)=>b.area-a.area)[0];
    const inner=parts.filter(p=>p!==largest && p.x>=largest.x && p.y>=largest.y &&
      p.x+p.width<=largest.x+largest.width && p.y+p.height<=largest.y+largest.height);
    const possibleJoker=largest.width/largest.height>.75 && largest.width/largest.height<1.35 &&
      inner.length>=2 && largest.area/(largest.width*largest.height)<.6;
    if(!possibleJoker)parts=parts.filter(p=>
      p.y+p.height>=largest.y-largest.height*.2 && p.y<=largest.y+largest.height*1.2 &&
      p.x+p.width>=largest.x-largest.height*.4 && p.x<=largest.x+largest.width+largest.height*.4);
    const clean=new Uint8Array(mask.length);
    for(const part of parts)for(let y=part.y;y<part.y+part.height;y++)for(let x=part.x;x<part.x+part.width;x++){
      const i=y*width+x;if(mask[i])clean[i]=1;
    }
    const left=Math.min(...parts.map(p=>p.x)),top=Math.min(...parts.map(p=>p.y));
    const right=Math.max(...parts.map(p=>p.x+p.width)),bottom=Math.max(...parts.map(p=>p.y+p.height));
    const raw=canvas.getContext('2d').getImageData(0,0,width,height).data;
    function crop(x0,y0,x1,y1,illumination){
      const output=document.createElement('canvas'),padding=20,scale=120/Math.max(1,y1-y0);
      const localHist=[new Uint32Array(256),new Uint32Array(256),new Uint32Array(256)];let count=0;
      for(let y=Math.max(0,y0-8);y<Math.min(height,y1+8);y++)for(let x=Math.max(0,x0-8);x<Math.min(width,x1+8);x++){
        const i=(y*width+x)*4;for(let c=0;c<3;c++)localHist[c][raw[i+c]]++;count++;
      }
      const localBackground=localHist.map(h=>Math.max(1,percentile(h,count,.9)));
      output.width=Math.ceil((x1-x0)*scale)+padding*2;output.height=160;
      const target=output.getContext('2d');target.fillStyle='#fff';target.fillRect(0,0,output.width,output.height);
      const source=document.createElement('canvas');source.width=x1-x0;source.height=y1-y0;
      const pixels=source.getContext('2d').createImageData(source.width,source.height);
      for(let y=y0;y<y1;y++)for(let x=x0;x<x1;x++){
        const i=((y-y0)*source.width+x-x0)*4,offset=(y*width+x)*4;
        const r=Math.min(255,raw[offset]*255/localBackground[0]),
          g=Math.min(255,raw[offset+1]*255/localBackground[1]),
          b=Math.min(255,raw[offset+2]*255/localBackground[2]);
        const strength=color==='red'?r-(g+b)/2:color==='orange'?r-b:Math.max(g,b)-r;
        const value=illumination?255-Math.min(255,illumination[y*width+x]*5):
          color==='black'?Math.min(r,g,b):Math.max(0,Math.min(255,255-strength*3));
        pixels.data[i]=pixels.data[i+1]=pixels.data[i+2]=value;pixels.data[i+3]=255;
      }
      source.getContext('2d').putImageData(pixels,0,0);target.drawImage(source,padding,padding,output.width-padding*2,120);
      return output;
    }
    const digits=[];
    for(const part of parts.filter(p=>p.height>=largest.height*.6).sort((a,b)=>a.x-b.x)){
      const previous=digits[digits.length-1];
      if(previous && part.x<=previous.right+1)previous.right=Math.max(previous.right,part.x+part.width);
      else digits.push({left:part.x,right:part.x+part.width});
    }
    const digitCount=digits.length===2?2:(digits.length===1 && (right-left)/(bottom-top)<.85?1:0);
    return {canvas:crop(left,top,right,bottom),color,possibleJoker,digitCount,rotation,
      digits:digitCount===2?digits.map(p=>crop(p.left,top,p.right,bottom)):[],
      alternative:size=>crop(left,top,right,bottom,illuminationInk(raw,width,height,cv,size))};
  }

  function recognize(module,api,canvas,modes) {
    const encoded=canvas.toDataURL('image/png').split(',')[1];
    module.FS.writeFile('/input',Uint8Array.from(atob(encoded),c=>c.charCodeAt(0)));
    try {
      if(api.SetImageFile(1,0)!==0)throw new Error('The number crop could not be decoded.');
      return modes.map(mode=>{
        api.SetVariable('tessedit_pageseg_mode',mode);api.Recognize(null);
        return {mode,text:api.GetUTF8Text().trim(),confidence:api.MeanTextConf()/100};
      });
    } finally {module.FS.unlink('/input');}
  }

  function selectNumber(readings,digitCount,character=false) {
    const pattern=character?/^[0-9]$/:/^(?:[1-9]|1[0-3])$/;
    const valid=readings.filter(r=>pattern.test(r.text) && r.confidence>=.3 &&
      (!digitCount || r.text.length===digitCount));
    const counts=new Map();
    for(const r of valid)counts.set(r.text,(counts.get(r.text)||0)+1);
    const confidence=text=>Math.max(...valid.filter(r=>r.text===text).map(r=>r.confidence));
    const ranked=[...counts].sort((a,b)=>confidence(b[0])-confidence(a[0]) || b[1]-a[1]);
    if(!ranked.length || (ranked.length>1 && Math.abs(confidence(ranked[0][0])-confidence(ranked[1][0]))<.1))return null;
    const number=ranked[0][0];
    return {text:number,confidence:Math.max(...valid.filter(r=>r.text===number).map(r=>r.confidence))};
  }

  function ready() {
    if (!initialization) initialization = (async () => {
      if (!root.TesseractCore || !root.RUMMY_READER_MODEL || !root.RUMMY_READER_WASM) throw new Error('Local reader assets are not bundled.');
      const binary=Uint8Array.from(atob(root.RUMMY_READER_WASM),c=>c.charCodeAt(0));
      delete root.RUMMY_READER_WASM;
      const module = await root.TesseractCore({wasmBinary:binary,print:()=>{},printErr:()=>{},TesseractProgress:()=>{}});
      module.FS.writeFile('/eng.traineddata',Uint8Array.from(atob(root.RUMMY_READER_MODEL),c=>c.charCodeAt(0)));
      delete root.RUMMY_READER_MODEL;
      const api = new module.TessBaseAPI();
      if (api.Init(null,'eng',1,undefined) !== 0) throw new Error('The bundled number reader could not start.');
      api.SetVariable('tessedit_char_whitelist','0123456789');
      api.SetVariable('tessedit_pageseg_mode','7');
      return {module,api};
    })();
    return initialization;
  }

  let visionInitialization;
  function visionReady() {
    if(!visionInitialization)visionInitialization=new Promise((resolve,reject)=>{
      function initialize(){
        try {
        if(typeof root.RummyVisionRuntime!=='function'){
          reject(new Error('The bundled tile finder could not start.'));return;
        }
        const cv=root.RummyVisionRuntime({
          print:()=>{},printErr:()=>{},onAbort:()=>reject(new Error('The bundled tile finder stopped.'))});
        cv.then(()=>resolve({cv}));
        } catch(error) {reject(error);}
      }
      if(root.RummyVisionRuntime){initialize();return;}
      const script=document.createElement('script');
      const version=document.querySelector('meta[name="rummy-preview-source"]')?.content;
      script.src='recognition/vision-core.js'+(version?'?v='+encodeURIComponent(version):'');
      script.onerror=()=>reject(new Error('The bundled tile finder could not load.'));
      script.onload=initialize;
      document.head.append(script);
    });
    return visionInitialization;
  }

  function groupFaces(boxes, imageWidth, zone) {
    const length=(a,b)=>Math.hypot(a.x-b.x,a.y-b.y);
    const faces=boxes.map(box=>{
      const q=box.quad;
      const width=(length(q[0],q[1])+length(q[3],q[2]))/2;
      const height=(length(q[0],q[3])+length(q[1],q[2]))/2;
      const turned=width>height;
      const a=turned?q[1]:q[0],b=turned?q[2]:q[1];
      const size=Math.hypot(b.x-a.x,b.y-a.y);
      return {...box,cx:q.reduce((s,p)=>s+p.x,0)/4,cy:q.reduce((s,p)=>s+p.y,0)/4,
        faceWidth:Math.min(width,height),faceHeight:Math.max(width,height),turned,
        ax:(b.x-a.x)/size,ay:(b.y-a.y)/size};
    });
    const vertical=faces.filter(p=>p.turned).length>faces.length/2;
    for(const face of faces)if(vertical?face.ay<0:face.ax<0){face.ax=-face.ax;face.ay=-face.ay;}
    const parent=faces.map((_,i)=>i);
    const find=i=>parent[i]===i?i:(parent[i]=find(parent[i]));
    for(let i=0;i<faces.length;i++)for(let j=i+1;j<faces.length;j++){
      const a=faces[i],b=faces[j];
      if(a.ax*b.ax+a.ay*b.ay<.7)continue;
      let ax=a.ax+b.ax,ay=a.ay+b.ay;const norm=Math.hypot(ax,ay);ax/=norm;ay/=norm;
      const dx=b.cx-a.cx,dy=b.cy-a.cy;
      const along=Math.abs(dx*ax+dy*ay),across=Math.abs(dx*ay-dy*ax);
      const gap=along-(a.faceWidth+b.faceWidth)/2;
      if(along>Math.min(a.faceWidth,b.faceWidth)*.4 &&
        gap>=-Math.min(a.faceWidth,b.faceWidth)*.4 && gap<=Math.min(a.faceWidth,b.faceWidth)*.4 &&
        across<Math.min(a.faceHeight,b.faceHeight)*.37)parent[find(j)]=find(i);
    }
    const groups=new Map();
    faces.forEach((face,i)=>{const key=find(i);if(!groups.has(key))groups.set(key,[]);groups.get(key).push(face);});
    const rows=[...groups.values()].sort((a,b)=>{
      const position=g=>g.reduce((s,p)=>s+(vertical?imageWidth-p.cx:p.cy),0)/g.length;
      return position(a)-position(b);
    });
    const output=[];
    for(const [index,row] of rows.entries()){
      row.sort((a,b)=>vertical?a.cy-b.cy:a.cx-b.cx);
      for(const [order,box] of row.entries())output.push({...box,
        suggestedMeld:zone==='rack'?1:index+1,suggestedOrder:zone==='rack'?output.length+1:order+1});
    }
    return output;
  }

  async function read(image, {zone='table',progress=()=>{},cancelled=()=>false}={}) {
    const scale = Math.min(1,MAX_DIMENSION/Math.max(image.naturalWidth,image.naturalHeight));
    const canvas = document.createElement('canvas');
    canvas.width = Math.max(1,Math.round(image.naturalWidth*scale));canvas.height = Math.max(1,Math.round(image.naturalHeight*scale));
    const context = canvas.getContext('2d',{willReadFrequently:true});context.drawImage(image,0,0,canvas.width,canvas.height);
    const pixels = context.getImageData(0,0,canvas.width,canvas.height);
    progress('Looking for tiles on this device…');await pause();
    const {cv}=await visionReady();
    if(cancelled())throw new Error('Photo reading was replaced.');
    const geometry=await root.RummyTileGeometry.detect(cv,pixels,{zone,cancelled,progress});
    if(geometry.boxes.length>160)throw new Error('Too many tile faces were detected. Crop the photo and try again.');
    const boxes=groupFaces(geometry.boxes,pixels.width,zone);
    const {module,api} = boxes.length ? await ready() : {};
    const detail=document.createElement('canvas');
    const detailScale=Math.min(1,2800/Math.max(image.naturalWidth,image.naturalHeight));
    detail.width=Math.max(1,Math.round(image.naturalWidth*detailScale));
    detail.height=Math.max(1,Math.round(image.naturalHeight*detailScale));
    const detailContext=detail.getContext('2d',{willReadFrequently:true});
    detailContext.drawImage(image,0,0,detail.width,detail.height);
    const observations = [],source=cv.matFromImageData(detailContext.getImageData(0,0,detail.width,detail.height));
    try {
    for (let i = 0; i < boxes.length; i++) {
      if (cancelled()) throw new Error('Photo reading was replaced.');
      const box=boxes[i];
      let quad=box.quad;
      const length=(a,b)=>Math.hypot(a.x-b.x,a.y-b.y);
      if(zone!=='rack' && length(quad[0],quad[1])+length(quad[3],quad[2])>length(quad[0],quad[3])+length(quad[1],quad[2]))quad=[quad[1],quad[2],quad[3],quad[0]];
      quad=quad.map(p=>({x:p.x*detail.width/canvas.width,y:p.y*detail.height/canvas.height}));
      const face=root.RummyTileGeometry.warp(cv,source,quad,180,264);
      const ink=glyph(face,cv);
      let text='', candidates=[], ocrReadings=[], digitReadings=[], illuminationReadings=[], digitCount=ink?.digitCount||0;
      if (!ink) {text='Unreadable tile face';}
      else if (ink.possibleJoker) {text='Possible joker face';candidates=[{tile:{joker:true},confidence:.5}];}
      else if (!digitCount) {text='Unclear printed digits';}
      else {
        ocrReadings=recognize(module,api,ink.canvas,digitCount===1?['7','13','10']:['7','13']);
        let chosen=selectNumber(ocrReadings,digitCount);
        if (!chosen && digitCount===2) {
          digitReadings=ink.digits.map(canvas=>recognize(module,api,canvas,['10','13']));
          const digits=digitReadings.map(readings=>selectNumber(readings,1,true));
          if(digits.every(Boolean))chosen=selectNumber([{text:digits.map(d=>d.text).join(''),
            confidence:Math.min(...digits.map(d=>d.confidence))}],2);
        } else if (!chosen) {
          ocrReadings.push(...recognize(module,api,ink.canvas,['8']));
          chosen=selectNumber(ocrReadings,digitCount);
        }
        if(!chosen){
          // Uneven reflections need a local illumination estimate. Accept only
          // agreement across distinct windows; unresolved or conflicting crops
          // remain errors for editing rather than guesses from game legality.
          const votes=[];
          for(const size of [31,51,71]){
            const readings=recognize(module,api,ink.alternative(size),digitCount===1?['7','13','10']:['7','13']);
            const candidate=selectNumber(readings,digitCount);
            illuminationReadings.push({size,readings});
            if(candidate && candidate.confidence>=.5)votes.push(candidate);
          }
          if(votes.length>=2 && votes.every(v=>v.text===votes[0].text))chosen={text:votes[0].text,confidence:Math.min(...votes.map(v=>v.confidence))};
        }
        text=[...new Set(ocrReadings.map(r=>r.text).filter(Boolean))].join(' / ');
        if(chosen && ink.color) {
          candidates=[{tile:{color:ink.color,number:Number(chosen.text)},confidence:chosen.confidence}];
        } else if(text) {
          text='Uncertain or partial OCR: '+text;
        }
      }
      observations.push({rect:{x:box.x/canvas.width,y:box.y/canvas.height,width:box.width/canvas.width,height:box.height/canvas.height},
        text:text||'Uncertain number or color',candidates,ocrReadings,digitReadings,illuminationReadings,digitCount,rotation:ink?.rotation||0,sideways:box.turned,suggestedMeld:box.suggestedMeld,suggestedOrder:box.suggestedOrder});
      progress(`Reading tiles on this device… ${i+1} of ${boxes.length}`);await pause();
    }
    } finally {source.delete();}
    for(const meld of new Set(observations.map(o=>o.suggestedMeld))){
      const row=observations.filter(o=>o.suggestedMeld===meld);
      if(zone==='table' && row.every(o=>o.sideways) && row.filter(o=>o.rotation===180).length>row.length/2){
        row.reverse();row.forEach((o,i)=>{o.suggestedOrder=i+1;});
      }
    }
    observations.sort((a,b)=>a.suggestedMeld-b.suggestedMeld || a.suggestedOrder-b.suggestedOrder);
    return {observations,recognizer:{name:'local-ivory-tiles-tesseract',pipelineRevision:5,coreVersion:'6.1.2',visionVersion:'4.12.0-release.1',
      modelRevision:'87416418657359cb625c412a48b6e1d6d41c29bd',maxDimension:MAX_DIMENSION,detailDimension:2800,
      geometry:geometry.diagnostics},
      notice:`Proposed ${observations.length} tiles locally. Scores are uncalibrated; unreadable values require editing.`};
  }
  const api={read,ready,colorOf};
  root.RummyReader=api;
  if (typeof module!=='undefined' && module.exports) module.exports=api;
})(globalThis);
