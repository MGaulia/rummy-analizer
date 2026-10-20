/* Raw photo observations and explicit user edits stay separate. */
(function (root) {
  'use strict';
  const colors = new Set(['black', 'blue', 'orange', 'red']);
  const clone = value => JSON.parse(JSON.stringify(value));
  function validTile(tile) {
    return tile && (tile.joker === true
      ? (!tile.color && !tile.number)
      : colors.has(tile.color) && Number.isInteger(tile.number) && tile.number >= 1 && tile.number <= 13);
  }
  function sameTile(a,b) {
    return a.joker === true ? b.joker === true : b.joker !== true && a.color === b.color && a.number === b.number;
  }
  function proposedTile(raw) {
    const candidates = raw.candidates || [], tile = candidates[0]?.tile;
    // Scores are uncalibrated. Accept one distinct valid reading, not a guess
    // between conflicting candidates or a fallback after an invalid candidate.
    return validTile(tile) && candidates.every(c => validTile(c.tile) && sameTile(tile,c.tile)) ? tile : null;
  }
  function observationReading(observation, index) {
    return observation.correction || {
      tile: proposedTile(observation.raw),
      meld: observation.raw.suggestedMeld,
      order: observation.raw.suggestedOrder || index + 1
    };
  }
  class PhotoSession {
    constructor(zone, source, observations = [], {error = ''} = {}) {
      if (!['table', 'rack'].includes(zone) || !source) throw new Error('Photo needs a zone and source identity');
      this.zone = zone; this.source = source; this.readError = error;
      this.observations = observations.map((o, i) => ({
        id: `${zone}-${source.slice(0, 20)}-${i}`, raw: clone(o), manual: false, correction: null, excluded: false
      }));
    }
    add(rect) {
      const id = `${this.zone}-${this.source.slice(0, 20)}-${this.observations.length}`;
      this.observations.push({id, raw: {rect:clone(rect), candidates: [], text: ''}, manual: true, correction: null, excluded: false});
      return id;
    }
    edit(id, tile, meld, order) {
      const index = this.observations.findIndex(o => o.id === id);
      if (index < 0) throw new Error('Unknown observation');
      const o = this.observations[index];
      if (order === undefined) order = o.raw.suggestedOrder || index + 1;
      // Store invalid/blank drafts too: clearing an input must never fall back to
      // a stale OCR proposal. pieces() supplies the actionable validation error.
      o.correction = {order, tile: tile ? clone(tile) : {}, meld: this.zone === 'table' ? meld : null};
      o.excluded = false;
    }
    exclude(id, excluded = true) {
      const o = this.observations.find(o => o.id === id); if (!o) throw new Error('Unknown observation');
      o.excluded = Boolean(excluded);
    }
    reading(id) {
      const index = this.observations.findIndex(o => o.id === id);
      if (index < 0) throw new Error('Unknown observation');
      return clone(observationReading(this.observations[index], index));
    }
    pieces() {
      if (this.readError) throw new Error(`Could not read the ${this.zone} photo: ${this.readError}`);
      const pieces = [];
      for (const [index, observation] of this.observations.entries()) {
        if (observation.excluded) continue;
        const value = clone(observationReading(observation, index));
        const number = index + 1;
        if (!validTile(value.tile)) throw new Error(`Could not read ${this.zone} tile ${number}. Edit its color and number or remove it.`);
        if (this.zone === 'table' && (!Number.isInteger(value.meld) || value.meld < 1)) throw new Error(`Enter the meld number for table tile ${number}.`);
        if (this.zone === 'table' && (!Number.isInteger(value.order) || value.order < 1)) throw new Error(`Enter the order for table tile ${number}.`);
        pieces.push({
          id: observation.id,
          tile: value.tile,
          meld: this.zone === 'table' ? value.meld : null,
          order: value.order
        });
      }
      if (this.zone === 'rack' && !pieces.length) throw new Error('No rack tiles were read. Add a tile or choose another rack photo.');
      return pieces;
    }
  }
  function positionFromPhotos(table, rack, includeRack = true) {
    if (!table) throw new Error('Choose a table photo first');
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
