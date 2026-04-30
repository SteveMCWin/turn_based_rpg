let state = null;
let pendingBonuses = {};
let remainingPoints = 0;
let undoStack = [];

(async () => {
  state = await loadState();
  if (!state) return;
  renderPage();
  initTooltips();
  const savedScroll = sessionStorage.getItem('mapScroll');
  if (savedScroll !== null) {
    sessionStorage.removeItem('mapScroll');
    requestAnimationFrame(() => { document.querySelector('.floor-list').scrollTop = parseInt(savedScroll, 10); });
  }
})();

function renderPage() {
  renderHeroPanel();
  renderFloors();
  if (state.pending_level_up) showLevelUpOverlay();
}

function renderHeroPanel() {
  const h = state.player;
  const portrait = document.getElementById('hero-panel-portrait');
  if (portrait) portrait.src = heroSprite(h.id);
  const badge = document.getElementById('endless-mode-badge');
  if (badge) badge.classList.toggle('hidden', !state.is_endless);
  document.getElementById('hero-name').textContent      = h.name;
  document.getElementById('hero-hp-text').textContent   = `HP: ${h.current_hp} / ${maxHP(h)}`;
  document.getElementById('hero-mana-text').textContent = `MP: ${h.current_mana || 0} / ${maxMana(h)}`;
  document.getElementById('hero-hp-bar').style.width    = pct(h.current_hp, maxHP(h));
  document.getElementById('hero-mana-bar').style.width  = pct(h.current_mana || 0, maxMana(h));
  document.getElementById('hero-stats').innerHTML = `
    <span>LVL ${h.level}</span>
    <span>${statIcon('attack')}ATK ${effAtk(h)}</span>
    <span>${statIcon('defense')}DEF ${effDef(h)}</span>
    <span>${statIcon('magic')}MAG ${effMag(h)}</span>
    <span class="xp-text">XP ${h.current_xp || 0}</span>
  `;
  document.getElementById('hero-gold').innerHTML = `<img class="stat-icon" src="/sprites/gold.png" alt="Gold" onerror="this.style.display='none'">${h.current_gold || 0}`;
  document.getElementById('equipped-list').innerHTML = (h.equipped_moves || []).map(id => {
    const move = state.moves?.[id];
    const lm   = (h.learned_moves || []).find(m => m.move_id === id);
    return `<li data-tooltip="${escHtml(moveTooltipHTML(move))}">${escHtml(move?.name || id)} <span class="move-level">Lv.${lm?.level || 1}</span></li>`;
  }).join('');
}

function renderFloors() {
  const container = document.getElementById('floors');
  container.innerHTML = (state.floors || []).map((floor, fi) => {
    const nodes = (floor.Rooms || []).map(room => renderRoomNode(room, floor)).join('');
    return `
      <div class="floor-section ${floor.IsCompleted ? 'floor-done' : ''}">
        <div class="floor-label">Floor ${fi + 1}</div>
        <div class="floor-row">${nodes}</div>
      </div>`;
  }).join('');
  requestAnimationFrame(drawConnections);
}

function renderRoomNode(room, floor) {
  const enc  = room.Encounter;
  const kind = enc?.kind;
  const floorHasCompleted = (floor.Rooms || []).some(r => r.IsCompleted);

  const isBoss     = kind === 'boss';
  const domId      = 'room-' + room.Id.replace(',', '_');
  const tooltip    = roomTooltipHTML(room);

  const isRematch = room.IsCompleted && (kind === 'monster' || isBoss) && room.CanEnter && !state.pending_level_up;
  const canClick  = isRematch || (!room.IsCompleted && room.CanEnter && !state.pending_level_up);
  const isLocked  = !room.CanEnter || (floorHasCompleted && !room.IsCompleted && !isRematch);

  const classes = ['room-node', kind === 'monster' ? 'monster' : isBoss ? 'boss' : 'event'];
  if (room.IsCompleted) classes.push('done');
  if (isLocked)         classes.push('locked');
  if (canClick)         classes.push('can-enter');

  const spriteName = kind === 'monster' ? 'monster_encounter' : isBoss ? 'boss_encounter' : 'event_encounter';
  const onclick    = canClick ? `onclick="enterRoom('${room.Id}')"` : '';
  const check      = room.IsCompleted && !isRematch ? '<span class="room-done-check">✓</span>' : '';
  const envLabel   = (kind === 'monster' || isBoss) && room.Environment?.Name
    ? `<span class="room-env-label">${escHtml(room.Environment.Name)}</span>` : '';

  return `<div class="${classes.join(' ')}" id="${domId}" data-tooltip="${escHtml(tooltip)}" ${onclick}>
    <img src="/sprites/${spriteName}.png" alt="${kind}">${check}${envLabel}
  </div>`;
}

function roomTooltipHTML(room) {
  const enc  = room.Encounter;
  const kind = enc?.kind;
  if ((kind === 'monster' || kind === 'boss') && enc?.monster) {
    const m   = enc.monster;
    const env = room.Environment;
    let tip = `<b>${escHtml(m.name)}</b> Lv.${m.level}`;
    if (kind === 'boss') tip += ' <span class="tooltip-badge">BOSS</span>';
    tip += `<br>HP ${m.current_hp}/${maxHP(m)} &middot; ATK ${effAtk(m)} DEF ${effDef(m)} MAG ${effMag(m)}`;
    if (env?.Name) tip += `<br><span class="tooltip-badge">${escHtml(env.Name)}</span> ${escHtml(env.Description)}`;
    return tip;
  }
  if (room.IsCompleted && enc?.event?.description) {
    return escHtml(enc.event.description);
  }
  return 'Unknown event';
}

function getOffsetFrom(el, container) {
  let x = 0, y = 0;
  let node = el;
  while (node && node !== container) {
    x += node.offsetLeft;
    y += node.offsetTop;
    node = node.offsetParent;
  }
  return { x, y };
}

function findRoom(id) {
  for (const floor of (state.floors || []))
    for (const room of (floor.Rooms || []))
      if (room.Id === id) return room;
  return null;
}

function isRoomAccessible(room) {
  if (!room) return false;
  const kind = room.Encounter?.kind;
  const isBoss = kind === 'boss';
  const isRematch = room.IsCompleted && (kind === 'monster' || isBoss) && room.CanEnter;
  return isRematch || (!room.IsCompleted && room.CanEnter);
}

function drawConnections() {
  const container = document.getElementById('floors');
  const old = document.getElementById('connections-svg');
  if (old) old.remove();

  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.id = 'connections-svg';
  svg.setAttribute('aria-hidden', 'true');
  container.prepend(svg);

  // Find the completed room on the highest floor — that's the one whose exits to highlight
  let lastCompleted = null;
  for (let fi = (state.floors || []).length - 1; fi >= 0; fi--) {
    const r = (state.floors[fi].Rooms || []).find(r => r.IsCompleted);
    if (r) { lastCompleted = r; break; }
  }

  (state.floors || []).forEach(floor => {
    (floor.Rooms || []).forEach(room => {
      if (!room.NextRoomIDs?.length) return;
      const fromEl = document.getElementById('room-' + room.Id.replace(',', '_'));
      if (!fromEl) return;
      const fo = getOffsetFrom(fromEl, container);
      const fx = fo.x + fromEl.offsetWidth  / 2;
      const fy = fo.y + fromEl.offsetHeight / 2;

      room.NextRoomIDs.forEach(nextId => {
        const toEl = document.getElementById('room-' + nextId.replace(',', '_'));
        if (!toEl) return;
        const to = getOffsetFrom(toEl, container);
        const tx = to.x + toEl.offsetWidth  / 2;
        const ty = to.y + toEl.offsetHeight / 2;

        const toRoom    = findRoom(nextId);
        const active    = room.Id === lastCompleted?.Id && isRoomAccessible(toRoom);
        const stroke    = active ? '#ffaa5e' : '#544e68';
        const thickness = active ? '3' : '2';

        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', fx); line.setAttribute('y1', fy);
        line.setAttribute('x2', tx); line.setAttribute('y2', ty);
        line.setAttribute('stroke', stroke);
        line.setAttribute('stroke-width', thickness);
        svg.appendChild(line);
      });
    });
  });
}

async function enterRoom(roomId) {
  if (state.pending_level_up) return;
  try {
    state = await postAction('/game/room/enter', { room_id: roomId });
    sessionStorage.setItem('mapScroll', document.querySelector('.floor-list').scrollTop);
    if (state.in_battle) {
      window.location.href = '/battle';
    } else {
      window.location.href = `/event?room=${encodeURIComponent(roomId)}`;
    }
  } catch (e) {
    alert(e.message);
  }
}

async function saveAndExit() {
  try {
    await postAction('/game/save');
    window.location.href = '/';
  } catch (e) {
    alert('Save failed: ' + e.message);
  }
}

// ---- Level-up overlay ----

function showLevelUpOverlay() {
  const pending   = state.pending_level_up;
  pendingBonuses  = { health: 0, attack: 0, defense: 0, magic: 0, mana: 0 };
  remainingPoints = pending.manual_points;
  undoStack       = [];
  renderOverlay();
  document.getElementById('levelup-overlay').classList.remove('hidden');
}

function renderOverlay() {
  document.getElementById('points-remaining').textContent = remainingPoints;
  const h = state.player;
  document.getElementById('overlay-stats').innerHTML = ['health','attack','defense','magic','mana'].map(stat => {
    const base  = (h[stat] || 0) + (h.level_bonuses?.[stat] || 0);
    const bonus = pendingBonuses[stat] || 0;
    const extra = bonus > 0 ? ` <span class="overlay-pending">(+${bonus})</span>` : '';
    return `<div class="overlay-stat-row">
      <span class="overlay-stat-name">${stat}</span>
      <span class="overlay-stat-value">${base}${extra}</span>
      <button class="btn btn-small btn-secondary" onclick="allocateStat('${stat}')" ${remainingPoints <= 0 ? 'disabled' : ''}>+</button>
    </div>`;
  }).join('');
}

function allocateStat(stat) {
  if (remainingPoints <= 0) return;
  pendingBonuses[stat]++;
  remainingPoints--;
  undoStack.push(stat);
  renderOverlay();
}

function undoAllocation() {
  if (!undoStack.length) return;
  pendingBonuses[undoStack.pop()]--;
  remainingPoints++;
  renderOverlay();
}

async function confirmAllocation(allocations) {
  if (allocations === undefined) allocations = { ...pendingBonuses };
  try {
    state = await postAction('/game/levelup', { allocations });
    document.getElementById('levelup-overlay').classList.add('hidden');
    renderPage();
  } catch (e) {
    alert(e.message);
  }
}
