#!/usr/bin/env bash

set -euo pipefail

# Always run from the repository root
cd "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."

BUILD_DIR="${BUILD_DIR:-build}"
mkdir -p "$BUILD_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required but was not found in PATH."
  echo "Install Go from https://go.dev/dl/ and try again."
  exit 1
fi

echo "[hytale-server-manager] Building HSM (Hytale Server Manager CLI)..."
go build -ldflags="-s -w" -o "${BUILD_DIR}/hsm" ./src/cmd/hytale-tui

echo "[hytale-server-manager] Launching HSM with DEBUG logging enabled..."

# Only show sudo warning if not running as root
if [ "$EUID" -ne 0 ]; then
	echo ""
	echo "⚠️  Note: HSM requires sudo privileges for:"
	echo "   • Creating system user (hytaleservermanager)"
	echo "   • Creating /var/lib/hytale directory"
	echo "   • Managing server processes"
	echo ""
	echo "If you see permission errors, run: sudo ${BUILD_DIR}/hsm"
	echo ""
fi

DEBUG=1 exec "${BUILD_DIR}/hsm"
