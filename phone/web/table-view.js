/* Responsive view of a position already validated by the local Go engine. */
(function (root) {
  'use strict';
  const NS = 'http://www.w3.org/2000/svg';
  const element = (tag, className, text) => {
    const node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined) node.textContent = text;
    return node;
  };

  function tile(piece, played) {
    const joker = piece.tile.joker === true;
    const role = joker ? piece.as : null;
    const label = joker ? (role ? `Joker as ${role.color} ${role.number}` : 'Joker') : `${piece.tile.color} ${piece.tile.number}`;
    const holder = element('div', 'art-tile' + (played ? ' played' : ''));
    holder.setAttribute('role', 'listitem');
    const svg = document.createElementNS(NS, 'svg');
    svg.setAttribute('viewBox', '0 0 54 80');
    svg.setAttribute('width', '54');
    svg.setAttribute('height', '80');
    svg.setAttribute('focusable', 'false');
    const group = document.createElementNS(NS, 'g');
    for (const [name, value] of Object.entries({
      'data-piece-id':piece.id, 'data-played':String(played),
      'data-tile-color':joker ? '' : piece.tile.color,
      'data-tile-number':joker ? 0 : piece.tile.number,
      'data-joker':String(joker), 'data-role-color':role?.color || '',
      'data-role-number':role?.number || 0
    })) group.setAttribute(name, value);
    const title = document.createElementNS(NS, 'title');
    title.textContent = label + (played ? ' · played from your rack' : '');
    const rect = document.createElementNS(NS, 'rect');
    for (const [name, value] of Object.entries({x:0,y:2,width:54,height:76})) rect.setAttribute(name, value);
    group.append(title, rect);
    svg.append(group);
    root.RummyTileArt.decorate(svg);
    svg.querySelector('.tile-art-caption')?.remove();
    holder.append(svg);
    if (joker) {
      const caption = element('span', 'tile-art-caption joker-role');
      if (role) caption.append(element('span', '', 'Joker as '), element('span', '', `${role.color} ${role.number}`));
      else caption.textContent = 'Joker';
      holder.append(caption);
    }
    return holder;
  }

  function section(title, count, className) {
    const section = element('div', 'board-section ' + className);
    const heading = element('div', 'board-heading');
    heading.append(element('h3', '', title), element('span', 'board-count', count));
    section.append(heading);
    return section;
  }

  function render(position, playedIds = [], {label = 'Suggested table', showRack = true} = {}) {
    if (!position || !Array.isArray(position.table) || !Array.isArray(position.rack)) throw new Error('A validated table and rack are required to render a move');
    const played = new Set(playedIds);
    const board = element('div', 'table-board');
    const total = position.table.reduce((n, meld) => n + meld.length, 0);
    const table = section(label, `${total} tiles · ${position.table.length} ${position.table.length === 1 ? 'meld' : 'melds'}`, 'board-table');
    if (played.size) table.append(element('p', 'board-legend', `${played.size} from your rack · amber outlines`));
    const melds = element('div', 'melds-grid');
    position.table.forEach((meld, index) => {
      const card = element('div', 'meld-card');
      card.append(element('span', 'meld-label', `Meld ${index+1}`));
      const row = element('div', 'meld-row');
      row.setAttribute('role', 'list');
      row.setAttribute('aria-label', `Meld ${index+1}`);
      for (const piece of meld) row.append(tile(piece, played.has(piece.id)));
      card.append(row);
      melds.append(card);
    });
    if (!position.table.length) melds.append(element('p', 'board-empty', 'Table is empty'));
    table.append(melds);
    board.append(table);
    if (showRack) {
      const rack = section('Remaining rack', `${position.rack.length} ${position.rack.length === 1 ? 'tile' : 'tiles'}`, 'board-rack');
      const row = element('div', 'rack-row');
      row.setAttribute('role', 'list');
      row.setAttribute('aria-label', 'Remaining rack');
      for (const piece of position.rack) row.append(tile(piece, false));
      if (!position.rack.length) row.append(element('p', 'board-empty', 'Rack is empty'));
      rack.append(row);
      board.append(rack);
    }
    return board;
  }

  root.RummyTableView = {render};
})(globalThis);
