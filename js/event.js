let state = null;

(async () => {
  state = await loadState();
  if (!state) return;

  const roomId = new URLSearchParams(window.location.search).get('room');
  let room = null;
  for (const floor of (state.floors || []))
    for (const r of (floor.Rooms || []))
      if (r.Id === roomId) { room = r; break; }

  const ev = room?.Encounter?.event;
  const content = document.getElementById('event-content');

  if (!ev) {
    content.innerHTML = '<p>Nothing of note happened.</p>';
  } else {
    const parts = Object.entries(ev.stat_affected || {}).map(([stat, val]) => {
      const sign   = val >= 0 ? '+' : '';
      const colour = val >= 0 ? 'var(--green)' : 'var(--accent)';
      return `<span style="color:${colour}">${sign}${val} ${escHtml(stat)}</span>`;
    });
    content.innerHTML = `
      <p class="event-description">${escHtml(ev.description)}</p>
      <p class="event-delta">${parts.join(' &middot; ')}</p>`;
  }
})();
