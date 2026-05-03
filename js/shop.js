function openShop() {
  if (state?.in_battle) return;
  renderShop();
  document.getElementById('shop-overlay').classList.remove('hidden');
}

function closeShop() {
  document.getElementById('shop-overlay').classList.add('hidden');
}

function closeShopOnBackdrop(e) {
  if (e.target === document.getElementById('shop-overlay')) closeShop();
}

function renderShop() {
  const gold = state?.player?.current_gold || 0;
  document.getElementById('shop-gold-display').innerHTML = `<img class="stat-icon" src="/sprites/gold.png" alt="Gold" data-tooltip="Gold" onerror="this.style.display='none'"> ${gold}`;
  document.getElementById('shop-error').classList.add('hidden');

  const shopItems = state?.shop?.items || [];
  document.getElementById('shop-stock').innerHTML = shopItems.length
    ? shopItems.map(item => shopItemCardHTML(item, gold)).join('')
    : '<p class="muted-text">Sold out.</p>';

  const inventory = state?.player?.item_pool || [];
  if (!inventory.length) {
    document.getElementById('shop-inventory').innerHTML = '<p class="muted-text">Nothing to sell.</p>';
    return;
  }
  // Count duplicates
  const counts = {};
  for (const id of inventory) counts[id] = (counts[id] || 0) + 1;
  const unique = [...new Set(inventory)];
  document.getElementById('shop-inventory').innerHTML = unique.map(id => {
    const item = state?.items?.[id];
    if (!item) return '';
    const sellPrice = Math.floor(item.price * (state?.settings?.sell_modifier || 0.6));
    const count = counts[id] > 1 ? ` ×${counts[id]}` : '';
    return `<div class="shop-item-card">
      <div class="shop-item-info">
        <span class="shop-item-name">${escHtml(item.name)}${count}</span>
        <span class="shop-item-desc">${escHtml(item.description)}</span>
      </div>
      <button class="btn btn-small btn-secondary" onclick="sellItem('${escHtml(item.id)}')">
        Sell <span class="shop-price sell-price">${sellPrice}g</span>
      </button>
    </div>`;
  }).join('');
}

function shopItemCardHTML(item, playerGold) {
  const canAfford = playerGold >= item.price;
  return `<div class="shop-item-card ${canAfford ? '' : 'shop-item-unaffordable'}">
    <div class="shop-item-info">
      <span class="shop-item-name">${itemTypeIcon(item.item_type, true)} ${escHtml(item.name)}</span>
      <span class="shop-item-desc">${escHtml(item.description)}</span>
    </div>
    <button class="btn btn-small btn-primary" onclick="buyItem('${escHtml(item.id)}')" ${canAfford ? '' : 'disabled'}>
      Buy <span class="shop-price">${item.price}g</span>
    </button>
  </div>`;
}

async function buyItem(itemId) {
  try {
    state = await postAction('/game/shop/buy', { item_id: itemId });
    renderHeroPanel();
    renderShop();
  } catch (e) {
    showShopError(e.message);
  }
}

async function sellItem(itemId) {
  try {
    state = await postAction('/game/shop/sell', { item_id: itemId });
    renderHeroPanel();
    renderShop();
  } catch (e) {
    showShopError(e.message);
  }
}

function showShopError(msg) {
  const el = document.getElementById('shop-error');
  el.textContent = msg;
  el.classList.remove('hidden');
}
