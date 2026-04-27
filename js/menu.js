let selectedHeroId = null;

(async () => {
  const res = await fetch('/game/heroes');
  if (!res.ok) return;
  const heroes = await res.json();
  renderHeroCards(heroes);
})();

const STAT_ABBR = { attack: 'ATK', defense: 'DEF', magic: 'MAG', health: 'HP', mana: 'MP' };

function envPrefsHTML(envEffects) {
  if (!envEffects || !Object.keys(envEffects).length) return '';
  const buffs = [], debuffs = [];
  for (const [envId, eff] of Object.entries(envEffects)) {
    const stat  = STAT_ABBR[eff.stat_affected] || eff.stat_affected;
    const sign  = eff.delta > 0 ? '+' : '';
    const name  = envId.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
    const entry = `${name} (${sign}${eff.delta} ${stat})`;
    if (eff.delta > 0) buffs.push(entry);
    else debuffs.push(entry);
  }
  let html = '';
  if (buffs.length)   html += `<div class="env-pref env-buff">▲ ${buffs.join(' · ')}</div>`;
  if (debuffs.length) html += `<div class="env-pref env-debuff">▼ ${debuffs.join(' · ')}</div>`;
  return html ? `<div class="hero-card-env">${html}</div>` : '';
}

function renderHeroCards(heroes) {
  const grid = document.getElementById('hero-select-grid');
  grid.innerHTML = heroes.map(h => {
    const startingItems = (h.equipment || []).map(item => {
      const icon = item.item_type === 'weapon' ? '⚔' : item.item_type === 'armor' ? '🛡' : '💎';
      return `<span class="hero-card-item" title="${escHtml(item.name)}">${icon} ${escHtml(item.name)}</span>`;
    }).join('');

    return `
    <div class="hero-card" data-id="${escHtml(h.id)}" onclick="selectHero(this)">
      <div class="hero-card-name">${escHtml(h.name)}</div>
      <div class="hero-card-desc">${escHtml(h.description || '')}</div>
      <div class="hero-card-stats">
        <span title="Health">❤ ${h.health}</span>
        <span title="Attack">⚔ ${h.attack}</span>
        <span title="Defense">🛡 ${h.defense}</span>
        <span title="Magic">✦ ${h.magic}</span>
        <span title="Mana">◈ ${h.mana}</span>
      </div>
      ${startingItems ? `<div class="hero-card-items">${startingItems}</div>` : ''}
      ${envPrefsHTML(h.env_effects)}
    </div>`;
  }).join('');

  const first = grid.firstElementChild;
  if (first) selectHero(first);
}

function selectHero(el) {
  document.querySelectorAll('.hero-card').forEach(c => c.classList.remove('selected'));
  el.classList.add('selected');
  selectedHeroId = el.dataset.id;
  document.getElementById('btn-new-run').disabled = false;
}

document.getElementById('btn-new-run').addEventListener('click', async () => {
  if (!selectedHeroId) return;
  const btn = document.getElementById('btn-new-run');
  const err = document.getElementById('menu-error');
  btn.disabled = true;
  btn.textContent = 'Loading...';
  err.classList.add('hidden');
  try {
    await postAction('/game/new', { hero_id: selectedHeroId });
    window.location.href = '/map';
  } catch (e) {
    err.textContent = e.message;
    err.classList.remove('hidden');
    btn.disabled = false;
    btn.textContent = 'Start Run';
  }
});

// ---- Settings modal ----

function openSettings() {
  document.getElementById('settings-overlay').classList.remove('hidden');
  document.getElementById('settings-main').classList.remove('hidden');
  document.getElementById('settings-saves').classList.add('hidden');
  document.getElementById('modal-title').textContent = 'Settings';
}

function closeSettings() {
  document.getElementById('settings-overlay').classList.add('hidden');
}

function closeSettingsOnBackdrop(e) {
  if (e.target === document.getElementById('settings-overlay')) closeSettings();
}

async function showSaves() {
  document.getElementById('settings-main').classList.add('hidden');
  document.getElementById('settings-saves').classList.remove('hidden');
  document.getElementById('modal-title').textContent = 'Saved Runs';

  const list  = document.getElementById('saves-list');
  const empty = document.getElementById('saves-empty');
  list.innerHTML = '<p class="muted-text">Loading...</p>';
  empty.classList.add('hidden');

  try {
    const res = await fetch('/game/saves');
    if (!res.ok) throw new Error();
    const saves = await res.json();
    list.innerHTML = '';
    if (!saves || saves.length === 0) {
      empty.classList.remove('hidden');
      return;
    }
    list.innerHTML = saves.map(s => `
      <div class="save-slot">
        <span class="save-label">Save #${s.id}</span>
        <span class="save-date">${s.saved_at}</span>
        <button class="btn btn-small btn-secondary" onclick="loadSave(${s.id})">Load</button>
        <button class="btn btn-small btn-secondary" onclick="deleteSave(${s.id}, this)">Delete</button>
      </div>`).join('');
  } catch (_) {
    list.innerHTML = '<p class="muted-text">Failed to load saves.</p>';
  }
}

function backToSettings() {
  document.getElementById('settings-main').classList.remove('hidden');
  document.getElementById('settings-saves').classList.add('hidden');
  document.getElementById('modal-title').textContent = 'Settings';
}

async function loadSave(id) {
  try {
    await postAction(`/game/saves/${id}/load`);
    window.location.href = '/map';
  } catch (e) {
    alert('Load failed: ' + e.message);
  }
}

async function deleteSave(id, btn) {
  try {
    const res = await fetch(`/game/saves/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Delete failed');
    btn.closest('.save-slot').remove();
  } catch (e) {
    alert(e.message);
  }
}
