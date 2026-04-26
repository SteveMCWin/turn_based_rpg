document.getElementById('btn-new-run').addEventListener('click', async () => {
  const btn = document.getElementById('btn-new-run');
  const err = document.getElementById('menu-error');
  btn.disabled = true;
  btn.textContent = 'Loading...';
  err.classList.add('hidden');
  try {
    await postAction('/game/new');
    window.location.href = '/map';
  } catch (e) {
    err.textContent = e.message;
    err.classList.remove('hidden');
    btn.disabled = false;
    btn.textContent = 'Start New Run';
  }
});

(async () => {
  try {
    const res = await fetch('/game/saves');
    if (!res.ok) return;
    const saves = await res.json();
    if (!saves || saves.length === 0) return;
    document.getElementById('saves-section').classList.remove('hidden');
    document.getElementById('saves-list').innerHTML = saves.map(s => `
      <div class="save-slot">
        <span class="save-label">Save #${s.id}</span>
        <span class="save-date">${s.saved_at}</span>
        <button class="btn btn-small btn-secondary" onclick="loadSave(${s.id})">Load</button>
        <button class="btn btn-small btn-secondary" onclick="deleteSave(${s.id}, this)">Delete</button>
      </div>`).join('');
  } catch (_) {}
})();

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
