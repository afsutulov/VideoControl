#!/usr/bin/env bash
set -euo pipefail

APP=VideoControl
CONFIG=/etc/videocontrol.conf
SERVICE=/etc/systemd/system/videocontrol.service

if [ ! -f "./$APP" ]; then
  echo "Binary ./$APP not found. Build first:"
  echo "  go build -o VideoControl ."
  exit 1
fi

install -m 0755 "./$APP" /usr/local/bin/VideoControl

if [ ! -f "$CONFIG" ]; then
  install -m 0644 configs/videocontrol.conf.example.json "$CONFIG"
  echo "Created $CONFIG. Edit it before starting the service."
fi

install -m 0644 systemd/videocontrol.service "$SERVICE"
systemctl daemon-reload
systemctl enable videocontrol

echo "Installed. Start with:"
echo "  systemctl start videocontrol"
echo "Logs:"
echo "  journalctl -u videocontrol -f"
