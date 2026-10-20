/* Raw photo observations and explicit user corrections stay separate. */
(function (root) {
  'use strict';
  const colors = new Set(['black', 'blue', 'orange', 'red']);
  function validTile(tile) {
    return tile && (tile.joker === true
      ? (!tile.color && !tile.number)
      : colors.has(tile.color) && Number.isInteger(tile.number) && tile.number >= 1 && tile.number <= 13);
  }
  class PhotoSession {
    constructor(zone, source, observations = []) {
      if (!['table', 'rack'].includes(zone) || !source) throw new Error('Photo needs a zone and source identity');
      this.zone = zone; this.source = source; this.confirmed = false;
      this.observations = observations.map((o, i) => ({
        id: `${zone}-${source.slice(0, 20)}-${i}`, raw: JSON.parse(JSON.stringify(o)), manual: false, correction: null, excluded: false
      }));
    }
    add(rect) {
      const id = `${this.zone}-${this.source.slice(0, 20)}-${this.observations.length}`;
      this.observations.push({id, raw: {rect, candidates: [], text: ''}, manual: true, correction: null, excluded: false});
      this.confirmed = false; return id;
    }
    correct(id, tile, meld, order) {
      const o = this.observations.find(o => o.id === id);
      if (!o || !validTile(tile) || (this.zone === 'table' && (!Number.isInteger(meld) || meld < 1))) throw new Error('Confirm a valid tile and table meld number');
      if (order === undefined) order = this.observations.indexOf(o) + 1;
      if (this.zone === 'table' && (!Number.isInteger(order) || order < 1)) throw new Error('Confirm a tile order within the meld');
      o.correction = {order, tile: JSON.parse(JSON.stringify(tile)), meld: this.zone === 'table' ? meld : null};
      o.excluded = false; this.confirmed = false;
    }
    exclude(id, excluded) {
      const o = this.observations.find(o => o.id === id); if (!o) throw new Error('Unknown observation');
      o.excluded = excluded; this.confirmed = false;
    }
    confirm(complete) {
      if (!complete || this.observations.some(o => !o.excluded && !o.correction)) throw new Error('Correct every reading and check for missing tiles before confirming');
      this.confirmed = true;
    }
    pieces() {
      if (!this.confirmed) throw new Error(`${this.zone} photo still needs confirmation`);
      return this.observations.filter(o => !o.excluded).map(o => ({id: o.id, tile: {...o.correction.tile}, meld: o.correction.meld, order: o.correction.order}));
    }
  }
  function positionFromPhotos(table, rack, includeRack = true) {
    if (!table) throw new Error('Confirm a table photo first');
    if (includeRack && !rack) throw new Error('A separate rack photo is required to search');
    if (rack && table.source === rack.source) throw new Error('Table and rack must be separate photos');
    const melds = new Map();
    for (const p of table.pieces()) {
      if (!melds.has(p.meld)) melds.set(p.meld, []);
      const meld = melds.get(p.meld);
      if (meld.some(q => q.order === p.order)) throw new Error('Two table tiles have the same order in a meld');
      meld.push({id: p.id, tile: p.tile, order: p.order});
    }
    return {
      table: [...melds.entries()].sort((a,b) => a[0] - b[0]).map(([,pieces]) => pieces.sort((a,b) => a.order - b.order).map(({id,tile}) => ({id,tile}))),
      rack: includeRack ? rack.pieces().map(p => ({id: p.id, tile: p.tile})) : []
    };
  }
  const api = {PhotoSession, positionFromPhotos, validTile};
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  root.RummyPhotos = api;
})(globalThis);
