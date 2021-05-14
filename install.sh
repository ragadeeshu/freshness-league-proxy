#!/bin/sh -e

[ $(id -u) != "0" ] || {
  echo "Don't run this as root!" >&2
  exit 1
}

[ -x "$(which systemctl)" ] || {
  echo "systemctl not found. Are you not using Systemd?" >&2
  exit 1
}

dir=$(dirname $(realpath $0))
port=${1:-8080}

echo "# Building binary..."
cd $dir
go build

echo "# Changing permissions on file to be a bit more secure (still quite insecure, though)..."
chmod 700 $dir/freshness-league-proxy

echo "# Installing service file for running freshness-league-proxy on port $port..."

sudo tee /etc/systemd/system/freshness-league-proxy.service > /dev/null << EOF
[Unit]
After=network-online.target

[Service]
User=$(id -u)
Group=$(id -g)
WorkingDirectory=$dir
ExecStart=$dir/freshness-league-proxy $port

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable freshness-league-proxy.service

echo "# Starting freshness-league-proxy on port $port..."
sudo systemctl start freshness-league-proxy.service
