#!/bin/sh
set -e
cd "$(dirname "$0")"

echo "Cleaning..."
rm -rf web/dist web/node_modules

echo "Building frontend..."
cd web
npm install
npm run build
cd ..

echo "Building backend..."
go build -trimpath -ldflags="-s -w" -o pmos-panel ./cmd/server

echo "Build complete: pmos-panel"
ls -lh pmos-panel
