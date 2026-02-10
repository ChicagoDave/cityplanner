#!/bin/bash
set -e

echo "=== CityPlanner Development Environment Setup ==="

# --- Go ---
echo ""
echo "[1/3] Installing Go 1.25.7..."
curl -sLO https://go.dev/dl/go1.25.7.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.7.linux-amd64.tar.gz
rm go1.25.7.linux-amd64.tar.gz

# Add to PATH if not already there
if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
  echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
fi
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

echo "  Go version: $(go version)"

# --- Node.js (already installed via nvm, just verify) ---
echo ""
echo "[2/3] Verifying Node.js..."
if command -v node &> /dev/null; then
  echo "  Node version: $(node --version)"
  echo "  npm version: $(npm --version)"
else
  echo "  ERROR: Node.js not found. Install via nvm: nvm install 22"
  exit 1
fi

# --- Go module init ---
echo ""
echo "[3/3] Initializing Go module..."
cd /mnt/c/repotemp/cityplanner/solver
go mod init github.com/ChicagoDave/cityplanner
go mod tidy

echo ""
echo "=== Setup complete ==="
echo "Restart your shell or run: source ~/.bashrc"
