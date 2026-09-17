<script lang="ts">
  import { onMount } from 'svelte';

  interface Proc {
    pid: number;
    user: string;
    cpu: number;
    mem: number;
    command: string;
  }

  let processes = $state<Proc[]>([]);
  let search = $state('');
  let sortBy = $state<'cpu' | 'mem' | 'pid'>('cpu');
  let loading = $state(false);
  let confirmKill = $state<number | null>(null);

  onMount(() => loadProcesses());

  async function loadProcesses() {
    loading = true;
    const res = await fetch('/api/processes');
    processes = await res.json();
    loading = false;
  }

  async function killProcess(pid: number) {
    await fetch(`/api/processes/${pid}/kill`, { method: 'POST' });
    confirmKill = null;
    await loadProcesses();
  }

  function filteredProcesses() {
    let filtered = processes;
    if (search) {
      const q = search.toLowerCase();
      filtered = filtered.filter(p =>
        p.command.toLowerCase().includes(q) ||
        p.user.toLowerCase().includes(q) ||
        String(p.pid).includes(q)
      );
    }
    return filtered.sort((a, b) => {
      if (sortBy === 'cpu') return b.cpu - a.cpu;
      if (sortBy === 'mem') return b.mem - a.mem;
      return a.pid - b.pid;
    });
  }
</script>

<div class="processes-page">
  <div class="toolbar">
    <input type="text" placeholder="Search processes..." bind:value={search} />
    <div class="sort-btns">
      <button class:active={sortBy === 'cpu'} onclick={() => sortBy = 'cpu'}>CPU</button>
      <button class:active={sortBy === 'mem'} onclick={() => sortBy = 'mem'}>RAM</button>
      <button class:active={sortBy === 'pid'} onclick={() => sortBy = 'pid'}>PID</button>
    </div>
    <button onclick={loadProcesses} disabled={loading}>Refresh</button>
  </div>

  <div class="table-wrapper">
    <table>
      <thead>
        <tr>
          <th>PID</th>
          <th>User</th>
          <th>CPU%</th>
          <th>RAM (MB)</th>
          <th>Command</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each filteredProcesses() as proc}
          <tr>
            <td class="mono">{proc.pid}</td>
            <td>{proc.user}</td>
            <td class="mono">{proc.cpu.toFixed(1)}</td>
            <td class="mono">{proc.mem.toFixed(1)}</td>
            <td class="cmd">{proc.command}</td>
            <td>
              {#if confirmKill === proc.pid}
                <div class="confirm-kill">
                  <span>Kill?</span>
                  <button class="btn-yes" onclick={() => killProcess(proc.pid)}>Y</button>
                  <button class="btn-no" onclick={() => confirmKill = null}>N</button>
                </div>
              {:else}
                <button class="btn-kill" onclick={() => confirmKill = proc.pid}>Kill</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<style>
  .processes-page { padding: 0; }
  .toolbar { display: flex; gap: 0.75rem; margin-bottom: 1rem; align-items: center; }
  .sort-btns { display: flex; gap: 0.25rem; }
  .sort-btns button.active { background: var(--accent-bg); color: var(--accent); border-color: var(--accent); }
  input {
    flex: 1; padding: 0.5rem 0.75rem; border: 1px solid var(--border);
    border-radius: 6px; background: var(--input-bg); color: var(--text-h); font-size: 0.875rem; outline: none;
  }
  input:focus { border-color: var(--accent); }
  button {
    padding: 0.5rem 0.75rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--card); color: var(--text); font-size: 0.8125rem; cursor: pointer;
  }
  button:hover { background: var(--hover); }
  .table-wrapper { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; }
  th {
    text-align: left; padding: 0.5rem 0.75rem; font-size: 0.75rem; color: var(--text);
    text-transform: uppercase; letter-spacing: 0.5px; border-bottom: 1px solid var(--border);
  }
  td { padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--border); font-size: 0.8125rem; color: var(--text); }
  tr:hover { background: var(--hover); }
  .mono { font-family: var(--mono); }
  .cmd { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .btn-kill { color: var(--error); border-color: var(--error); font-size: 0.75rem; padding: 0.25rem 0.5rem; }
  .confirm-kill { display: flex; align-items: center; gap: 0.25rem; font-size: 0.75rem; color: var(--error); }
  .btn-yes { color: var(--error); border-color: var(--error); padding: 0.125rem 0.375rem; font-size: 0.75rem; }
  .btn-no { padding: 0.125rem 0.375rem; font-size: 0.75rem; }
</style>
