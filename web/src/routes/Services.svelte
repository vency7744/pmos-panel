<script lang="ts">
  import { onMount } from 'svelte';

  interface Service {
    name: string;
    active: string;
    loaded: string;
    status: string;
  }

  let services = $state<Service[]>([]);
  let search = $state('');
  let loading = $state(false);
  let actionMessage = $state('');

  onMount(() => loadServices());

  async function loadServices() {
    loading = true;
    const res = await fetch('/api/services');
    services = await res.json();
    loading = false;
  }

  async function serviceAction(name: string, action: string) {
    actionMessage = `${action}ing ${name}...`;
    const res = await fetch(`/api/services/${name}/${action}`, { method: 'POST' });
    const data = await res.json();
    actionMessage = data.status || data.error || 'done';
    setTimeout(() => actionMessage = '', 3000);
    await loadServices();
  }

  function filteredServices() {
    if (!search) return services;
    return services.filter(s => s.name.toLowerCase().includes(search.toLowerCase()));
  }

  function statusColor(status: string) {
    if (status === 'active') return 'var(--green)';
    if (status === 'failed') return 'var(--error)';
    if (status === 'activating') return '#eab308';
    return 'var(--text)';
  }
</script>

<div class="services-page">
  <div class="toolbar">
    <input type="text" placeholder="Search services..." bind:value={search} />
    <button onclick={loadServices} disabled={loading}>Refresh</button>
  </div>

  {#if actionMessage}
    <div class="action-msg">{actionMessage}</div>
  {/if}

  <div class="table-wrapper">
    <table>
      <thead>
        <tr>
          <th>Service</th>
          <th>Status</th>
          <th>Loaded</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each filteredServices() as svc}
          <tr>
            <td class="svc-name">{svc.name}</td>
            <td>
              <span class="status-badge" style="color:{statusColor(svc.active)}">
                {svc.active}
              </span>
            </td>
            <td class="loaded">{svc.loaded}</td>
            <td class="actions">
              {#if svc.active === 'active'}
                <button class="btn-stop" onclick={() => serviceAction(svc.name, 'stop')}>Stop</button>
                <button class="btn-restart" onclick={() => serviceAction(svc.name, 'restart')}>Restart</button>
              {:else}
                <button class="btn-start" onclick={() => serviceAction(svc.name, 'start')}>Start</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<style>
  .services-page { padding: 0; }
  .toolbar {
    display: flex; gap: 0.75rem; margin-bottom: 1rem;
  }
  input {
    flex: 1; padding: 0.5rem 0.75rem; border: 1px solid var(--border);
    border-radius: 6px; background: var(--input-bg); color: var(--text-h);
    font-size: 0.875rem; outline: none;
  }
  input:focus { border-color: var(--accent); }
  button {
    padding: 0.5rem 0.75rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--card); color: var(--text); font-size: 0.8125rem; cursor: pointer;
  }
  button:hover { background: var(--hover); }
  .action-msg {
    padding: 0.5rem 0.75rem; background: var(--accent-bg); border: 1px solid var(--accent);
    border-radius: 6px; font-size: 0.8125rem; color: var(--accent); margin-bottom: 1rem;
  }
  .table-wrapper { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; }
  th {
    text-align: left; padding: 0.625rem 0.75rem; font-size: 0.75rem;
    color: var(--text); text-transform: uppercase; letter-spacing: 0.5px;
    border-bottom: 1px solid var(--border);
  }
  td {
    padding: 0.625rem 0.75rem; border-bottom: 1px solid var(--border);
    font-size: 0.8125rem; color: var(--text);
  }
  tr:hover { background: var(--hover); }
  .svc-name { color: var(--text-h); font-family: var(--mono); }
  .loaded { font-size: 0.75rem; }
  .actions { display: flex; gap: 0.375rem; }
  .btn-start { color: var(--green); border-color: var(--green); }
  .btn-stop { color: var(--error); border-color: var(--error); }
  .btn-restart { color: #eab308; border-color: #eab308; }
</style>
