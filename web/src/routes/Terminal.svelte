<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  let container: HTMLDivElement;
  let terminal: any;
  let fitAddon: any;
  let ws: WebSocket | null = null;
  let resizeObserver: ResizeObserver | null = null;

  onMount(async () => {
    const { Terminal } = await import('@xterm/xterm');
    const { FitAddon } = await import('@xterm/addon-fit');
    const { WebLinksAddon } = await import('@xterm/addon-web-links');

    await import('@xterm/xterm/css/xterm.css');

    terminal = new Terminal({
      fontSize: 14,
      fontFamily: 'ui-monospace, "Cascadia Code", "Source Code Pro", Menlo, Consolas, monospace',
      theme: {
        background: '#0f1117',
        foreground: '#e5e7eb',
        cursor: '#3b82f6',
        selectionBackground: 'rgba(59,130,246,0.3)',
      },
      cursorBlink: true,
      scrollback: 10000,
    });

    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.loadAddon(new WebLinksAddon());
    terminal.open(container);
    fitAddon.fit();

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${protocol}//${location.host}/api/ws/terminal`);

    ws.onopen = () => {
      ws?.send(JSON.stringify({ type: 'resize', cols: terminal.cols, rows: terminal.rows }));
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'pong') return;
      } catch {
        terminal.write(event.data);
      }
    };

    ws.onclose = () => {
      terminal.write('\r\n\x1b[31m[Connection closed]\x1b[0m\r\n');
    };

    terminal.onData((data: string) => {
      ws?.send(JSON.stringify({ type: 'input', data }));
    });

    terminal.onResize(({ cols, rows }: { cols: number; rows: number }) => {
      ws?.send(JSON.stringify({ type: 'resize', cols, rows }));
    });

    resizeObserver = new ResizeObserver(() => {
      fitAddon?.fit();
    });
    resizeObserver.observe(container);
  });

  onDestroy(() => {
    ws?.close();
    terminal?.dispose();
    resizeObserver?.disconnect();
  });
</script>

<div class="terminal-wrapper">
  <div class="terminal-header">
    <span class="terminal-title">Terminal</span>
    <button class="terminal-btn" onclick={() => { fitAddon?.fit(); }}>Resize</button>
  </div>
  <div class="terminal-container" bind:this={container}></div>
</div>

<style>
  .terminal-wrapper {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 80px);
    background: #0f1117;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid var(--border);
  }
  .terminal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    background: var(--card);
    border-bottom: 1px solid var(--border);
  }
  .terminal-title { font-size: 0.8125rem; color: var(--text); }
  .terminal-btn {
    padding: 0.25rem 0.5rem; border: 1px solid var(--border); border-radius: 4px;
    background: transparent; color: var(--text); font-size: 0.75rem; cursor: pointer;
  }
  .terminal-btn:hover { background: var(--hover); }
  .terminal-container { flex: 1; padding: 4px; overflow: hidden; }
</style>
