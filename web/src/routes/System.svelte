<script lang="ts">
  import { onMount } from 'svelte';
  import { formatBytes, formatUptime } from '../lib/stats.js';
  import { connectStatsWS, type Stats } from '../lib/stats.js';

  let stats = $state<Stats | null>(null);

  onMount(() => {
    const ws = connectStatsWS((s) => { stats = s; });
    return () => ws.close();
  });
</script>

<div class="system-page">
  <h2>System Information</h2>

  {#if stats}
    <div class="info-grid">
      <div class="info-card">
        <h3>General</h3>
        <div class="info-row"><span>Architecture</span><span class="mono">{stats.cpu.arch}</span></div>
        <div class="info-row"><span>CPU Cores</span><span class="mono">{stats.cpu.num_cpu}</span></div>
        <div class="info-row"><span>CPU Model</span><span class="mono">{stats.cpu.model}</span></div>
        <div class="info-row"><span>Uptime</span><span class="mono">{formatUptime(stats.uptime)}</span></div>
        <div class="info-row"><span>Load 1/5/15</span><span class="mono">{stats.load_avg.load1.toFixed(2)} / {stats.load_avg.load5.toFixed(2)} / {stats.load_avg.load15.toFixed(2)}</span></div>
      </div>

      <div class="info-card">
        <h3>Memory</h3>
        <div class="info-row"><span>Total</span><span class="mono">{formatBytes(stats.memory.total)}</span></div>
        <div class="info-row"><span>Used</span><span class="mono">{formatBytes(stats.memory.used)}</span></div>
        <div class="info-row"><span>Available</span><span class="mono">{formatBytes(stats.memory.available)}</span></div>
        <div class="info-row"><span>Usage</span><span class="mono">{stats.memory.usage.toFixed(1)}%</span></div>
      </div>

      <div class="info-card">
        <h3>Storage</h3>
        <div class="info-row"><span>Total</span><span class="mono">{formatBytes(stats.storage.total)}</span></div>
        <div class="info-row"><span>Used</span><span class="mono">{formatBytes(stats.storage.used)}</span></div>
        <div class="info-row"><span>Free</span><span class="mono">{formatBytes(stats.storage.free)}</span></div>
        <div class="info-row"><span>Mount Point</span><span class="mono">{stats.storage.path}</span></div>
      </div>

      <div class="info-card">
        <h3>Temperature</h3>
        {#if stats.temperature.zones.length > 0}
          {#each stats.temperature.zones as zone}
            <div class="info-row">
              <span>{zone.type}</span>
              <span class="mono">{zone.temp.toFixed(1)}°C</span>
            </div>
          {/each}
        {:else}
          <div class="info-row"><span>No thermal sensors</span></div>
        {/if}
      </div>
    </div>
  {:else}
    <div class="loading">Loading system info...</div>
  {/if}
</div>

<style>
  .system-page { padding: 0; }
  h2 { margin: 0 0 1.5rem; font-size: 1rem; color: var(--text-h); }
  .info-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 1rem; }
  .info-card {
    background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1rem;
  }
  .info-card h3 { margin: 0 0 0.75rem; font-size: 0.875rem; color: var(--text-h); }
  .info-row {
    display: flex; justify-content: space-between; padding: 0.375rem 0;
    border-bottom: 1px solid var(--border); font-size: 0.8125rem;
  }
  .info-row:last-child { border-bottom: none; }
  .info-row span:first-child { color: var(--text); }
  .info-row span:last-child { color: var(--text-h); }
  .mono { font-family: var(--mono); }
  .loading { color: var(--text); padding: 2rem; text-align: center; }
</style>
