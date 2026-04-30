let state = null;

(async () => {
  state = await loadState();
  if (!state) return;
  render();
  initTooltips();
})();

function render() {
  if (state.in_battle) {
    document.getElementById('battle-warning').classList.remove('hidden');
  }

  const equipped      = state.player.equipment || [];
  const inventory     = state.player.item_pool  || [];
  const inBattle      = state.in_battle;
  const equippedTypes = new Set(equipped.map(e => e.item_type));

  // Equipped items
  const equippedEl    = document.getElementById('equipped-items');
  const equippedEmpty = document.getElementById('equipped-empty');
  if (equipped.length === 0) {
    equippedEl.innerHTML = '';
    equippedEmpty.classList.remove('hidden');
  } else {
    equippedEmpty.classList.add('hidden');
    equippedEl.innerHTML = equipped.map(item => `
      <div class="item-card equipped" data-tooltip="${escHtml(itemTooltipHTML(item))}">
        <div class="item-card-info">
          <strong>${escHtml(item.name)}</strong>
          <span class="item-type-tag">${itemIcon(item.item_type)} ${escHtml(item.item_type)}  ${formatStatBonuses(item)}</span>
          <small class="item-desc">${escHtml(item.description || '')}</small>
        </div>
        <button class="btn btn-small btn-unequip" onclick="unequipItem('${item.id}')" ${inBattle ? 'disabled' : ''}>Unequip</button>
      </div>`).join('');
  }

  // Inventory — group by id to show counts
  const counts = {};
  for (const id of inventory) counts[id] = (counts[id] || 0) + 1;

  const invEl    = document.getElementById('inventory-items');
  const invEmpty = document.getElementById('inventory-empty');
  const uniqueIds = Object.keys(counts);
  if (uniqueIds.length === 0) {
    invEl.innerHTML = '';
    invEmpty.classList.remove('hidden');
  } else {
    invEmpty.classList.add('hidden');
    invEl.innerHTML = uniqueIds.map(id => {
      const item = (state.items || {})[id];
      if (!item) return '';
      const isConsumable = item.item_type === 'consumable';
      const isEquipped   = equipped.some(e => e.id === id);
      const count        = counts[id] > 1 ? ` ×${counts[id]}` : '';

      const slotFull = !isConsumable && !isEquipped && equippedTypes.has(item.item_type);
      let btn;
      if (isConsumable) {
        btn = `<button class="btn btn-small btn-primary" onclick="useItem('${id}')" ${inBattle ? 'disabled' : ''}>Use</button>`;
      } else if (isEquipped) {
        btn = `<span class="tag-equipped">Equipped</span>`;
      } else {
        btn = `<button class="btn btn-small btn-equip" onclick="equipItem('${id}')" ${inBattle || slotFull ? 'disabled' : ''}>Equip</button>`;
      }

      return `<div class="item-card" data-tooltip="${escHtml(itemTooltipHTML(item))}">
        <div class="item-card-info">
          <strong>${escHtml(item.name)}${escHtml(count)}</strong>
          <span class="item-type-tag">${itemIcon(item.item_type)} ${escHtml(item.item_type)}  ${formatStatBonuses(item)}</span>
          <small class="item-desc">${escHtml(item.description || '')}</small>
        </div>
        ${btn}
      </div>`;
    }).join('');
  }
}

function itemIcon(type) { return itemTypeIcon(type); }

function formatStatBonuses(item) {
  const bonuses = Object.entries(item.stat_affected || {})
    .map(([k, v]) => `${k} ${v > 0 ? '+' : ''}${v}`)
    .join(', ');
  return bonuses ? `· ${bonuses}` : '';
}

async function equipItem(id) {
  try {
    state = await postAction('/game/items/equip', { item_id: id });
    render();
  } catch (e) {
    showError(e.message);
  }
}

async function unequipItem(id) {
  try {
    state = await postAction('/game/items/unequip', { item_id: id });
    render();
  } catch (e) {
    showError(e.message);
  }
}

async function useItem(id) {
  try {
    state = await postAction('/game/items/use', { item_id: id });
    render();
  } catch (e) {
    showError(e.message);
  }
}

function showError(msg) {
  const el = document.getElementById('items-error');
  el.textContent = msg;
  el.classList.remove('hidden');
}
