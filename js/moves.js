let state = null;

(async () => {
  state = await loadState();
  if (!state) return;
  render();
  initTooltips();
})();

function render() {
  const hero     = state.player;
  const max      = state.settings?.max_equipped_moves || 4;
  const equipped = hero.equipped_moves || [];

  document.getElementById('equipped-count').textContent = `(${equipped.length}/${max})`;
  document.getElementById('slot-count').textContent = `${max - equipped.length} slot(s) available`;

  document.getElementById('equipped-moves').innerHTML = equipped.map(id => {
    const move = state.moves?.[id];
    const lm   = (hero.learned_moves || []).find(m => m.move_id === id);
    return `<div class="move-card equipped" data-tooltip="${escHtml(moveTooltipHTML(move))}">
      <div class="move-card-info">
        <strong>${escHtml(move?.name || id)}</strong>
        <span class="move-level">Lv.${lm?.level || 1}</span>
        <small class="move-type">${move?.move_type || ''} · ${move?.primary || ''}</small>
      </div>
      <button class="btn btn-small btn-unequip" onclick="unequip('${id}')">Unequip</button>
    </div>`;
  }).join('');

  document.getElementById('all-moves').innerHTML = (hero.learned_moves || [])
    .filter(lm => !equipped.includes(lm.move_id))
    .map(lm => {
      const move     = state.moves?.[lm.move_id];
      const canEquip = equipped.length < max;
      return `<div class="move-card" data-tooltip="${escHtml(moveTooltipHTML(move))}">
        <div class="move-card-info">
          <strong>${escHtml(move?.name || lm.move_id)}</strong>
          <span class="move-level">Lv.${lm.level}</span>
          <small class="move-type">${move?.move_type || ''} · ${move?.primary || ''}</small>
        </div>
        <button class="btn btn-small btn-equip" onclick="equip('${lm.move_id}')" ${canEquip ? '' : 'disabled'}>Equip</button>
      </div>`;
    }).join('');
}

let movesErrorTimer = null;
function showMovesError(msg) {
  const el = document.getElementById('moves-error');
  el.textContent = msg;
  el.classList.remove('hidden');
  clearTimeout(movesErrorTimer);
  movesErrorTimer = setTimeout(() => el.classList.add('hidden'), 3000);
}

async function equip(id) {
  try {
    state = await postAction('/game/moves/equip', { move_id: id });
    render();
  } catch (e) {
    showMovesError(e.message);
  }
}

async function unequip(id) {
  try {
    state = await postAction('/game/moves/unequip', { move_id: id });
    render();
  } catch (e) {
    showMovesError(e.message);
  }
}
