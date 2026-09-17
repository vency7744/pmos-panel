<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  interface LogEntry {
    timestamp: string;
    priority: string;
    service: string;
    message: string;
  }

  let entries = $state<LogEntry[]>([]);
  let services = $state<string[]>([]);
  let selectedService = $state('');
  let query = $state('');
  let lines = $state(100);
  let loading = $state(false);
  let liveMode = $state(false);
  let ws: WebSocket | null = null;
  let logContainer: HTMLDivElement;

  onMount(async () => {
    await loadLogServices();
    await loadLogs();
  });

  onDestroy(() => { stopLive(); });

  async function loadLogServices() {
    const res = await fetch('/api/logs/services');
    if (res.ok) services = await res.json();
  }

  async function loadLogs() {
    loading = true;
    let url = `/api/logs?lines=${lines}`;
    if (selectedService) url += `&service=${encodeURIComponent(selectedService)}`;
    if (query) url += `&query=${encodeURIComponent(query)}`;
    const res = await fetch(url);
    if (res.ok) entries = await res.json();
    loading = false;
    scrollToBottom();
  }

  function startLive() {
    stopLive();
    liveMode = true;
    let url = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/api/ws/logs`;
    if (selectedService) url += `?service=${encodeURIComponent(selectedService)}`;
    ws = new WebSocket(url);
    ws.onmessage = (ev) => {
      try {
        const entry: LogEntry = JSON.parse(ev.data);
        entries = [...entries.slice(-499), entry];
        scrollToBottom();
      } catch {}
    };
    ws.onerror = () => { liveMode = false; };
    ws.onclose = () => { liveMode = false; };
  }

  function stopLive() {
    ws?.close();
    ws = null;
    liveMode = false;
  }

  function toggleLive() {
    if (liveMode) { stopLive(); } else { startLive(); }
  }

  function scrollToBottom() {
    if (logContainer) {
      requestAnimationFrame(() => { logContainer.scrollTop = logContainer.scrollHeight; });
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') loadLogs();
  }
</script>

<div class="logs-page">
  <div class="toolbar">
    <input type="text" placeholder="Search logs..." bind:value={query} onkeydown={handleKeydown} />
    <select bind:value={selectedService} onchange={() => { if (liveMode) { stopLive(); startLive(); } else { loadLogs(); } }}>
      <option value="">All services</option>
      {#each services as svc}
        <option value={svc}>{svc}</option>
      {/each}
    </select>
    <select bind:value={lines} onchange={loadLogs}>
      <option value={50}>50</option>
      <option value={100}>100</option>
      <option value={200}>200</option>
      <option value={500}>500</option>
    </select>
    <button onclick={loadLogs} disabled={loading}>Search</button>
    <button class:active={liveMode} onclick={toggleLive}>{liveMode ? 'Stop Live' : 'Live'}</button>
  </div>

  <div class="log-entries" bind:this={logContainer}>
    {#each entries as entry}
      <div class="log-line">
        <span class="log-time">{entry.timestamp}</span>
        <span class="log-msg">{entry.message}</span>
      </div>
    {/each}
    {#if entries.length === 0}
      <div class="empty">No log entries</div>
    {/if}
  </div>
</div>

<style>
  .logs-page { padding: 0; }
  .toolbar { display: flex; gap: 0.5rem; margin-bottom: 1rem; flex-wrap: wrap; }
  input {
    flex: 1; min-width: 200px; padding: 0.5rem 0.75rem; border: 1px solid var(--border);
    border-radius: 6px; background: var(--input-bg); color: var(--text-h); font-size: 0.875rem; outline: none;
  }
  input:focus { border-color: var(--accent); }
  select {
    padding: 0.5rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--card); color: var(--text); font-size: 0.8125rem; outline: none;
  }
  button {
    padding: 0.5rem 0.75rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--card); color: var(--text); font-size: 0.8125rem; cursor: pointer;
  }
  button:hover { background: var(--hover); }
  button.active { background: var(--accent-bg); color: var(--accent); border-color: var(--accent); }
  .log-entries {
    background: #0f1117; border: 1px solid var(--border); border-radius: 8px;
    padding: 0.75rem; max-height: calc(100vh - 200px); overflow-y: auto;
    font-family: var(--mono); font-size: 0.75rem;
  }
  .log-line {
    display: flex; gap: 0.75rem; padding: 0.125rem 0; line-height: 1.4;
  }
  .log-time { color: #6b7280; flex-shrink: 0; }
  .log-msg { color: #e5e7eb; word-break: break-all; }
  .empty { padding: 2rem; text-align: center; color: #6b7280; }
</style>
