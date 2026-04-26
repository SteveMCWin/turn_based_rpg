let state = null;
let pendingBonuses = {};
let remainingPoints = 0;
let undoStack = [];

(async () => {
  state = await loadState();
  if (!state) return;
  renderPage();
  initTooltips();
})();

function renderPage() {
  renderHeroPanel();
  renderFloors();
  if (state.pending_level_up) showLevelUpOverlay();
}

function renderHeroPanel() {
  const h = state.player;
  document.getElementById('hero-name').textContent      = h.name;
  document.getElementById('hero-hp-text').textContent   = `HP: ${h.current_hp} / ${maxHP(h)}`;
  document.getElementById('hero-mana-text').textContent = `MP: ${h.current_mana || 0} / ${maxMana(h)}`;
  document.getElementById('hero-hp-bar').style.width    = pct(h.current_hp, maxHP(h));
  document.getElementById('hero-mana-bar').style.width  = pct(h.current_mana || 0, maxMana(h));
  document.getElementById('hero-stats').innerHTML = `
    <span>LVL ${h.level}</span>
    <span>ATK ${effAtk(h)}</span>
    <span>DEF ${effDef(h)}</span>
    <span>MAG ${effMag(h)}</span>
    <span class="xp-text">XP ${h.current_xp || 0}</span>
  `;
  document.getElementById('equipped-list').innerHTML = (h.equipped_moves || []).map(id => {
    const move = state.moves?.[id];
    const lm   = (h.learned_moves || []).find(m => m.move_id === id);
    return `<li data-tooltip="${escHtml(moveTooltipHTML(move))}">${escHtml(move?.name || id)} <span class="move-level">Lv.${lm?.level || 1}</span></li>`;
  }).join('');
}

function renderFloors() {
  // Floor fields: Rooms (capital), IsCompleted (capital)
  // Room fields:  ID (capital), Encounter (capital), IsCompleted (capital), CanEnter (capital)
  // Encounter fields have json tags: kind, monster, event (lowercase)
  document.getElementById('floors').innerHTML = (state.floors || []).map((floor, fi) => {
    const rows = (floor.Rooms || []).map(room => renderRoomRow(room, floor)).join('');
    return `
      <div class="floor-block ${floor.IsCompleted ? 'floor-done' : ''}">
        <div class="floor-label">Floor ${fi + 1}${floor.IsCompleted ? ' ✓' : ''}</div>
        <div class="floor-rooms">${rows}</div>
      </div>`;
  }).join('');
}

function renderRoomRow(room, floor) {
  const enc     = room.Encounter;
  const kind    = enc?.kind;
  const monster = enc?.monster;
  const event   = enc?.event;

  // Disable sibling buttons if any room on this floor is completed
  const floorHasCompleted = (floor.Rooms || []).some(r => r.IsCompleted);

  let icon, label, action;

  if (kind === 'monster' && monster) {
    icon  = '⚔';
    label = `${escHtml(monster.name)} — Lv.${monster.level} (${monster.current_hp}/${maxHP(monster)} HP)`;
  } else {
    icon  = '✦';
    label = room.IsCompleted && event ? escHtml(event.description) : 'Unknown event';
  }

  if (room.IsCompleted && kind === 'monster' && room.CanEnter && !state.pending_level_up) {
    action = `<button class="btn btn-secondary btn-small" onclick="enterRoom('${room.ID}')">Rematch</button>`;
  } else if (room.IsCompleted) {
    action = `<span class="tag-done">✓</span>`;
  } else if (room.CanEnter && !state.pending_level_up) {
    if (kind === 'monster') {
      action = `<button class="btn btn-primary btn-small" onclick="enterRoom('${room.ID}')">Fight</button>`;
    } else {
      action = `<button class="btn btn-event btn-small" onclick="enterRoom('${room.ID}')">?</button>`;
    }
  } else {
    // locked: either not can_enter, or sibling was completed
    const reason = floorHasCompleted && !room.IsCompleted ? 'disabled' : 'locked';
    action = `<span class="tag-locked tag-${reason}">—</span>`;
  }

  const cls = room.IsCompleted ? 'room-row done' : (room.CanEnter ? 'room-row' : 'room-row locked');
  return `
    <div class="${cls}">
      <span class="room-icon ${kind}">${icon}</span>
      <span class="room-label">${label}</span>
      <span class="room-action">${action}</span>
    </div>`;
}

async function enterRoom(roomId) {
  if (state.pending_level_up) return;
  try {
    state = await postAction('/game/room/enter', { room_id: roomId });
    if (state.in_battle) {
      window.location.href = '/battle';
    } else {
      // Event was applied — go to event result page
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
