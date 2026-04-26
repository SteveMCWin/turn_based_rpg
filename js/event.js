let state = null;

(async () => {
  state = await loadState();
  if (!state) return;

  const roomId = new URLSearchParams(window.location.search).get('room');
  let room = null;
  for (const floor of (state.floors || []))
    for (const r of (floor.Rooms || []))
      if (r.ID === roomId) { room = r; break; }

  const ev = room?.Encounter?.event;
  const content = document.getElementById('event-content');

  if (!ev) {
    content.innerHTML = '<p>Nothing of note happened.</p>';
  } else {
    const sign   = ev.delta >= 0 ? '+' : '';
    const colour = ev.delta >= 0 ? 'var(--green)' : 'var(--accent)';
    content.innerHTML = `
      <p class="event-description">${escHtml(ev.description)}</p>
      <p class="event-delta" style="color:${colour}">
        ${escHtml(ev.stat_affected)}: ${sign}${ev.delta}
      </p>`;
  }
})();
