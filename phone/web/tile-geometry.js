/* Local, orientation-independent tile-face proposals. OpenCV is supplied by the
 * caller from the bundled runtime; this module performs no network requests. */
(function (root) {
  'use strict';

  const sleep = () => new Promise(resolve => setTimeout(resolve, 0));
  const cross = (a, b, p) => (b.x-a.x)*(p.y-a.y)-(b.y-a.y)*(p.x-a.x);
  const distance = (a, b) => Math.hypot(a.x-b.x, a.y-b.y);

  function area(points) {
    let sum = 0;
    for (let i=0;i<points.length;i++) {
      const a=points[i], b=points[(i+1)%points.length];
      sum += a.x*b.y-a.y*b.x;
    }
    return Math.abs(sum)/2;
  }

  function ordered(points) {
    const center=points.reduce((p,q)=>({x:p.x+q.x/4,y:p.y+q.y/4}),{x:0,y:0});
    const result=points.slice().sort((a,b)=>Math.atan2(a.y-center.y,a.x-center.x)-Math.atan2(b.y-center.y,b.x-center.x));
    let first=0;
    for (let i=1;i<4;i++) if (result[i].x+result[i].y<result[first].x+result[first].y) first=i;
    return result.slice(first).concat(result.slice(0,first));
  }

  function intersection(a, b, c, d) {
    const ab={x:b.x-a.x,y:b.y-a.y}, cd={x:d.x-c.x,y:d.y-c.y};
    const denominator=ab.x*cd.y-ab.y*cd.x;
    if (Math.abs(denominator)<1e-8) return b;
    const fraction=((c.x-a.x)*cd.y-(c.y-a.y)*cd.x)/denominator;
    return {x:a.x+fraction*ab.x,y:a.y+fraction*ab.y};
  }

  function overlap(a, b) {
    let polygon=a.quad;
    for (let i=0;i<4 && polygon.length;i++) {
      const start=b.quad[i], end=b.quad[(i+1)%4], input=polygon;
      polygon=[];
      for (let j=0;j<input.length;j++) {
        const previous=input[(j+input.length-1)%input.length], current=input[j];
        const wasInside=cross(start,end,previous)>=-1e-6, isInside=cross(start,end,current)>=-1e-6;
        if (wasInside!==isInside) polygon.push(intersection(previous,current,start,end));
        if (isInside) polygon.push(current);
      }
    }
    return polygon.length ? area(polygon)/Math.min(a.area,b.area) : 0;
  }

  function bounds(quad) {
    const x=Math.floor(Math.min(...quad.map(p=>p.x))), y=Math.floor(Math.min(...quad.map(p=>p.y)));
    return {x,y,width:Math.ceil(Math.max(...quad.map(p=>p.x)))-x,height:Math.ceil(Math.max(...quad.map(p=>p.y)))-y};
  }

  function percentile(values, fraction) {
    values.sort((a,b)=>a-b);
    if (!values.length) return 0;
    const index=(values.length-1)*fraction, lower=Math.floor(index), remainder=index-lower;
    return values[lower]+(values[Math.min(lower+1,values.length-1)]-values[lower])*remainder;
  }

  function histogramPercentile(histogram, count, fraction) {
    const target=Math.ceil(count*fraction);
    let sum=0;
    for (let i=0;i<256;i++) { sum+=histogram[i]; if (sum>=target) return i; }
    return 255;
  }

  function pointMat(cv, points) {
    return cv.matFromArray(4,1,cv.CV_32FC2,points.flatMap(p=>[p.x,p.y]));
  }

  // The returned Mat belongs to the caller. The input Mat is never deleted.
  function warped(cv, source, quad, width, height) {
    let inputPoints, outputPoints, transform, output;
    try {
      inputPoints=pointMat(cv,quad);
      outputPoints=pointMat(cv,[{x:0,y:0},{x:width-1,y:0},{x:width-1,y:height-1},{x:0,y:height-1}]);
      transform=cv.getPerspectiveTransform(inputPoints,outputPoints);
      output=new cv.Mat();
      cv.warpPerspective(source,output,transform,new cv.Size(width,height),cv.INTER_LINEAR,cv.BORDER_REPLICATE);
      return output;
    } catch (error) {
      output?.delete();
      throw error;
    } finally {
      transform?.delete(); outputPoints?.delete(); inputPoints?.delete();
    }
  }

  function warp(cv, image, quad, width, height) {
    const owned=!(image instanceof cv.Mat);
    const source=owned ? cv.matFromImageData(image) : image;
    let output;
    try {
      width=Math.max(1,Math.round(width)); height=Math.max(1,Math.round(height));
      output=warped(cv,source,quad,width,height);
      const canvas=document.createElement('canvas');
      canvas.width=width; canvas.height=height;
      canvas.getContext('2d').putImageData(new ImageData(new Uint8ClampedArray(output.data),width,height),0,0);
      return canvas;
    } finally {
      output?.delete();
      if (owned) source.delete();
    }
  }

  // Scan convex face interiors directly. The two histograms avoid per-pixel
  // objects and retain contrast information without copying a whole-image ROI.
  function appearance(quad, gray, chroma, width, height, minimumIvory) {
    const center=quad.reduce((p,q)=>({x:p.x+q.x/4,y:p.y+q.y/4}),{x:0,y:0});
    const short=Math.min(...quad.map((p,i)=>distance(p,quad[(i+1)%4])));
    const factor=1-Math.min(.12,3/Math.max(short,1));
    const inner=quad.map(p=>({x:center.x+(p.x-center.x)*factor,y:center.y+(p.y-center.y)*factor}));
    const box=bounds(inner), histogram=new Uint32Array(256), colored=new Uint32Array(256);
    let count=0, ivory=0;
    for (let y=Math.max(0,box.y);y<Math.min(height,box.y+box.height);y++) {
      const intersections=[];
      for (let i=0;i<4;i++) {
        const a=inner[i], b=inner[(i+1)%4], yy=y+.5;
        if ((a.y<=yy && b.y>yy)||(b.y<=yy && a.y>yy)) intersections.push(a.x+(yy-a.y)*(b.x-a.x)/(b.y-a.y));
      }
      if (intersections.length<2) continue;
      const left=Math.max(0,Math.ceil(Math.min(...intersections))), right=Math.min(width-1,Math.floor(Math.max(...intersections)));
      for (let x=left;x<=right;x++) {
        const pixel=y*width+x, value=gray[pixel];
        histogram[value]++; count++;
        if (chroma[pixel]>55) colored[value]++;
        if (value>minimumIvory && chroma[pixel]<90) ivory++;
      }
    }
    if (!count) return null;
    const reference=histogramPercentile(histogram,count,.80);
    let ink=0;
    for (let value=0;value<256;value++) {
      if (value<reference*.65) ink+=histogram[value];
      else if (value<reference*.92) ink+=colored[value];
    }
    return {ivory:ivory/count,ink:ink/count};
  }

  function framing(cv, source, quad) {
    let face, gray, mask, labels, stats, centers;
    try {
      face=warped(cv,source,quad,100,100);
      gray=new cv.Mat(); cv.cvtColor(face,gray,cv.COLOR_RGBA2GRAY);
      const histogram=new Uint32Array(256);
      for (const value of gray.data) histogram[value]++;
      const reference=histogramPercentile(histogram,10000,.8);
      mask=cv.Mat.zeros(100,100,cv.CV_8UC1);
      const rgba=face.data, data=gray.data, inkMask=mask.data;
      for (let y=3;y<97;y++) for (let x=3;x<97;x++) {
        const i=y*100+x, offset=i*4;
        const delta=Math.max(rgba[offset],rgba[offset+1],rgba[offset+2])-Math.min(rgba[offset],rgba[offset+1],rgba[offset+2]);
        inkMask[i]=data[i]<reference*.55 || (delta>65 && data[i]<reference*.92) ? 1 : 0;
      }
      labels=new cv.Mat(); stats=new cv.Mat(); centers=new cv.Mat();
      const count=cv.connectedComponentsWithStats(mask,labels,stats,centers,8,cv.CV_32S);
      let largest=0;
      for (let i=1;i<count;i++) largest=Math.max(largest,stats.data32S[i*5+4]);
      if (largest<=20) return {valid:false,reason:'no-ink'};
      let left=100,top=100,right=0,bottom=0;
      for (let i=1;i<count;i++) {
        const start=i*5;
        if (stats.data32S[start+4]<=Math.max(25,largest*.25)) continue;
        left=Math.min(left,stats.data32S[start]); top=Math.min(top,stats.data32S[start+1]);
        right=Math.max(right,stats.data32S[start]+stats.data32S[start+2]);
        bottom=Math.max(bottom,stats.data32S[start+1]+stats.data32S[start+3]);
      }
      const margin=Math.min(left,top,100-right,100-bottom);
      return {valid:margin>=6,margin};
    } finally {
      centers?.delete(); stats?.delete(); labels?.delete(); mask?.delete(); gray?.delete(); face?.delete();
    }
  }

  // A strip's outer envelope can include dark wedges above stepped tiles.
  // Refine an otherwise unframed child to its own ivory contour rather than
  // interpreting that background as ink or discarding the physical face.
  function refine(cv, source, quad) {
    let face,gray,mask,contours,hierarchy,from,to,inverse,pointInput,pointOutput;
    try {
      face=warped(cv,source,quad,100,100);gray=new cv.Mat();cv.cvtColor(face,gray,cv.COLOR_RGBA2GRAY);
      const histogram=new Uint32Array(256);for(const value of gray.data) histogram[value]++;
      const reference=histogramPercentile(histogram,10000,.8),rgba=face.data,pixels=gray.data;
      mask=cv.Mat.zeros(100,100,cv.CV_8UC1);const target=mask.data;
      for(let i=0;i<10000;i++) {
        const offset=i*4,delta=Math.max(rgba[offset],rgba[offset+1],rgba[offset+2])-Math.min(rgba[offset],rgba[offset+1],rgba[offset+2]);
        target[i]=pixels[i]>reference*.70 && delta<90 ? 255 : 0;
      }
      contours=new cv.MatVector();hierarchy=new cv.Mat();cv.findContours(mask,contours,hierarchy,cv.RETR_EXTERNAL,cv.CHAIN_APPROX_SIMPLE);
      let largest=-1,size=0;
      for(let i=0;i<contours.size();i++) {const contour=contours.get(i);try {const value=Math.abs(cv.contourArea(contour));if(value>size){largest=i;size=value;}}finally{contour.delete();}}
      if(largest<0 || size<4000) return null;
      let contour,hull,approx,points;
      try {
        contour=contours.get(largest);hull=new cv.Mat();approx=new cv.Mat();cv.convexHull(contour,hull);
        const perimeter=cv.arcLength(hull,true);
        for(const epsilon of [.015,.025,.04]) {
          cv.approxPolyDP(hull,approx,perimeter*epsilon,true);
          if(approx.rows===4 && cv.isContourConvex(approx)){points=ordered(Array.from({length:4},(_,i)=>({x:approx.data32S[i*2],y:approx.data32S[i*2+1]})));break;}
        }
      } finally {approx?.delete();hull?.delete();contour?.delete();}
      if(!points || area(points)<5500) return null;
      from=pointMat(cv,[{x:0,y:0},{x:99,y:0},{x:99,y:99},{x:0,y:99}]);to=pointMat(cv,quad);inverse=cv.getPerspectiveTransform(from,to);
      pointInput=pointMat(cv,points);pointOutput=new cv.Mat();cv.perspectiveTransform(pointInput,pointOutput,inverse);
      return ordered(Array.from({length:4},(_,i)=>({x:pointOutput.data32F[i*2],y:pointOutput.data32F[i*2+1]})));
    } finally {pointOutput?.delete();pointInput?.delete();inverse?.delete();to?.delete();from?.delete();hierarchy?.delete();contours?.delete();mask?.delete();gray?.delete();face?.delete();}
  }

  function framed(cv, source, candidate) {
    let result={...candidate,frame:framing(cv,source,candidate.quad)};
    if(!result.frame.valid && result.frame.reason!=='no-ink') {
      const quad=refine(cv,source,candidate.quad);
      if(quad) {const frame=framing(cv,source,quad);if(frame.valid) result={...candidate,quad,area:area(quad),frame};}
    }
    return result;
  }

  function seamValleys(gray, width, height, vertical) {
    const length=vertical ? height : width, perpendicular=vertical ? width : height;
    const first=Math.max(1,Math.floor(perpendicular*.15)), last=Math.max(first+1,Math.floor(perpendicular*.85));
    const at=(position,row)=>vertical ? gray[position*width+row] : gray[row*width+position];
    const scores=new Float32Array(length);
    for (let x=7;x<length-7;x++) {
      const contrasts=[],steps=[];
      for (let row=first;row<last;row++) {
        let dark=255,left=0,right=0;
        for (let offset=-1;offset<=1;offset++) dark=Math.min(dark,at(x+offset,row));
        for (let offset=3;offset<=7;offset++) {
          left=Math.max(left,at(x-offset,row)); right=Math.max(right,at(x+offset,row));
        }
        contrasts.push(Math.min(left,right)-dark);
        let before=0,after=0;
        for(let offset=1;offset<=3;offset++){before+=at(x-offset-1,row)/3;after+=at(x+offset,row)/3;}
        steps.push(Math.abs(after-before));
      }
      const capContrast=(start,end)=>{
        const drops=[];
        for(let row=start;row<end;row++) {
          let dark=255,left=0,right=0;
          for(let offset=-1;offset<=1;offset++) dark=Math.min(dark,at(x+offset,row));
          for(let offset=3;offset<=7;offset++){left=Math.max(left,at(x-offset,row));right=Math.max(right,at(x+offset,row));}
          drops.push(Math.min(left,right)-dark);
        }
        return percentile(drops,.85);
      };
      const capSize=Math.max(3,Math.floor(perpendicular*.15));
      const capScore=Math.min(capContrast(1,capSize),capContrast(perpendicular-capSize,perpendicular-1));
      scores[x]=Math.max(percentile(contrasts,.25),percentile(steps,.25),capScore);
    }
    const ranked=Array.from({length},(_,i)=>i).sort((a,b)=>scores[b]-scores[a]);
    const selected=[];
    for (const x of ranked) {
      if (scores[x]<7) break;
      if (x<8 || x>length-9 || Math.min(x,length-x)<=perpendicular*.2) continue;
      if (selected.some(previous=>Math.abs(previous-x)<Math.max(6,perpendicular*.18))) continue;
      selected.push(x);
      if (selected.length===10) break;
    }
    return {positions:selected.sort((a,b)=>a-b),scores,length};
  }

  function inkCuts(cv, face, grayMat, width, height, vertical) {
    let mask,labels,stats,centers;
    try {
      const histogram=new Uint32Array(256),gray=grayMat.data,rgba=face.data;
      for(const value of gray) histogram[value]++;
      const reference=histogramPercentile(histogram,gray.length,.8);
      mask=cv.Mat.zeros(height,width,cv.CV_8UC1);const target=mask.data;
      for(let y=3;y<height-3;y++)for(let x=3;x<width-3;x++) {
        const i=y*width+x,offset=i*4,delta=Math.max(rgba[offset],rgba[offset+1],rgba[offset+2])-Math.min(rgba[offset],rgba[offset+1],rgba[offset+2]);
        target[i]=gray[i]<reference*.55 || (delta>65 && gray[i]<reference*.92) ? 1 : 0;
      }
      labels=new cv.Mat();stats=new cv.Mat();centers=new cv.Mat();
      const count=cv.connectedComponentsWithStats(mask,labels,stats,centers,8,cv.CV_32S),data=stats.data32S,components=[];
      let largest=0;
      for(let i=1;i<count;i++) {
        const offset=i*5,x=data[offset],y=data[offset+1],w=data[offset+2],h=data[offset+3],size=data[offset+4];
        if(size<Math.max(6,width*height*.001) || x<=3 || y<=3 || x+w>=width-3 || y+h>=height-3) continue;
        largest=Math.max(largest,w,h);components.push({x,y,w,h});
      }
      const glyphs=components.filter(c=>Math.max(c.w,c.h)>=largest*.45).map(c=>({
        start:vertical?c.y:c.x,end:vertical?c.y+c.h:c.x+c.w,perpendicular:vertical?c.w:c.h
      })).sort((a,b)=>a.start-b.start);
      const clusters=[];
      for(const glyph of glyphs) {
        const previous=clusters[clusters.length-1];
        if(previous && glyph.start-previous.end<Math.max(previous.perpendicular,glyph.perpendicular)*.30) {
          previous.end=Math.max(previous.end,glyph.end);previous.perpendicular=Math.max(previous.perpendicular,glyph.perpendicular);
        } else clusters.push({...glyph});
      }
      if(clusters.length<2 || clusters.length>13) return [];
      const cuts=[];
      for(let i=1;i<clusters.length;i++) cuts.push(Math.round((clusters[i-1].end+clusters[i].start)/2));
      return cuts;
    } finally {centers?.delete();stats?.delete();labels?.delete();mask?.delete();}
  }

  function split(cv, source, candidate) {
    const q=candidate.quad;
    const originalWidth=Math.max(distance(q[0],q[1]),distance(q[2],q[3]));
    const originalHeight=Math.max(distance(q[1],q[2]),distance(q[3],q[0]));
    const scale=Math.min(1,480/Math.max(originalWidth,originalHeight));
    const width=Math.max(12,Math.round(originalWidth*scale)), height=Math.max(12,Math.round(originalHeight*scale));
    let face, grayMat, from, to, inverse;
    try {
      face=warped(cv,source,q,width,height);
      grayMat=new cv.Mat(); cv.cvtColor(face,grayMat,cv.COLOR_RGBA2GRAY);
      const gray=Uint8Array.from(grayMat.data);
      from=pointMat(cv,[{x:0,y:0},{x:width-1,y:0},{x:width-1,y:height-1},{x:0,y:height-1}]);
      to=pointMat(cv,q); inverse=cv.getPerspectiveTransform(from,to);
      const transform=points=>{
        let input,output;
        try {
          input=pointMat(cv,points);output=new cv.Mat();cv.perspectiveTransform(input,output,inverse);
          return ordered(Array.from({length:4},(_,i)=>({x:output.data32F[i*2],y:output.data32F[i*2+1]})));
        } finally {output?.delete();input?.delete();}
      };
      let best=null;
      for (const vertical of [false,true]) {
        const {positions,scores,length}=seamValleys(gray,width,height,vertical);
        const cuts=positions.slice();
        for (let attempt=0;attempt<12 && cuts.length;attempt++) {
          const stops=[0,...cuts,length-1],parts=[];
          for (let i=0;i<stops.length-1;i++) {
            const left=stops[i],right=stops[i+1];
            const points=vertical
              ? [{x:0,y:left},{x:width-1,y:left},{x:width-1,y:right},{x:0,y:right}]
              : [{x:left,y:0},{x:right,y:0},{x:right,y:height-1},{x:left,y:height-1}];
            const quad=transform(points),part=framed(cv,source,{quad,area:area(quad),method:'seam'});
            const lengths=part.quad.map((point,n)=>distance(point,part.quad[(n+1)%4]));
            // A seam through a two-digit number can look like a narrow face.
            // Genuine neighbouring faces retain a plausible side-length ratio.
            if(Math.max(...lengths)/Math.min(...lengths)>3) part.frame={valid:false,reason:'slender-fragment'};
            parts.push(part);
          }
          if (parts.every(part=>part.frame.valid)) {
            if (!best || parts.length>best.length) best=parts;
            break;
          }
          let remove=0,penalty=Infinity,score=Infinity;
          for (let i=0;i<cuts.length;i++) {
            const value=Number(parts[i].frame.valid)+Number(parts[i+1].frame.valid);
            if (value<penalty || (value===penalty && scores[cuts[i]]<score)) {
              remove=i;penalty=value;score=scores[cuts[i]];
            }
          }
          cuts.splice(remove,1);
        }
      }
      // Faint bevels can be a brightness step rather than a dark seam. Glyph
      // clusters provide an independent geometric proposal. Nearby digit parts
      // remain one number; each proposed face must still be framed and have a
      // plausible aspect, protecting real 10–13/11 from splitting into digits.
      for(const vertical of [false,true]) {
        const cuts=inkCuts(cv,face,grayMat,width,height,vertical);
        if(!cuts.length) continue;
        const length=vertical?height:width,stops=[0,...cuts,length-1],parts=[];
        for(let i=0;i<stops.length-1;i++) {
          const left=stops[i],right=stops[i+1],points=vertical
            ? [{x:0,y:left},{x:width-1,y:left},{x:width-1,y:right},{x:0,y:right}]
            : [{x:left,y:0},{x:right,y:0},{x:right,y:height-1},{x:left,y:height-1}];
          const quad=transform(points),lengths=quad.map((p,n)=>distance(p,quad[(n+1)%4]));
          if(Math.max(...lengths)/Math.min(...lengths)>2.6){parts.length=0;break;}
          parts.push(framed(cv,source,{quad,area:area(quad),method:'ink-spacing'}));
        }
        if(parts.length>1 && parts.every(part=>part.frame.valid) && (!best || parts.length>best.length)) best=parts;
      }
      return best || [candidate];
    } finally {
      inverse?.delete();to?.delete();from?.delete();grayMat?.delete();face?.delete();
    }
  }

  function housing(cv, imageData) {
    let mask,labels,stats,centers;
    try {
      const {width,height,data}=imageData;
      mask=cv.Mat.zeros(height,width,cv.CV_8UC1);
      const pixels=mask.data;
      for (let i=0;i<width*height;i++) {
        const r=data[i*4],g=data[i*4+1],b=data[i*4+2];
        pixels[i]=b>r*1.2 && b>g*1.12 && b-r>35 && b>50 ? 1 : 0;
      }
      labels=new cv.Mat();stats=new cv.Mat();centers=new cv.Mat();
      const count=cv.connectedComponentsWithStats(mask,labels,stats,centers,8,cv.CV_32S);
      let label=0,size=0;
      for (let i=1;i<count;i++) if (stats.data32S[i*5+4]>size) {label=i;size=stats.data32S[i*5+4];}
      if (!label || size<width*height*.003) return null;
      const offset=label*5,box={x:stats.data32S[offset],y:stats.data32S[offset+1],width:stats.data32S[offset+2],height:stats.data32S[offset+3]};
      if (Math.max(box.width,box.height)<Math.max(width,height)*.22) return null;
      const component=new Uint8Array(width*height),source=labels.data32S;
      for (let i=0;i<component.length;i++) component[i]=source[i]===label ? 1 : 0;
      return {pixels:component,width,height,box,area:size};
    } finally {centers?.delete();stats?.delete();labels?.delete();mask?.delete();}
  }

  function touchesHousing(quad, blue) {
    if (!blue) return false;
    const center=quad.reduce((p,q)=>({x:p.x+q.x/4,y:p.y+q.y/4}),{x:0,y:0});
    let contacts=0;
    for (const factor of [.96,1.06,1.16]) for (let side=0;side<4;side++) {
      const a=quad[side],b=quad[(side+1)%4];
      for (let i=0;i<=20;i++) {
        const t=i/20,x=Math.round(center.x+(a.x+(b.x-a.x)*t-center.x)*factor),y=Math.round(center.y+(a.y+(b.y-a.y)*t-center.y)*factor);
        if (x>=0 && y>=0 && x<blue.width && y<blue.height && blue.pixels[y*blue.width+x]) contacts++;
      }
    }
    return contacts>=3;
  }

  async function detect(cv, imageData, {zone='table',cancelled=()=>false,progress=()=>{}}={}) {
    if (!['table','rack'].includes(zone)) throw new Error('Invalid photo zone.');
    const check=()=>{if (cancelled()) throw new Error('Photo reading cancelled.');};
    check();
    let source,grayMat,blur,kernel2,kernel3,kernel5,kernel7;
    try {
      source=cv.matFromImageData(imageData);grayMat=new cv.Mat();blur=new cv.Mat();
      cv.cvtColor(source,grayMat,cv.COLOR_RGBA2GRAY);cv.GaussianBlur(grayMat,blur,new cv.Size(3,3),0);
      const {width,height,data}=imageData,gray=Uint8Array.from(grayMat.data),chroma=new Uint8Array(width*height),histogram=new Uint32Array(256);
      for (let i=0;i<chroma.length;i++) {
        const r=data[i*4],g=data[i*4+1],b=data[i*4+2];
        chroma[i]=Math.max(r,g,b)-Math.min(r,g,b);histogram[gray[i]]++;
      }
      const reference=histogramPercentile(histogram,gray.length,.98),exposure=Math.min(1,reference/215);
      const blue=housing(cv,imageData),minimumArea=width*height*.00025,candidates=[],seen=new Set();
      kernel2=cv.Mat.ones(2,2,cv.CV_8UC1);kernel3=cv.Mat.ones(3,3,cv.CV_8UC1);kernel5=cv.Mat.ones(5,5,cv.CV_8UC1);kernel7=cv.Mat.ones(7,7,cv.CV_8UC1);
      const white=threshold=>{
        const mask=cv.Mat.zeros(height,width,cv.CV_8UC1),cutoff=threshold*exposure,target=mask.data;
        for (let i=0;i<gray.length;i++) target[i]=gray[i]>cutoff && chroma[i]<90 ? 255 : 0;
        return mask;
      };
      const inspect=async(mask,method,prepare=()=>{})=>{
        let contours,hierarchy;
        try {
          check();prepare(mask);contours=new cv.MatVector();hierarchy=new cv.Mat();
          cv.findContours(mask,contours,hierarchy,cv.RETR_LIST,cv.CHAIN_APPROX_SIMPLE);
          for (let i=0;i<contours.size();i++) {
            let contour,approx;
            try {
              contour=contours.get(i);
              const contourArea=Math.abs(cv.contourArea(contour));
              if (contourArea<minimumArea*.5) continue;
              const perimeter=cv.arcLength(contour,true);
              approx=new cv.Mat();
              for (const epsilon of [.015,.025,.04]) {
                cv.approxPolyDP(contour,approx,perimeter*epsilon,true);
                if (approx.rows!==4 || !cv.isContourConvex(approx)) continue;
                const quad=ordered(Array.from({length:4},(_,n)=>({x:approx.data32S[n*2],y:approx.data32S[n*2+1]})));
                const size=area(quad),rectangle=cv.minAreaRect(approx),rw=rectangle.size.width,rh=rectangle.size.height;
                if (size<minimumArea || size>width*height*.5 || Math.min(rw,rh)<10 || Math.max(rw,rh)/Math.min(rw,rh)>12 || size/(rw*rh)<.6) continue;
                const key=quad.map(p=>p.x+','+p.y).join(';');
                if (seen.has(key)) continue;
                seen.add(key);
                const looks=appearance(quad,gray,chroma,width,height,140*exposure);
                if (!looks || looks.ivory<.6 || looks.ink>.32) continue;
                const blank=looks.ink<.005;
                if (blank && (size/contourArea<.94 || Math.max(rw,rh)/Math.min(rw,rh)<1.2 || Math.max(rw,rh)/Math.min(rw,rh)>2.2)) continue;
                if (blank && quad.every(p=>p.x<2 || p.y<2 || p.x>width-3 || p.y>height-3)) continue;
                candidates.push({quad,area:size,ivory:looks.ivory,blank,method});
              }
            } finally {approx?.delete();contour?.delete();}
          }
        } finally {hierarchy?.delete();contours?.delete();mask.delete();}
        await sleep();
      };
      progress('Locating tile faces on this device…');
      for (const [low,high] of [[15,50],[35,100],[60,150]]) {
        let edge;
        try {
          edge=new cv.Mat();cv.Canny(blur,edge,low,high);
          await inspect(edge.clone(),'edge');
          await inspect(new cv.Mat(),'edge-close',closed=>cv.morphologyEx(edge,closed,cv.MORPH_CLOSE,kernel3));
        } finally {edge?.delete();}
      }
      for (const threshold of [140,160,180,200,220]) await inspect(white(threshold),'ivory');
      for (const [low,high] of [[5,15],[10,30],[15,45]]) {
        let edge,closed,dilated,inverse;
        try {
          edge=new cv.Mat();closed=new cv.Mat();dilated=new cv.Mat();inverse=new cv.Mat();
          cv.Canny(blur,edge,low,high);cv.morphologyEx(edge,closed,cv.MORPH_CLOSE,kernel3);cv.dilate(closed,dilated,kernel2);cv.bitwise_not(dilated,inverse);
          for (const threshold of [140,170,200]) {
            await inspect(white(threshold),'separated-ivory',mask=>cv.bitwise_and(mask,inverse,mask));
          }
        } finally {inverse?.delete();dilated?.delete();closed?.delete();edge?.delete();}
      }
      for (const threshold of [140,170,200]) {
        let mask;
        try {
          mask=white(threshold);
          for (const kernel of [kernel3,kernel5,kernel7]) {
            await inspect(new cv.Mat(),'opened-ivory',opened=>cv.morphologyEx(mask,opened,cv.MORPH_OPEN,kernel));
          }
        } finally {mask?.delete();}
      }
      check();
      const distinct=[];
      candidates.sort((a,b)=>b.ivory-a.ivory || b.area-a.area);
      for (const candidate of candidates) {
        if (distinct.some(other=>candidate.area/other.area>.60 && candidate.area/other.area<1.66 && overlap(candidate,other)>.65)) continue;
        distinct.push(candidate);
      }
      const faces=[];
      for (let i=0;i<distinct.length;i++) {
        check();
        const candidate=distinct[i];
        for (const proposed of split(cv,source,candidate)) {
          const part=proposed.frame ? proposed : framed(cv,source,proposed);
          if (!part.frame.valid && !part.blank) continue;
          const lengths=part.quad.map((point,n)=>distance(point,part.quad[(n+1)%4]));
          // Proposals may start as long touching strips; only their resulting
          // faces may survive. Thin quadrilaterals in wood grain are background.
          if(Math.max(...lengths)/Math.min(...lengths)>3) continue;
          const touches=touchesHousing(part.quad,blue);
          if (blue && (zone==='rack' ? !touches : touches)) continue;
          faces.push(part);
        }
        if (i%12===0) await sleep();
      }
      // Container suppression uses validated faces, not arbitrary interior
      // contours such as the holes in numerals or embossed logos.
      const validated=faces.filter(face=>!face.blank);
      const typical=validated.length ? percentile(validated.map(face=>face.area),.5) : 0;
      const physical=faces.filter(face=>{
        if (face.blank && typical && face.area<typical*.1) return false;
        const children=validated.filter(other=>other.area<face.area*.65 && overlap(face,other)>.8);
        return !(children.length>1 && children.reduce((sum,p)=>sum+p.area,0)>face.area*.5);
      });
      const boxes=[];
      for (const face of physical.sort((a,b)=>b.area-a.area)) {
        if (boxes.some(other=>overlap(face,other)>.75)) continue;
        boxes.push({...bounds(face.quad),quad:face.quad,area:face.area,method:face.method});
      }
      return {boxes,diagnostics:{candidateCount:candidates.length,distinctCount:distinct.length,exposureReference:reference,housing:blue ? {box:blue.box,area:blue.area} : null}};
    } finally {
      kernel7?.delete();kernel5?.delete();kernel3?.delete();kernel2?.delete();blur?.delete();grayMat?.delete();source?.delete();
    }
  }

  root.RummyTileGeometry={detect,warp};
  if (typeof module!=='undefined' && module.exports) module.exports=root.RummyTileGeometry;
})(globalThis);
