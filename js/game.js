async function loadState() {
  const res = await fetch('/game/state');
  if (!res.ok) {
    window.location.href = '/';
    return null;
  }
  return res.json();
}

async function postAction(url, body) {
  const options = { method: 'POST' };
  if (body !== undefined) {
    options.headers = { 'Content-Type': 'application/json' };
    options.body = JSON.stringify(body);
  }
  const res = await fetch(url, options);
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

// Stats fields are flat (Stats is an anonymous embed in Entity).
// Entity-level fields have json tags (lowercase). Floor/Room do NOT, use capitals there.
function itemStatBonus(e, stat) {
  return (e.equipment || []).reduce((sum, item) => {
    if (item.item_type === 'consumable') return sum;
    return sum + (item.stat_affected?.[stat] || 0);
  }, 0);
}

function maxHP(e)   { return (e.health  || 0) + (e.level_stats?.health  || 0) + itemStatBonus(e, 'health'); }
function maxMana(e) { return (e.mana    || 0) + (e.level_stats?.mana    || 0) + itemStatBonus(e, 'mana'); }
function effAtk(e)  { return (e.attack  || 0) + (e.level_stats?.attack  || 0) + itemStatBonus(e, 'attack'); }
function effDef(e)  { return (e.defense || 0) + (e.level_stats?.defense || 0) + itemStatBonus(e, 'defense'); }
function effMag(e)  { return (e.magic   || 0) + (e.level_stats?.magic   || 0) + itemStatBonus(e, 'magic'); }

function itemTooltipHTML(item) {
  if (!item) return '';
  const bonuses = Object.entries(item.stat_affected || {})
    .map(([k, v]) => `${k}: ${v > 0 ? '+' : ''}${v}`).join(', ');
  return `<strong>${escHtml(item.name)}</strong><br>${escHtml(item.description || '')}`
    + (bonuses ? `<br><span class="tooltip-badge">${bonuses}</span>` : '')
    + `<br><span class="tooltip-badge">${item.item_type}</span>`;
}

function pct(val, max) {
  return max > 0 ? Math.max(0, Math.round(val / max * 100)) + '%' : '0%';
}

// move.effect is an array in the new model; cost_amount is on the move itself
function moveTooltipHTML(move) {
  if (!move) return '';
  const effects = (move.effect || []).map(e => {
    if (e.type === 'dot')      return `DoT: ${e.delta}/turn for ${e.duration}t`;
    if (e.type === 'stat_mod') return `${e.stat_affected} ${e.delta > 0 ? '+' : ''}${e.delta} for ${e.duration}t (${e.target})`;
    return '';
  }).filter(Boolean).join('<br>');
  const cost = (move.cost_amount || 0) > 0 ? `<br><span class="mana-cost">${move.cost_amount} MP</span>` : '';
  return `<strong>${escHtml(move.name)}</strong><br>${escHtml(move.description || '')}`
    + `<br><span class="tooltip-badge">${move.move_type}</span>`
    + ` <span class="tooltip-badge">${move.primary}</span>`
    + cost + (effects ? '<br>' + effects : '');
}

function initTooltips() {
  const tip = document.getElementById('tooltip');
  if (!tip) return;
  document.addEventListener('mouseover', e => {
    const el = e.target.closest('[data-tooltip]');
    if (!el) { tip.style.display = 'none'; return; }
    tip.innerHTML = el.dataset.tooltip;
    tip.style.display = 'block';
  });
  document.addEventListener('mousemove', e => {
    if (tip.style.display === 'none') return;
    tip.style.left = (e.clientX + 14) + 'px';
    tip.style.top  = (e.clientY - tip.offsetHeight - 14) + 'px';
  });
  document.addEventListener('mouseout', e => {
    if (e.target.closest('[data-tooltip]')) tip.style.display = 'none';
  });
}

function escHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

const _HERO_SPRITE_ALIAS = { mage: 'wizard' };
function heroSprite(id)       { return '/sprites/' + (_HERO_SPRITE_ALIAS[id] || id) + '.png'; }
function monsterSprite(id)    { return '/sprites/' + id + '.png'; }
function itemTypeSprite(type) { return '/sprites/' + type + '.png'; }

const _STAT_SPRITE = { attack: 'weapon', defense: 'armor', magic: 'magic', health: 'health', mana: 'mana' };
function statIcon(stat) {
  const src   = '/sprites/' + (_STAT_SPRITE[stat] || stat) + '.png';
  const label = stat.charAt(0).toUpperCase() + stat.slice(1);
  return `<img class="stat-icon" src="${src}" alt="${label}" data-tooltip="${label}" onerror="this.style.display='none'">`;
}
function itemTypeIcon(type, small) {
  const cls = small ? 'item-type-icon-sm' : 'item-type-icon';
  return `<img class="${cls}" src="${itemTypeSprite(type)}" alt="${type}" onerror="this.style.display='none'">`;
}
