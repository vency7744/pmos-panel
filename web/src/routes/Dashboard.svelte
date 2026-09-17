<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { logout as apiLogout } from '../lib/api.js';
  import { isAuthenticated, username } from '../lib/stores.js';
  import { connectStatsWS, formatBytes, formatUptime, type Stats } from '../lib/stats.js';
  import Terminal from './Terminal.svelte';
  import Services from './Services.svelte';
  import Processes from './Processes.svelte';
  import Files from './Files.svelte';
  import Logs from './Logs.svelte';
  import NetworkView from './Network.svelte';
  import SystemView from './System.svelte';
  import PowerView from './Power.svelte';
  import SettingsView from './Settings.svelte';

  let currentSection = $state('dashboard');
  let stats = $state<Stats | null>(null);
  let ws: WebSocket | null = null;
  let cpuHistory = $state<number[]>([]);
  let ramHistory = $state<number[]>([]);
  let sidebarOpen = $state(false);

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: '●' },
    { id: 'terminal', label: 'Terminal', icon: '>' },
    { id: 'files', label: 'Files', icon: '📁' },
    { id: 'processes', label: 'Processes', icon: '⚡' },
    { id: 'services', label: 'Services', icon: '⚙' },
    { id: 'logs', label: 'Logs', icon: '📋' },
    { id: 'network', label: 'Network', icon: '🌐' },
    { id: 'system', label: 'System', icon: '💻' },
    { id: 'power', label: 'Power', icon: '⏻' },
    { id: 'settings', label: 'Settings', icon: '☰' },
  ];

  onMount(() => {
    ws = connectStatsWS((s) => {
      stats = s;
      cpuHistory = [...cpuHistory.slice(-29), s.cpu.usage];
      ramHistory = [...ramHistory.slice(-29), s.memory.usage];
    });
  });

  onDestroy(() => { ws?.close(); });

  async function handleLogout() {
    ws?.close();
    await apiLogout();
    isAuthenticated.logout();
    username.set('');
  }

  function tempColor(temp: number): string {
    if (temp < 50) return 'var(--green)';
    if (temp < 70) return '#eab308';
    return 'var(--error)';
  }

  function barColor(usage: number): string {
    if (usage < 50) return 'var(--accent)';
    if (usage < 80) return '#eab308';
    return 'var(--error)';
  }

  function navigate(section: string) {
    currentSection = section;
    sidebarOpen = false;
  }
</script>

<div class="app">
  <button class="hamburger" onclick={() => sidebarOpen = !sidebarOpen} aria-label="Menu">
    <span class="hamburger-line"></span>
    <span class="hamburger-line"></span>
    <span class="hamburger-line"></span>
  </button>

  {#if sidebarOpen}
    <div class="overlay" onclick={() => sidebarOpen = false} role="presentation"></div>
  {/if}

  <aside class="sidebar" class:open={sidebarOpen}>
    <div class="sidebar-header">
      <h2>PMOS Panel</h2>
      <span class="status" class:connected={!!stats}>{stats ? 'ONLINE' : 'CONNECTING'}</span>
    </div>
    <nav>
      {#each navItems as item}
        <button class="nav-item" class:active={currentSection === item.id} onclick={() => navigate(item.id)}>
          <span class="nav-icon">{item.icon}</span>
          {item.label}
        </button>
      {/each}
    </nav>
    <div class="sidebar-footer">
      <span class="username">{$username}</span>
      <button class="logout-btn" onclick={handleLogout}>Logout</button>
    </div>
  </aside>

  <main class="content">
    <div class="content-header">
      <h1>{navItems.find(n => n.id === currentSection)?.label || 'Dashboard'}</h1>
    </div>
    <div class="content-body">
      {#if currentSection === 'dashboard'}
        <div class="status-card">
          <div class="status-indicator" class:connected={!!stats}></div>
          <span>{stats ? 'System Online' : 'Connecting...'}</span>
        </div>
        {#if stats}
          <div class="stats-grid">
            <div class="stat-card">
              <div class="stat-label">CPU</div>
              <div class="stat-value">{stats.cpu.usage.toFixed(1)}%</div>
              <div class="stat-bar"><div class="stat-bar-fill" style="width:{stats.cpu.usage}%;background:{barColor(stats.cpu.usage)}"></div></div>
              <div class="stat-sub">{stats.cpu.model || `${stats.cpu.num_cpu} cores`}</div>
            </div>
            <div class="stat-card">
              <div class="stat-label">RAM</div>
              <div class="stat-value">{formatBytes(stats.memory.used)} / {formatBytes(stats.memory.total)}</div>
              <div class="stat-bar"><div class="stat-bar-fill" style="width:{stats.memory.usage}%;background:{barColor(stats.memory.usage)}"></div></div>
              <div class="stat-sub">{stats.memory.usage.toFixed(1)}% used</div>
            </div>
            <div class="stat-card">
              <div class="stat-label">Storage</div>
              <div class="stat-value">{formatBytes(stats.storage.used)} / {formatBytes(stats.storage.total)}</div>
              <div class="stat-bar"><div class="stat-bar-fill" style="width:{stats.storage.usage}%;background:{barColor(stats.storage.usage)}"></div></div>
              <div class="stat-sub">{stats.storage.path}</div>
            </div>
            <div class="stat-card">
              <div class="stat-label">Temperature</div>
              {#if stats.temperature.zones.length > 0}
                {#each stats.temperature.zones.slice(0, 2) as zone}
                  <div class="stat-value" style="color:{tempColor(zone.temp)}">{zone.temp.toFixed(1)}°C</div>
                  <div class="stat-sub">{zone.type}</div>
                {/each}
              {:else}
                <div class="stat-value">N/A</div>
              {/if}
            </div>
          </div>
          <div class="charts-row">
            <div class="chart-card">
              <h3>CPU History</h3>
              <div class="sparkline">
                {#each cpuHistory as val, i}
                  <div class="bar" style="height:{Math.max(val, 2)}%;background:{barColor(val)};left:{(i / 30) * 100}%"></div>
                {/each}
              </div>
            </div>
            <div class="chart-card">
              <h3>RAM History</h3>
              <div class="sparkline">
                {#each ramHistory as val, i}
                  <div class="bar" style="height:{Math.max(val, 2)}%;background:{barColor(val)};left:{(i / 30) * 100}%"></div>
                {/each}
              </div>
            </div>
          </div>
          <div class="info-section">
            <div class="info-card">
              <h3>Uptime</h3>
              <p class="info-value">{formatUptime(stats.uptime)}</p>
            </div>
            <div class="info-card">
              <h3>Load Average</h3>
              <p class="info-value">{stats.load_avg.load1.toFixed(2)} / {stats.load_avg.load5.toFixed(2)} / {stats.load_avg.load15.toFixed(2)}</p>
            </div>
            <div class="info-card">
              <h3>Network</h3>
              {#if stats.network.interfaces.length > 0}
                {#each stats.network.interfaces.slice(0, 3) as iface}
                  <p class="net-row"><span class="net-name">{iface.name}</span><span class="net-traffic">↓{formatBytes(iface.rx_bytes)} ↑{formatBytes(iface.tx_bytes)}</span></p>
                {/each}
              {:else}
                <p>No interfaces</p>
              {/if}
            </div>
          </div>
        {:else}
          <div class="placeholder">Connecting to system monitor...</div>
        {/if}

      {:else if currentSection === 'terminal'}
        <Terminal />
      {:else if currentSection === 'services'}
        <Services />
      {:else if currentSection === 'processes'}
        <Processes />
      {:else if currentSection === 'files'}
        <Files />
      {:else if currentSection === 'logs'}
        <Logs />
      {:else if currentSection === 'network'}
        <NetworkView />
      {:else if currentSection === 'system'}
        <SystemView />
      {:else if currentSection === 'power'}
        <PowerView />
      {:else if currentSection === 'settings'}
        <SettingsView />
      {/if}
    </div>
  </main>
</div>

<style>
  .app { display: flex; min-height: 100vh; }
  .sidebar { width: 240px; background: var(--sidebar); border-right: 1px solid var(--border); display: flex; flex-direction: column; flex-shrink: 0; }
  .sidebar-header { padding: 1.25rem 1rem; border-bottom: 1px solid var(--border); }
  .sidebar-header h2 { margin: 0 0 0.25rem; font-size: 1.125rem; color: var(--text-h); }
  .status { font-size: 0.75rem; color: #9ca3af; font-weight: 500; }
  .status.connected { color: var(--green); }
  nav { flex: 1; padding: 0.5rem; overflow-y: auto; }
  .nav-item { width: 100%; display: flex; align-items: center; gap: 0.625rem; padding: 0.5rem 0.75rem; border: none; border-radius: 6px; background: transparent; color: var(--text); font-size: 0.875rem; cursor: pointer; text-align: left; transition: background 0.15s, color 0.15s; }
  .nav-item:hover { background: var(--hover); color: var(--text-h); }
  .nav-item.active { background: var(--accent-bg); color: var(--accent); }
  .nav-icon { width: 1.25rem; text-align: center; font-size: 0.75rem; }
  .sidebar-footer { padding: 1rem; border-top: 1px solid var(--border); display: flex; flex-direction: column; gap: 0.5rem; }
  .username { font-size: 0.875rem; color: var(--text-h); }
  .logout-btn { width: 100%; padding: 0.5rem; border: 1px solid var(--border); border-radius: 6px; background: transparent; color: var(--text); font-size: 0.8125rem; cursor: pointer; }
  .logout-btn:hover { background: var(--hover); }
  .content { flex: 1; min-width: 0; }
  .content-header { padding: 1rem 1.5rem; border-bottom: 1px solid var(--border); }
  .content-header h1 { margin: 0; font-size: 1.25rem; color: var(--text-h); }
  .content-body { padding: 1.5rem; }

  .status-card { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1rem; background: var(--green-bg); border: 1px solid var(--green-border); border-radius: 8px; color: var(--green); font-weight: 500; font-size: 0.875rem; margin-bottom: 1.5rem; }
  .status-indicator { width: 8px; height: 8px; border-radius: 50%; background: #9ca3af; }
  .status-indicator.connected { background: var(--green); animation: pulse 2s infinite; }
  @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

  .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 1rem; margin-bottom: 1.5rem; }
  .stat-card { background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1rem; }
  .stat-label { font-size: 0.75rem; color: var(--text); margin-bottom: 0.375rem; text-transform: uppercase; letter-spacing: 0.5px; }
  .stat-value { font-size: 1.25rem; color: var(--text-h); font-weight: 500; font-family: var(--mono); margin-bottom: 0.375rem; }
  .stat-bar { height: 4px; background: var(--border); border-radius: 2px; overflow: hidden; margin-bottom: 0.375rem; }
  .stat-bar-fill { height: 100%; border-radius: 2px; transition: width 0.5s ease; }
  .stat-sub { font-size: 0.75rem; color: var(--text); }

  .charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1.5rem; }
  .chart-card { background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1rem; }
  .chart-card h3 { margin: 0 0 0.75rem; font-size: 0.8125rem; color: var(--text); }
  .sparkline { position: relative; height: 60px; display: flex; align-items: flex-end; }
  .bar { position: absolute; bottom: 0; width: 3%; border-radius: 2px 2px 0 0; transition: height 0.5s ease; }

  .info-section { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1rem; }
  .info-card { background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1rem; }
  .info-card h3 { margin: 0 0 0.5rem; font-size: 0.875rem; color: var(--text-h); }
  .info-card p { margin: 0; color: var(--text); font-size: 0.8125rem; }
  .info-value { font-family: var(--mono); font-size: 1rem; color: var(--text-h); }
  .net-row { display: flex; justify-content: space-between; padding: 0.25rem 0; }
  .net-name { color: var(--text-h); }
  .net-traffic { font-family: var(--mono); font-size: 0.75rem; }

  .placeholder { display: flex; align-items: center; justify-content: center; height: 300px; color: var(--text); font-size: 0.875rem; }

  .hamburger {
    display: none; position: fixed; top: 0.75rem; left: 0.75rem; z-index: 101;
    flex-direction: column; gap: 4px; padding: 0.5rem; border: 1px solid var(--border);
    border-radius: 6px; background: var(--card); cursor: pointer;
  }
  .hamburger-line { display: block; width: 18px; height: 2px; background: var(--text); border-radius: 1px; }
  .overlay { display: none; }

  @media (max-width: 768px) {
    .hamburger { display: flex; }
    .sidebar { position: fixed; left: -240px; top: 0; bottom: 0; z-index: 100; transition: left 0.2s ease; }
    .sidebar.open { left: 0; }
    .overlay { display: block; position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 99; }
    .content-body { padding: 1rem; }
    .stats-grid { grid-template-columns: 1fr 1fr; }
    .charts-row { grid-template-columns: 1fr; }
  }
</style>
