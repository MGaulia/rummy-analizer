/* Decorate confirmed diagrams with the bundled Printables tile renders.
 * Physical occurrence identity and joker roles still come from the Go engine.
 */
(function (root) {
  'use strict';
  const NS = 'http://www.w3.org/2000/svg';
  const rows = {red:0,blue:1,black:2,orange:3};
  const numberPattern = /^(?:[1-9]|1[0-3])$/;

  function decorate(svg) {
    if (!svg || svg.namespaceURI !== NS || svg.localName !== 'svg') return svg;
    const source = root.document.querySelector('meta[name="rummy-preview-source"]')?.content;
    const href = 'tile-art/tiles-atlas.png' + (source ? '?v='+encodeURIComponent(source) : '');
    for (const piece of svg.querySelectorAll('[data-piece-id]')) {
      if (piece.getAttribute('data-tile-art') === 'true') continue;
      const joker = piece.getAttribute('data-joker');
      const color = piece.getAttribute('data-tile-color');
      const number = piece.getAttribute('data-tile-number');
      const roleColor = piece.getAttribute('data-role-color');
      const roleNumber = piece.getAttribute('data-role-number');
      if (joker !== 'true' && joker !== 'false') continue;
      const isJoker = joker === 'true';
      if (!isJoker && (!Object.hasOwn(rows,color) || !numberPattern.test(number))) continue;
      const assigned = isJoker && Object.hasOwn(rows,roleColor) && numberPattern.test(roleNumber);
      if (isJoker && !assigned && (roleColor || (roleNumber && roleNumber !== '0'))) continue;
      const rect = [...piece.children].find(child => child.localName === 'rect');
      if (!rect) continue;
      const x = Number(rect.getAttribute('x')), y = Number(rect.getAttribute('y'));
      const width = Number(rect.getAttribute('width')), height = Number(rect.getAttribute('height'));
      if (![x,y,width,height].every(Number.isFinite) || width <= 0 || height <= 0) continue;

      const imageHeight = height;
      const imageWidth = imageHeight*192/288;
      const imageX = x+(width-imageWidth)/2;
      const row = isJoker ? rows.black : rows[color];
      const column = isJoker ? 13 : Number(number)-1;
      const sprite = svg.ownerDocument.createElementNS(NS,'svg');
      sprite.setAttribute('x',imageX);
      sprite.setAttribute('y',y);
      sprite.setAttribute('width',imageWidth);
      sprite.setAttribute('height',imageHeight);
      sprite.setAttribute('viewBox',`${column*192} ${row*288} 192 288`);
      sprite.setAttribute('overflow','hidden');
      sprite.setAttribute('aria-hidden','true');
      sprite.setAttribute('focusable','false');
      sprite.setAttribute('class','tile-art-sprite');
      const image = svg.ownerDocument.createElementNS(NS,'image');
      image.setAttribute('href',href);
      image.setAttribute('width','2688');
      image.setAttribute('height','1152');
      sprite.append(image);

      for (const child of [...piece.children]) {
        if (child.localName === 'rect' || child.localName === 'text') child.remove();
      }
      piece.append(sprite);
      if (piece.getAttribute('data-played') === 'true') {
        const outline = svg.ownerDocument.createElementNS(NS,'rect');
        outline.setAttribute('x',imageX);
        outline.setAttribute('y',y);
        outline.setAttribute('width',imageWidth);
        outline.setAttribute('height',imageHeight);
        outline.setAttribute('rx','4');
        outline.setAttribute('fill','none');
        outline.setAttribute('stroke','#007a41');
        outline.setAttribute('stroke-width','3');
        outline.setAttribute('class','tile-art-played');
        piece.append(outline);
      }
      if (isJoker) {
        const caption = svg.ownerDocument.createElementNS(NS,'text');
        caption.setAttribute('x',x+width/2);
        caption.setAttribute('y',y+height+12);
        caption.setAttribute('text-anchor','middle');
        caption.setAttribute('font-size','10');
        caption.setAttribute('fill','#253d31');
        caption.setAttribute('class','tile-art-caption');
        caption.textContent = assigned ? `Joker as ${roleColor} ${roleNumber}` : 'Joker';
        if (assigned) {
          caption.setAttribute('textLength',width-4);
          caption.setAttribute('lengthAdjust','spacingAndGlyphs');
        }
        piece.append(caption);
      }
      const title = [...piece.children].find(child => child.localName === 'title');
      if (title) {
        piece.setAttribute('role','img');
        piece.setAttribute('aria-label',title.textContent);
      }
      piece.setAttribute('data-tile-art','true');
    }
    return svg;
  }

  root.RummyTileArt = {decorate};
})(globalThis);
