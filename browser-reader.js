/* Local tile proposals. No fetch, worker, remote inference, or game-rule guessing.
 * Every value, color, joker and meld boundary still requires user confirmation.
 */
(function (root) {
  'use strict';
  const MAX_DIMENSION = 1400, MAX_TILES = 160;
  const pause = () => new Promise(resolve => setTimeout(resolve, 0));
  let initialization;

  function colorOf(r, g, b) {
    const high = Math.max(r, g, b), low = Math.min(r, g, b), delta = high - low;
    if (high < 125 && delta < 65) return 'black';
    if (delta < 65 || high < 95) return null;
    let hue;
    if (high === r) hue = ((g - b) / delta) % 6;
    else if (high === g) hue = (b - r) / delta + 2;
    else hue = (r - g) / delta + 4;
    hue = (hue * 60 + 360) % 360;
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

  function detect(image) {
    const {data, width, height} = image, mask = new Uint8Array(width * height);
    for (let p = 0; p < mask.length; p++) {
      const r = data[p*4], g = data[p*4+1], b = data[p*4+2];
      mask[p] = (r+g+b)/3 > 165 && Math.max(r,g,b)-Math.min(r,g,b) < 75 && b > r*.7 ? 1 : 0;
    }
    const regions = components(mask, width, height, Math.max(80, width*height*.0006));
    const tiles = [];
    for (const box of regions) {
      const ratio = box.width / box.height, fill = box.area / (box.width*box.height);
      if (box.height < height*.04 || box.width < width*.018 || box.area > width*height*.65 ||
          ratio < .3 || ratio > 12 || fill < .4) continue;
      // Adjacent tiles can share one ivory component. Split a strip near tile
      // proportions, preferring dark seams over equal-width boundaries.
      const count = ratio > 1.12 ? Math.min(13, Math.round(ratio/.68)) : 1;
      const cuts = [box.x], step = box.width/count;
      for (let i = 1; i < count; i++) {
        const expected = box.x + i*step;
        let cut = Math.round(expected), best = Infinity;
        for (let x = Math.round(expected-step*.16); x <= expected+step*.16; x++) {
          let white = 0;
          for (let y = box.y; y < box.y+box.height; y++) white += mask[y*width+x] || 0;
          const score = white/box.height + Math.abs(x-expected)/step*.35;
          if (score < best) { best = score; cut = x; }
        }
        cuts.push(cut);
      }
      cuts.push(box.x+box.width);
      for (let i = 0; i < count; i++) tiles.push({x:cuts[i],y:box.y,width:cuts[i+1]-cuts[i],height:box.height});
    }
    if (tiles.length > MAX_TILES) throw new Error('Too many possible tiles. Crop closer to the table or rack, then try again.');
    return tiles;
  }

  function group(boxes) {
    const rows = [];
    for (const box of [...boxes].sort((a,b) => a.y-b.y)) {
      const row = rows.find(row => Math.abs(row[0].y+row[0].height/2-box.y-box.height/2) < Math.min(row[0].height,box.height)*.45);
      if (row) row.push(box); else rows.push([box]);
    }
    const output = []; let meld = 0;
    for (const row of rows) {
      let previous, order = 0;
      for (const box of row.sort((a,b) => a.x-b.x)) {
        if (!previous || box.x-previous.x-previous.width > Math.min(box.width,previous.width)*.55) { meld++; order = 0; }
        output.push({...box,suggestedMeld:meld,suggestedOrder:++order}); previous = box;
      }
    }
    return output;
  }

  function glyph(image, box) {
    const x0 = Math.round(box.x+box.width*.12), y0 = Math.round(box.y+box.height*.1);
    const width = Math.max(1,Math.round(box.width*.76)), height = Math.max(1,Math.round(box.height*.52));
    const mask = new Uint8Array(width*height), counts = {black:0,blue:0,orange:0,red:0};
    for (let y = 0; y < height; y++) for (let x = 0; x < width; x++) {
      const i = ((y+y0)*image.width+x+x0)*4;
      const color = colorOf(image.data[i],image.data[i+1],image.data[i+2]);
      if (color) {mask[y*width+x] = 1; counts[color]++;}
    }
    const parts = components(mask,width,height,Math.max(3,width*height*.001));
    if (!parts.length) return null;
    const sorted = Object.entries(counts).sort((a,b) => b[1]-a[1]);
    const color = sorted[0][1] >= 12 && sorted[0][1] > sorted[1][1]*1.4 ? sorted[0][0] : null;
    const largest = [...parts].sort((a,b) => b.area-a.area)[0];
    const aspect = largest.width/largest.height;
    const inside = parts.filter(p => p!==largest && p.x>=largest.x && p.y>=largest.y &&
      p.x+p.width<=largest.x+largest.width && p.y+p.height<=largest.y+largest.height);
    const possibleJoker = aspect > .8 && aspect < 1.25 && inside.length >= 2 &&
      largest.area/(largest.width*largest.height) < .55;
    const x = Math.min(...parts.map(p=>p.x)), y = Math.min(...parts.map(p=>p.y));
    const right = Math.max(...parts.map(p=>p.x+p.width)), bottom = Math.max(...parts.map(p=>p.y+p.height));
    const canvas = document.createElement('canvas'), scale = 120/Math.max(1,bottom-y), padding = 20;
    canvas.width = Math.ceil((right-x)*scale)+padding*2; canvas.height = 120+padding*2;
    const context = canvas.getContext('2d'); context.fillStyle = '#fff';context.fillRect(0,0,canvas.width,canvas.height);
    const source = document.createElement('canvas');source.width=right-x;source.height=bottom-y;
    const pixels = source.getContext('2d').createImageData(source.width,source.height);
    for (let sy = y; sy < bottom; sy++) for (let sx = x; sx < right; sx++) {
      const i = ((sy-y)*source.width+sx-x)*4, value = mask[sy*width+sx] ? 0 : 255;
      pixels.data[i]=pixels.data[i+1]=pixels.data[i+2]=value;pixels.data[i+3]=255;
    }
    source.getContext('2d').putImageData(pixels,0,0);
    context.drawImage(source,padding,padding,canvas.width-padding*2,120);
    return {canvas,color,possibleJoker};
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

  async function read(image, {progress=()=>{},cancelled=()=>false}={}) {
    const scale = Math.min(1,MAX_DIMENSION/Math.max(image.naturalWidth,image.naturalHeight));
    const canvas = document.createElement('canvas');
    canvas.width = Math.max(1,Math.round(image.naturalWidth*scale));canvas.height = Math.max(1,Math.round(image.naturalHeight*scale));
    const context = canvas.getContext('2d',{willReadFrequently:true});context.drawImage(image,0,0,canvas.width,canvas.height);
    const pixels = context.getImageData(0,0,canvas.width,canvas.height);
    progress('Looking for tiles on this device…');await pause();
    const boxes = group(detect(pixels));
    const {module,api} = boxes.length ? await ready() : {};
    const observations = [];
    for (let i = 0; i < boxes.length; i++) {
      if (cancelled()) throw new Error('Photo reading was replaced.');
      const box=boxes[i], ink=glyph(pixels,box);
      let text='', confidence=0, candidates=[];
      if (!ink) {text='Unreadable tile face';}
      else if (ink.possibleJoker) {text='Possible joker face';candidates=[{tile:{joker:true},confidence:.5}];}
      else {
        const encoded=ink.canvas.toDataURL('image/png').split(',')[1];
        module.FS.writeFile('/input',Uint8Array.from(atob(encoded),c=>c.charCodeAt(0)));
        if (api.SetImageFile(1,0)!==0) throw new Error('The number crop could not be decoded.');
        api.SetVariable('tessedit_pageseg_mode','7');
        api.Recognize(null);text=api.GetUTF8Text().trim();confidence=api.MeanTextConf()/100;
        if (!/^(?:[1-9]|1[0-3])$/.test(text) || confidence<.45) {
          api.SetVariable('tessedit_pageseg_mode','13');
          api.Recognize(null);
          const alternate=api.GetUTF8Text().trim(), score=api.MeanTextConf()/100;
          if (/^(?:[1-9]|1[0-3])$/.test(alternate) && (score>confidence || !text)) {text=alternate;confidence=score;}
        }
        module.FS.unlink('/input');
        if (/^(?:[1-9]|1[0-3])$/.test(text) && confidence>=.3 && ink.color) {
          candidates=[{tile:{color:ink.color,number:Number(text)},confidence}];
        }
      }
      observations.push({rect:{x:box.x/canvas.width,y:box.y/canvas.height,width:box.width/canvas.width,height:box.height/canvas.height},
        text:text||'Uncertain number or color',candidates,suggestedMeld:box.suggestedMeld,suggestedOrder:box.suggestedOrder});
      progress(`Reading tiles on this device… ${i+1} of ${boxes.length}`);await pause();
    }
    return {observations,recognizer:{name:'local-ivory-tiles-tesseract',pipelineRevision:1,coreVersion:'6.1.2',
      modelRevision:'87416418657359cb625c412a48b6e1d6d41c29bd',maxDimension:MAX_DIMENSION},
      notice:`Proposed ${observations.length} tiles locally. Check every value, color, joker, missing tile and meld boundary. Scores are uncalibrated; confirm each tile before continuing.`};
  }
  const api={read,ready,colorOf,detect,group};
  root.RummyReader=api;
  if (typeof module!=='undefined' && module.exports) module.exports=api;
})(globalThis);
