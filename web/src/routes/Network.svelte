<script lang="ts">
  import { onMount } from 'svelte';
  import { formatBytes } from '../lib/stats.js';

  interface NetIface {
    name: string;
    status: string;
    ip4: string[];
    ip6: string[];
    mac: string;
    speed: string;
    rx_bytes: number;
    tx_bytes: number;
    rx_packets: number;
    tx_packets: number;
  }

  let ifaces = $state<NetIface[]>([]);
  let loading = $state(false);

  onMount(() => loadNetwork());

  async function loadNetwork() {
    loading = true;
    const res = await fetch('/api/network');
    if (res.ok) ifaces = await res.json();
    loading = false;
  }

  function statusColor(status: string) {
    return status === 'up' ? 'var(--green)' : 'var(--text)';
  }
</script>

<div class="network-page">
  <div class="toolbar">
    <h2>Network Interfaces</h2>
    <button onclick={loadNetwork} disabled={loading}>Refresh</button>
  </div>

  <div class="ifaces-grid">
    {#each ifaces as iface}
      <div class="iface-card">
        <div class="iface-header">
          <span class="iface-name">{iface.name}</span>
          <span class="iface-status" style="color:{statusColor(iface.status)}">{iface.status}</span>
        </div>

        <div class="iface-details">
          {#if iface.mac}
            <div class="detail-row">
              <span class="detail-label">MAC</span>
              <span class="detail-value mono">{iface.mac}</span>
            </div>
          {/if}

          {#if iface.speed && iface.speed !== 'unknown Mbps'}
            <div class="detail-row">
              <span class="detail-label">Speed</span>
              <span class="detail-value">{iface.speed}</span>
            </div>
          {/if}

          {#each iface.ip4 as ip}
            <div class="detail-row">
              <span class="detail-label">IPv4</span>
              <span class="detail-value mono">{ip}</span>
            </div>
          {/each}

          {#each iface.ip6 as ip}
            <div class="detail-row">
              <span class="detail-label">IPv6</span>
              <span class="detail-value mono">{ip}</span>
            </div>
          {/each}

          <div class="detail-row">
            <span class="detail-label">RX</span>
            <span class="detail-value mono">{formatBytes(iface.rx_bytes)} ({iface.rx_packets} pkts)</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">TX</span>
            <span class="detail-value mono">{formatBytes(iface.tx_bytes)} ({iface.tx_packets} pkts)</span>
          </div>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .network-page { padding: 0; }
  .toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
  .toolbar h2 { margin: 0; font-size: 1rem; color: var(--text-h); }
  button {
    padding: 0.5rem 0.75rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--card); color: var(--text); font-size: 0.8125rem; cursor: pointer;
  }
  button:hover { background: var(--hover); }
  .ifaces-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 1rem; }
  .iface-card {
    background: var(--card); border: 1px solid var(--border); border-radius: 8px; overflow: hidden;
  }
  .iface-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 0.75rem 1rem; border-bottom: 1px solid var(--border);
  }
  .iface-name { font-weight: 500; color: var(--text-h); }
  .iface-status { font-size: 0.75rem; text-transform: uppercase; font-weight: 500; }
  .iface-details { padding: 0.75rem 1rem; }
  .detail-row { display: flex; justify-content: space-between; padding: 0.25rem 0; }
  .detail-label { font-size: 0.75rem; color: var(--text); }
  .detail-value { font-size: 0.8125rem; color: var(--text-h); }
  .mono { font-family: var(--mono); }
</style>
