#!/bin/bash
# =============================================================================
# Hytale Server Installation Script
# Installs the hsm binary globally for system-wide use
# =============================================================================

set -euo pipefail

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="hsm"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
C_RESET='\033[0m'
C_BOLD='\033[1m'
C_GREEN='\033[0;32m'
C_YELLOW='\033[1;33m'
C_RED='\033[0;31m'
C_CYAN='\033[0;36m'

log_info() {
    echo -e "${C_CYAN}[INFO]${C_RESET} $1"
}

log_success() {
    echo -e "${C_GREEN}[✓]${C_RESET} $1"
}

log_warn() {
    echo -e "${C_YELLOW}[WARN]${C_RESET} $1"
}

log_error() {
    echo -e "${C_RED}[ERROR]${C_RESET} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root (use sudo)"
    exit 1
fi

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    log_error "Go is not installed. Please install Go 1.19 or later."
    log_info "Visit: https://golang.org/doc/install"
    exit 1
fi

log_info "Building hsm binary..."

# Build the binary
cd "$PROJECT_DIR"
if ! go build -ldflags="-s -w" -o "/tmp/${BINARY_NAME}" ./src/cmd/hytale-tui; then
    log_error "Failed to build binary"
    exit 1
fi

log_info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."

# Install binary
if cp "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"; then
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    log_success "Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
else
    log_error "Failed to install binary"
    exit 1
fi

# Clean up temp file
rm -f "/tmp/${BINARY_NAME}"

log_success "Installation complete!"
echo ""
echo "You can now run '${BINARY_NAME}' from anywhere to open the TUI."
echo "Or run 'sudo ${BINARY_NAME}' if needed for server management."
