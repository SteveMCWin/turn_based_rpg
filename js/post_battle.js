let state = null;

(async () => {
  state = await loadState();
  if (!state) return;

  const result  = state.last_battle_result;
  const content = document.getElementById('result-content');

  if (!result) {
    window.location.href = '/map';
    return;
  }

  if (result.player_won) {
    content.innerHTML = `<h1 class="win-title">Victory!</h1>
      <p>You defeated <strong>${escHtml(result.monster_name)}</strong>.</p>`;

    if (result.learned_move) {
      const lm  = result.learned_move;
      const def = state.moves?.[lm.move_id];
      const verb = lm.level === 1 ? 'Learned' : 'Leveled up';
      document.getElementById('learned-move').innerHTML = `
        <h3>${verb}: ${escHtml(def?.name || lm.move_id)}</h3>
        <p>${escHtml(def?.description || '')}</p>
        <span class="move-level-badge">Level ${lm.level}</span>`;
      document.getElementById('learned-move').classList.remove('hidden');
    }

    if (state.is_endless) {
      document.getElementById('btn-continue').onclick = () => { window.location.href = '/map'; };
    } else {
      const allDone = (state.floors || []).length > 0
        && state.floors.every(f => f.IsCompleted);
      if (allDone) {
        content.innerHTML = `<h1 class="win-title">You Win!</h1><p>All floors cleared!</p>`;
        document.getElementById('btn-continue').textContent = 'Play Again';
        document.getElementById('btn-continue').onclick = () => { window.location.href = '/'; };
      } else {
        document.getElementById('btn-continue').onclick = () => { window.location.href = '/map'; };
      }
    }
  } else {
    if (state.is_endless && result.floor_reached) {
      content.innerHTML = `<h1 class="lose-title">Fallen</h1>
        <p><strong>${escHtml(result.monster_name)}</strong> ended your run on floor ${result.floor_reached}.</p>`;
      document.getElementById('btn-continue').classList.add('hidden');
    } else {
      content.innerHTML = `<h1 class="lose-title">Defeated</h1>
        <p><strong>${escHtml(result.monster_name)}</strong> was too powerful.</p>`;
      document.getElementById('btn-continue').textContent = 'Back to Map';
      document.getElementById('btn-continue').onclick = () => { window.location.href = '/map'; };
    }
    document.getElementById('btn-restart').classList.remove('hidden');
    document.getElementById('btn-restart').onclick = () => { window.location.href = '/'; };
  }
})();
