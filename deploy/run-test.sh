#!/bin/sh
/tmp/pmos-panel &
PANEL_PID=$!
echo "Panel started with PID $PANEL_PID"
sleep 2
wget -q -O - "http://localhost:8080/api/health" 2>&1 || echo "Health check via wget failed"
kill -0 $PANEL_PID 2>/dev/null && echo "Process alive" || echo "Process dead"
