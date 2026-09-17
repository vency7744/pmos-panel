<script lang="ts">
  let confirmAction = $state<'reboot' | 'shutdown' | null>(null);
  let message = $state('');

  async function execute(action: 'reboot' | 'shutdown') {
    message = `${action}ing system...`;
    const res = await fetch(`/api/power/${action}`, { method: 'POST' });
    const data = await res.json();
    message = data.status || data.error || 'done';
    confirmAction = null;
  }
</script>

<div class="power-page">
  <h2>Power Management</h2>

  {#if message}
    <div class="message">{message}</div>
  {/if}

  <div class="power-grid">
    <div class="power-card">
      <div class="power-icon reboot">↺</div>
      <h3>Reboot</h3>
      <p>Restart the system safely.</p>
      {#if confirmAction === 'reboot'}
        <div class="confirm">
          <span>Are you sure?</span>
          <button class="btn-confirm" onclick={() => execute('reboot')}>Yes, Reboot</button>
          <button class="btn-cancel" onclick={() => confirmAction = null}>Cancel</button>
        </div>
      {:else}
        <button class="btn-reboot" onclick={() => confirmAction = 'reboot'}>Reboot</button>
      {/if}
    </div>

    <div class="power-card">
      <div class="power-icon shutdown">⏻</div>
      <h3>Shutdown</h3>
      <p>Power off the system completely.</p>
      {#if confirmAction === 'shutdown'}
        <div class="confirm">
          <span>Are you sure?</span>
          <button class="btn-confirm" onclick={() => execute('shutdown')}>Yes, Shutdown</button>
          <button class="btn-cancel" onclick={() => confirmAction = null}>Cancel</button>
        </div>
      {:else}
        <button class="btn-shutdown" onclick={() => confirmAction = 'shutdown'}>Shutdown</button>
      {/if}
    </div>
  </div>
</div>

<style>
  .power-page { padding: 0; }
  h2 { margin: 0 0 1.5rem; font-size: 1rem; color: var(--text-h); }
  .message {
    padding: 0.75rem 1rem; background: var(--accent-bg); border: 1px solid var(--accent);
    border-radius: 6px; font-size: 0.875rem; color: var(--accent); margin-bottom: 1.5rem;
  }
  .power-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; max-width: 700px; }
  .power-card {
    background: var(--card); border: 1px solid var(--border); border-radius: 12px;
    padding: 1.5rem; text-align: center;
  }
  .power-icon {
    font-size: 2.5rem; margin-bottom: 0.75rem;
  }
  .power-icon.reboot { color: #eab308; }
  .power-icon.shutdown { color: var(--error); }
  .power-card h3 { margin: 0 0 0.5rem; font-size: 1rem; color: var(--text-h); }
  .power-card p { margin: 0 0 1.5rem; font-size: 0.8125rem; color: var(--text); }
  button {
    padding: 0.625rem 1.25rem; border: none; border-radius: 8px; font-size: 0.875rem;
    font-weight: 500; cursor: pointer; transition: opacity 0.2s;
  }
  button:hover { opacity: 0.9; }
  .btn-reboot { background: #eab308; color: #000; }
  .btn-shutdown { background: var(--error); color: #fff; }
  .confirm { display: flex; flex-direction: column; gap: 0.5rem; align-items: center; }
  .confirm span { color: var(--error); font-weight: 500; font-size: 0.875rem; }
  .btn-confirm { background: var(--error); color: #fff; }
  .btn-cancel { background: var(--border); color: var(--text); }
</style>
