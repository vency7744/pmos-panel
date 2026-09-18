#!/bin/sh
# PMOS Panel - OpenRC Install Script
# Run this on the device after copying pmos-panel binary
set -e

PANEL_BIN="${1:-/tmp/pmos-panel}"
CONFIG_FILE="/home/${USER:-fw}/config.json"
FM_ROOT="/home/${USER:-fw}/pmos-panel"

if [ ! -f "$PANEL_BIN" ]; then
    echo "Error: Binary not found at $PANEL_BIN"
    echo "Usage: $0 [path-to-binary]"
    exit 1
fi

echo "=== PMOS Panel Installer (OpenRC) ==="

# 1. Config
if [ ! -f "$CONFIG_FILE" ]; then
    mkdir -p "$(dirname "$CONFIG_FILE")"
    cat > "$CONFIG_FILE" << EOF
{
  "server": { "port": 8080 },
  "auth": { "session_timeout_minutes": 60 },
  "logging": { "level": "info", "format": "text" },
  "file_path": { "root": "$FM_ROOT" }
}
EOF
    echo "[OK] Config created: $CONFIG_FILE"
else
    echo "[OK] Config exists: $CONFIG_FILE"
fi

mkdir -p "$FM_ROOT"

# 2. OpenRC service
cat > /etc/init.d/pmos-panel << 'INITEOF'
#!/sbin/openrc-run

name="pmos-panel"
description="PMOS Control Panel"

command="/tmp/pmos-panel"
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
INITEOF
chmod +x /etc/init.d/pmos-panel
echo "[OK] OpenRC service created: /etc/init.d/pmos-panel"

# 3. Enable on boot
rc-update add pmos-panel default 2>/dev/null
echo "[OK] Enabled on boot"

# 4. Start
rc-service pmos-panel start 2>/dev/null
echo "[OK] Service started"

echo ""
echo "=== Done ==="
echo "Panel: http://$(hostname -I | awk '{print $1}'):8080"
echo "Login: admin / admin1234"
