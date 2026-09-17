export interface ThermalZone {
  name: string;
  type: string;
  temp: number;
}

export interface NetworkInterface {
  name: string;
  rx_bytes: number;
  tx_bytes: number;
  rx_pkts: number;
  tx_pkts: number;
  status: string;
}

export interface Stats {
  cpu: {
    usage: number;
    num_cpu: number;
    model: string;
    arch: string;
    core_temp: number;
  };
  memory: {
    total: number;
    used: number;
    available: number;
    usage: number;
  };
  storage: {
    total: number;
    used: number;
    free: number;
    usage: number;
    path: string;
  };
  network: {
    interfaces: NetworkInterface[];
  };
  temperature: {
    zones: ThermalZone[];
  };
  uptime: number;
  load_avg: {
    load1: number;
    load5: number;
    load15: number;
  };
  timestamp: string;
}

export function connectStatsWS(
  onMessage: (stats: Stats) => void,
  onError?: (err: Event) => void
): WebSocket {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${protocol}//${location.host}/api/ws/stats`);

  ws.onmessage = (event) => {
    try {
      const stats: Stats = JSON.parse(event.data);
      onMessage(stats);
    } catch {}
  };

  ws.onerror = (err) => {
    onError?.(err);
  };

  return ws;
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

export function formatUptime(seconds: number): string {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);

  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}
