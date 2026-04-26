let state = null;

(async () => {
  state = await loadState();
  if (!state) return;
  renderAll();
  initTooltips();
})();

function renderAll() {
  const monster = currentMonster();
  renderCombatant('hero', state.player);
  if (monster) renderCombatant('monster', monster);
  renderMonsterInfo(monster);
  renderMoveButtons();
  renderBattleLog();
}

function currentMonster() {
  if (!state.current_room_id || !state.floors) return null;
  for (const floor of state.floors)
    for (const room of floor.Rooms)       // capital R
      if (room.ID === state.current_room_id)  // capital ID
        return room.Encounter?.monster || null;  // capital Encounter, .monster lowercase
  return null;
}

function renderCombatant(side, entity) {
  const hp  = entity.current_hp;
  const mp  = entity.current_mana || 0;
  const mhp = maxHP(entity);
  const mmp = maxMana(entity);

  document.getElementById(`${side}-name`).textContent    = entity.name || 'Knight';
  document.getElementById(`${side}-hp-text`).textContent = `HP: ${hp} / ${mhp}`;
  document.getElementById(`${side}-hp-bar`).style.width  = pct(hp, mhp);

  const manaWrap = document.getElementById(`${side}-mana-wrap`);
  const manaText = document.getElementById(`${side}-mana-text`);
  if (mmp > 0) {
    manaWrap.style.display = '';
    manaText.textContent   = `MP: ${mp} / ${mmp}`;
    document.getElementById(`${side}-mana-bar`).style.width = pct(mp, mmp);
  } else {
    manaWrap.style.display = 'none';
    manaText.textContent   = '';
  }

  // TurnsRemaining has no json tag — capital T
  document.getElementById(`${side}-effects`).innerHTML = (entity.status_effects || []).map(se => {
    const label = se.type === 'stat_mod'
      ? `${se.stat_affected} ${se.delta > 0 ? '+' : ''}${se.delta}`
      : `DoT ${se.delta}`;
    return `<span class="effect-tag">${label} (${se.TurnsRemaining}t)</span>`;
  }).join('');
}

function renderMonsterInfo(monster) {
  if (!monster) return;
  document.getElementById('monster-stats').innerHTML =
    `<span>ATK ${effAtk(monster)}</span>`
    + `<span>DEF ${effDef(monster)}</span>`
    + `<span>MAG ${effMag(monster)}</span>`
    + `<span>Lv.${monster.level}</span>`;

  // monster.learned_moves is []LearnedMove with {move_id, level}
  document.getElementById('monster-skills').innerHTML = (monster.learned_moves || []).map(lm => {
    const move = state.moves?.[lm.move_id];
    return `<span class="skill-tag" data-tooltip="${escHtml(moveTooltipHTML(move))}">${escHtml(move?.name || lm.move_id)}</span>`;
  }).join('');
}

function renderMoveButtons() {
  const grid = document.getElementById('move-buttons');
  if (!state.in_battle) { grid.innerHTML = ''; return; }

  const bonusPct = state.settings?.move_level_bonus_percent || 0;
  grid.innerHTML = (state.player.equipped_moves || []).map(id => {
    const move     = state.moves?.[id];
    const lm       = (state.player.learned_moves || []).find(m => m.move_id === id);
    const lvl      = lm?.level || 1;
    const effVal   = Math.floor((move?.base_value || 0) * (1 + (lvl - 1) * bonusPct / 100));
    const cost     = move?.cost_amount || 0;
    const cantAfford = cost > 0 && (state.player.current_mana || 0) < cost;
    const costBadge  = cost > 0 ? `<br><small class="mana-cost">${cost} MP</small>` : '';
    return `<button class="btn btn-move"
      onclick="submitMove('${id}')"
      data-tooltip="${escHtml(moveTooltipHTML(move))}"
      ${cantAfford ? 'disabled' : ''}>
      ${escHtml(move?.name || id)}<br><small>Lv.${lvl} (${effVal})</small>${costBadge}
    </button>`;
  }).join('');
}

function renderBattleLog() {
  const el = document.getElementById('battle-log');
  el.innerHTML = '';
  // battle_log is []string in the new model — plain strings, no class
  for (const line of (state.battle_log || [])) {
    const p = document.createElement('p');
    p.textContent = line;
    el.appendChild(p);
  }
  el.scrollTop = el.scrollHeight;
}

async function submitMove(moveId) {
  document.getElementById('move-buttons').innerHTML = '<p class="waiting-text">Processing…</p>';
  try {
    const result = await postAction('/game/battle/move', { move_id: moveId });
    state = result.game_state;
    renderAll();
    if (result.battle_over) {
      document.getElementById('move-buttons').innerHTML =
        `<button class="btn btn-primary" onclick="window.location.href='/post-battle'">Continue</button>`;
    }
  } catch (e) {
    alert(e.message);
    renderMoveButtons();
  }
}
