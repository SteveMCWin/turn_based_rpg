let state = null;
let battleLog = JSON.parse(sessionStorage.getItem('battleInitialLog') || '[]');
sessionStorage.removeItem('battleInitialLog');

(async () => {
  state = await loadState();
  if (!state) return;
  renderAll();
  initTooltips();
  if (state.waiting_for_monster) setTimeout(resolveMonsterTurn, 500);
})();

function renderAll() {
  const monster = currentMonster();
  renderCombatant('hero', state.player);
  if (monster) renderCombatant('monster', monster);
  renderMonsterInfo(monster);
  renderEnvironment();
  renderMoveButtons();
  renderBattleLog();
}

function renderEnvironment() {
  const el = document.getElementById('env-banner');
  if (!el) return;
  const env = currentRoom()?.Environment;
  if (!env?.Name) return;
  el.innerHTML = `<span class="env-banner-name">${escHtml(env.Name)}</span><span class="env-banner-desc">${escHtml(env.Description)}</span>`;
}

function currentRoom() {
  if (!state.current_room_id || !state.floors) return null;
  for (const floor of state.floors)
    for (const room of floor.Rooms)
      if (room.Id === state.current_room_id) return room;
  return null;
}

function currentMonster() {
  return currentRoom()?.Encounter?.monster || null;
}

function renderCombatant(side, entity) {
  const hp  = entity.current_hp;
  const mp  = entity.current_mana || 0;
  const mhp = maxHP(entity);
  const mmp = maxMana(entity);

  const spriteEl = document.getElementById(`${side}-sprite`);
  if (spriteEl) spriteEl.src = side === 'hero' ? heroSprite(entity.id) : monsterSprite(entity.id);

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

  document.getElementById(`${side}-items`).innerHTML = (entity.equipment || []).map(item =>
    `<span data-tooltip="${escHtml(itemTooltipHTML(item))}">${itemTypeIcon(item.item_type)}</span>`
  ).join('');
}

function renderMonsterInfo(monster) {
  if (!monster) return;
  document.getElementById('monster-stats').innerHTML =
    `<span>Lv.${monster.level}</span>`
    + `<span>${statIcon('attack')}ATK ${effAtk(monster)}</span>`
    + `<span>${statIcon('defense')}DEF ${effDef(monster)}</span>`
    + `<span>${statIcon('magic')}MAG ${effMag(monster)}</span>`
    + `<span>${statIcon('health')}HP ${monster.current_hp}/${maxHP(monster)}</span>`
    + (maxMana(monster) > 0 ? `<span>${statIcon('mana')}MP ${monster.current_mana}/${maxMana(monster)}</span>` : '');

  // monster.learned_moves is []LearnedMove with {move_id, level}
  document.getElementById('monster-skills').innerHTML = (monster.learned_moves || []).map(lm => {
    const move = state.moves?.[lm.move_id];
    return `<span class="skill-tag" data-tooltip="${escHtml(moveTooltipHTML(move))}">${escHtml(move?.name || lm.move_id)}</span>`;
  }).join('');
}

function renderMoveButtons() {
  const grid = document.getElementById('move-buttons');
  if (!state.in_battle || state.waiting_for_monster) { grid.innerHTML = ''; return; }

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
  for (const line of battleLog) {
    const p = document.createElement('p');
    p.textContent = line;
    el.appendChild(p);
  }
  el.scrollTop = el.scrollHeight;
}

async function submitMove(moveId) {
  try {
    const result = await postAction('/game/battle/move', { move_id: moveId });
    if (result.new_log_lines?.length) battleLog.push(...result.new_log_lines);
    state = result.game_state;
    renderAll();
    if (result.battle_over) {
      showContinueButton();
      return;
    }
    setTimeout(resolveMonsterTurn, 1000);
  } catch (e) {
    alert(e.message);
    renderMoveButtons();
  }
}

async function resolveMonsterTurn() {
  try {
    const result = await postAction('/game/battle/monster-move');
    if (result.new_log_lines?.length) battleLog.push(...result.new_log_lines);
    state = result.game_state;
    renderAll();
    if (result.battle_over) showContinueButton();
  } catch (e) {
    alert(e.message);
  }
}

function showContinueButton() {
  document.getElementById('move-buttons').innerHTML =
    `<button class="btn btn-primary" onclick="window.location.href='/post-battle'">Continue</button>`;
}
