async function loadState() {
  const res = await fetch('/game/state');
  if (!res.ok) {
    window.location.href = '/';
    return null;
  }
  return res.json();
}

async function postAction(url, body = {}) {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

// Stats fields are flat (Stats is an anonymous embed in Entity).
// Entity-level fields have json tags (lowercase). Floor/Room do NOT — use capitals there.
function maxHP(e)   { return (e.health   || 0) + (e.level_bonuses?.health   || 0); }
function maxMana(e) { return (e.mana     || 0) + (e.level_bonuses?.mana     || 0); }
function effAtk(e)  { return (e.attack   || 0) + (e.level_bonuses?.attack   || 0); }
function effDef(e)  { return (e.defense  || 0) + (e.level_bonuses?.defense  || 0); }
function effMag(e)  { return (e.magic    || 0) + (e.level_bonuses?.magic    || 0); }

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
    tip.style.top  = (e.clientY + 14) + 'px';
  });
  document.addEventListener('mouseout', e => {
    if (e.target.closest('[data-tooltip]')) tip.style.display = 'none';
  });
}

function escHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
