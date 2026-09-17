#!/bin/sh
set -e

APP_NAME="pmos-panel"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"

echo "Building $APP_NAME..."
cd "$(dirname "$0")"
go build -trimpath -ldflags="-s -w" -o "$APP_NAME" ./cmd/server

echo "Installing binary..."
sudo cp "$APP_NAME" "$INSTALL_DIR/$APP_NAME"
sudo chmod +x "$INSTALL_DIR/$APP_NAME"

echo "Creating user..."
id "$APP_NAME" 2>/dev/null || sudo useradd -r -s /bin/false "$APP_NAME"

echo "Installing service..."
sudo cp deploy/pmos-panel.service "$SERVICE_DIR/"
sudo systemctl daemon-reload
sudo systemctl enable "$APP_NAME"
sudo systemctl start "$APP_NAME"

echo "Done! Service started on port 8080"
echo "Default credentials: admin / admin1234"
echo "Change immediately: sudo systemctl stop $APP_NAME"
