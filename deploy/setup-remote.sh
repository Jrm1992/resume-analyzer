#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <host> [user]"
  exit 1
fi

REMOTE_HOST="$1"
REMOTE_USER="${2:-deploy}"

echo "Setting up resume-analyzer on ${REMOTE_USER}@${REMOTE_HOST}..."

ssh "${REMOTE_USER}@${REMOTE_HOST}" << 'REMOTE'
sudo mkdir -p /opt/resume-analyzer
sudo chown deploy:deploy /opt/resume-analyzer

sudo cp /dev/stdin /etc/systemd/system/resume-analyzer.service << 'SERVICE'
[Unit]
Description=Resume Analyzer
After=network.target

[Service]
Type=simple
User=deploy
WorkingDirectory=/opt/resume-analyzer
ExecStart=/opt/resume-analyzer/resume-analyzer
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
SERVICE

sudo systemctl daemon-reload
sudo systemctl enable resume-analyzer

echo "Service installed. Run 'make deploy' to build and deploy."
REMOTE
