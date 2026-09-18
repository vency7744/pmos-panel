# PMOS Control Panel

Web-based management panel untuk perangkat [postmarketOS](https://postmarketos.org).

Monitor dan kontrol HP postmarketOS dari browser di perangkat lain dalam jaringan yang sama.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat&logo=svelte)
![License](https://img.shields.io/badge/License-MIT-green)

## Fitur

| Feature | Status |
|---------|--------|
| Dashboard (CPU, RAM, Storage, Temp, Load, Uptime) | Done |
| Terminal (WebSocket PTY, xterm.js) | Done |
| File Manager (browse, upload, download, delete, mkdir) | Done |
| Process Manager (list, search, sort, kill) | Done |
| Service Manager (OpenRC auto-detect, start/stop/restart) | Done |
| Log Viewer + Live Streaming (WebSocket) | Done |
| Network Monitor (read-only) | Done |
| System Info | Done |
| Power (Reboot/Shutdown) | Done |
| Authentication (sessions, rate limiting) | Done |
| Mobile Responsive (hamburger menu) | Done |
| Dark Mode | Done |

## Tech Stack

- **Backend:** Go (net/http, gorilla/websocket, creack/pty)
- **Frontend:** Svelte 5 + TypeScript + CSS
- **Deployment:** Single binary with embedded frontend
- **Target:** postmarketOS ARM64 (OpenRC / systemd)

## Project Structure

```
pmos-panel/
├── cmd/server/              # Entry point
│   ├── main.go
│   └── embed.go
├── internal/
│   ├── api/                 # HTTP handlers + WebSocket
│   ├── auth/                # Authentication & sessions
│   ├── config/              # Configuration
│   ├── filemanager/         # File operations (sandboxed)
│   ├── journald/            # Log streaming (OpenRC + systemd)
│   ├── logger/              # Structured logging (slog)
│   ├── monitor/             # System monitoring (/proc, /sys)
│   ├── netinfo/             # Network interface info
│   ├── power/               # Reboot/shutdown
│   ├── process/             # Process list & kill
│   ├── service/             # Service manager (OpenRC + systemd)
│   └── terminal/            # WebSocket PTY terminal
├── web/                     # Svelte frontend
│   ├── src/
│   │   ├── routes/          # Page components
│   │   └── lib/             # Utilities & stores
│   └── dist/                # Built frontend (embedded)
├── deploy/
│   ├── pmos-panel.service   # systemd service file
│   ├── pmos-panel-openrc    # OpenRC init script
│   └── config.example.json  # Example config
├── PLAN.md                  # Full project plan (12 phases)
├── AGENTS.md                # AI agent instructions
├── README.md
├── build.sh                 # Build script
├── install.sh               # Install script (systemd)
└── install-openrc.sh        # Install script (OpenRC)
```

## Quick Start

### Build (di laptop, cross-compile)

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build Go binary (Linux ARM64)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w" -o pmos-panel ./cmd/server
```

> Tidak perlu Go/Node.js di HP! Binary di-embed langsung semua.

### Deploy ke HP (OpenRC)

```bash
# Copy binary ke HP
scp pmos-panel fw@<device-ip>:/tmp/

# Login ke HP
ssh fw@<device-ip>

# Copy ke lokasi persisten (butuh root)
sudo cp /tmp/pmos-panel /usr/local/bin/pmos-panel
sudo chmod +x /usr/local/bin/pmos-panel

# Buat config
cat > ~/config.json << 'EOF'
{
  "server": { "port": 8080 },
  "auth": { "session_timeout_minutes": 60 },
  "logging": { "level": "info", "format": "text" },
  "file_path": { "root": "/home/fw" }
}
EOF

# Buat OpenRC service (butuh root)
sudo tee /etc/init.d/pmos-panel << 'EOF'
#!/sbin/openrc-run
name="pmos-panel"
description="PMOS Control Panel"
command="/usr/local/bin/pmos-panel"
command_args="/home/fw/config.json"
command_background=true
pidfile="/run/pmos-panel.pid"
output_log="/var/log/pmos-panel.log"
error_log="/var/log/pmos-panel.log"
depend() {
    need net
    after firewall
}
start_pre() {
    nft add rule inet filter input tcp dport 8080 accept 2>/dev/null || true
}
EOF
sudo chmod +x /etc/init.d/pmos-panel

# Enable & start
sudo rc-update add pmos-panel default
sudo rc-service pmos-panel start
```

### Atau pakai install script

```bash
# Setelah copy binary ke /tmp
sudo sh install-openrc.sh
```

## Configuration

Via JSON file atau environment variables:

| Env Var | Description | Default |
|---------|-------------|---------|
| `PMOS_PANEL_PORT` | Server port | 8080 |
| `PMOS_PANEL_HOST` | Bind address | 0.0.0.0 |
| `PMOS_PANEL_SECRET_KEY` | Session secret | random (crypto/rand) |
| `PMOS_PANEL_SESSION_TIMEOUT` | Session timeout (min) | 60 |
| `PMOS_PANEL_LOG_LEVEL` | Log level | info |
| `PMOS_PANEL_LOG_FORMAT` | Log format | text |
| `PMOS_PANEL_FILE_ROOT` | File manager root | ~/pmos-panel |
| `PMOS_PANEL_USERNAME` | Default username | admin |
| `PMOS_PANEL_PASSWORD` | Default password | admin1234 |

## API

```
POST /api/login
POST /api/logout
GET  /api/me
GET  /api/health
GET  /api/stats
GET  /api/processes
GET  /api/services
GET  /api/files?path=/
GET  /api/files/read?path=/
GET  /api/files/download?path=/
GET  /api/logs?lines=100&service=&query=
GET  /api/logs/services
GET  /api/network
POST /api/files/upload
POST /api/files/write|mkdir|delete|rename
POST /api/services/{name}/start|stop|restart|enable|disable
POST /api/processes/{pid}/kill
POST /api/power/reboot|shutdown
WS   /api/ws/stats        (realtime CPU/RAM/temp setiap 2 detik)
WS   /api/ws/terminal     (interactive PTY shell)
WS   /api/ws/logs         (live log streaming)
```

## Security

- Password hashing: salt + iterated SHA-256
- Sessions: HTTP-only cookie, SameSite=Strict
- Rate limiting: 5 login attempts / 5 min per IP
- WebSocket: origin validation (Origin == Host)
- File manager: path traversal protection, sandboxed root
- Input validation on all API parameters
- Security headers: X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Referrer-Policy
- Init system auto-detected (OpenRC / systemd)

## Init System

Panel otomatis detect init system yang dipakai:

| Init | Commands Used |
|------|---------------|
| OpenRC | `rc-service`, `rc-update`, `logread`, `reboot`/`poweroff` |
| systemd | `systemctl`, `journalctl`, `shutdown` |

## License

MIT
