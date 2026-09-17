# PMOS Control Panel

Web-based management panel untuk perangkat [postmarketOS](https://postmarketos.org).

Monitor dan kontrol HP postmarketOS dari browser di perangkat lain dalam jaringan yang sama.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat&logo=svelte)
![License](https://img.shields.io/badge/License-MIT-green)

## Screenshot

```
┌───────────────────────────────────────┐
│ PMOS CONTROL PANEL                   │
├───────────────────────────────────────┤
│ ● ONLINE                             │
│                                       │
│ CPU        23%                        │
│ RAM        612 MB / 2 GB              │
│ STORAGE    8.2 GB / 32 GB             │
│ TEMP       42°C                       │
│ UPTIME     2d 14h                     │
└───────────────────────────────────────┘
```

## Fitur

| Feature | Status |
|---------|--------|
| Dashboard (CPU, RAM, Storage, Temp) | Done |
| Terminal (WebSocket PTY) | Done |
| File Manager | Done |
| Process Manager | Done |
| Systemd Service Manager | Done |
| Log Viewer + Live Streaming | Done |
| Network Monitor | Done |
| System Info | Done |
| Power (Reboot/Shutdown) | Done |
| Authentication | Done |
| Mobile Responsive | Done |
| Dark Mode | Done |

## Tech Stack

- **Backend:** Go (net/http, gorilla/websocket, creack/pty)
- **Frontend:** Svelte 5 + TypeScript + CSS
- **Deployment:** Single binary with embedded frontend
- **Target:** postmarketOS ARM64 (systemd)

## Project Structure

```
pmos-panel/
├── cmd/server/          # Entry point
│   ├── main.go
│   └── embed.go
├── internal/
│   ├── api/             # HTTP handlers + WebSocket
│   ├── auth/            # Authentication & sessions
│   ├── config/          # Configuration
│   ├── filemanager/     # File operations
│   ├── journald/        # Log streaming
│   ├── logger/          # Structured logging
│   ├── monitor/         # System monitoring (/proc, /sys)
│   ├── netinfo/         # Network interface info
│   ├── power/           # Reboot/shutdown
│   ├── process/         # Process list & kill
│   ├── service/         # systemd service manager
│   └── terminal/        # WebSocket PTY terminal
├── web/                 # Svelte frontend
│   ├── src/
│   │   ├── routes/      # Page components
│   │   └── lib/         # Utilities & stores
│   └── dist/            # Built frontend
├── deploy/              # Systemd service files
├── PLAN.md              # Full project plan
├── AGENTS.md            # AI agent instructions
├── build.sh             # Build script
└── install.sh           # Install script
```

## Quick Start

### Build

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build Go binary (Linux ARM64)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w" -o pmos-panel ./cmd/server
```

### Deploy to Device

```bash
# Copy binary to device
scp pmos-panel fwk@<device-ip>:/tmp/

# On device
ssh fwk@<device-ip>
chmod +x /tmp/pmos-panel

# Create config
cat > ~/config.json << 'EOF'
{
  "server": { "port": 8080 },
  "auth": { "session_timeout_minutes": 60 },
  "logging": { "level": "info", "format": "text" },
  "file_path": { "root": "/home/fwk/pmos-panel" }
}
EOF

# Run
/tmp/pmos-panel ~/config.json
```

### Auto-Start (systemd)

```bash
# User service
mkdir -p ~/.config/systemd/user
cat > ~/.config/systemd/user/pmos-panel.service << 'EOF'
[Unit]
Description=PMOS Control Panel
After=network.target

[Service]
ExecStart=/tmp/pmos-panel %h/config.json
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable --now pmos-panel

# Open firewall port (needs root)
sudo nft add rule inet filter input tcp dport 8080 accept
```

## Configuration

Via JSON file or environment variables:

| Env Var | Description | Default |
|---------|-------------|---------|
| `PMOS_PANEL_PORT` | Server port | 8080 |
| `PMOS_PANEL_HOST` | Bind address | 0.0.0.0 |
| `PMOS_PANEL_SECRET_KEY` | Session secret | random |
| `PMOS_PANEL_SESSION_TIMEOUT` | Session timeout (min) | 60 |
| `PMOS_PANEL_LOG_LEVEL` | Log level | info |
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
GET  /api/logs?lines=100
GET  /api/network
POST /api/services/{name}/start|stop|restart|enable|disable
POST /api/processes/{pid}/kill
POST /api/files/upload
POST /api/files/write|mkdir|delete|rename
POST /api/power/reboot|shutdown
WS   /api/ws/stats
WS   /api/ws/terminal
WS   /api/ws/logs
```

## Security

- Password hashing with salt + iterated SHA-256
- Secure HTTP-only sessions with SameSite=Strict
- Rate limiting on login (5 attempts / 5 min per IP)
- WebSocket origin validation
- Path traversal protection on file manager
- Input validation on all API parameters
- Security headers (X-Content-Type-Options, X-Frame-Options, etc.)

## License

MIT
