#!/bin/bash
#
# install.sh - xsh Installation Script (Linux & macOS only)
# Usage: curl -sSL https://raw.githubusercontent.com/benoitpetit/xsh/master/scripts/install.sh | bash
#
# For Windows, download the .exe directly from:
# https://github.com/benoitpetit/xsh/releases/latest

set -e

# Configuration
REPO="benoitpetit/xsh"
BINARY_NAME="xsh"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# OS Detection
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            PLATFORM="linux"
            ;;
        darwin)
            PLATFORM="darwin"
            ;;
        mingw*|msys*|cygwin*)
            print_error "Windows detected. This bash script does not support Windows."
            print_info "Please download the .exe file directly from:"
            print_info "https://github.com/benoitpetit/xsh/releases/latest"
            exit 1
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    print_info "Detected platform: ${PLATFORM}-${ARCH}"
}

# Dependency Check
check_deps() {
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
}

# Installation from GitHub releases
install_from_release() {
    print_info "Fetching latest release..."

    # Get latest version
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        print_error "Unable to fetch latest version"
        exit 1
    fi

    print_info "Version: ${LATEST_VERSION}"

    # Build URL - direct binary (uncompressed)
    ASSET_NAME="${BINARY_NAME}-${PLATFORM}-${ARCH}"
    if [ "$PLATFORM" = "windows" ]; then
        ASSET_NAME="${ASSET_NAME}.exe"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${ASSET_NAME}"

    # Download
    print_info "Downloading from: ${DOWNLOAD_URL}"
    TMP_DIR=$(mktemp -d)
    curl -sSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${BINARY_NAME}"

    if [ ! -f "${TMP_DIR}/${BINARY_NAME}" ]; then
        print_error "Download failed"
        exit 1
    fi

    chmod +x "${TMP_DIR}/${BINARY_NAME}"

    # Installation
    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        print_warning "Permission required to install in ${INSTALL_DIR}"
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    rm -rf "$TMP_DIR"

    print_success "xsh installed in ${INSTALL_DIR}/${BINARY_NAME}"
}

# Installation via Go
install_from_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go version: ${GO_VERSION}"

    print_info "Installing via go install..."
    go install "github.com/${REPO}@latest"

    # Verify binary is in PATH
    if command -v xsh &> /dev/null; then
        print_success "xsh installed"
    else
        print_warning "xsh installed but may not be in PATH"
        print_info "Make sure \$GOPATH/bin or \$HOME/go/bin is in your PATH"
    fi
}

# Installation from source
install_from_source() {
    if ! command -v go &> /dev/null; then
        print_error "Go is required to compile from source"
        exit 1
    fi

    print_info "Cloning repository..."
    TMP_DIR=$(mktemp -d)
    git clone "https://github.com/${REPO}.git" "${TMP_DIR}/xsh"

    print_info "Building..."
    cd "${TMP_DIR}/xsh"
    go build -o "${TMP_DIR}/xsh" .

    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMP_DIR}/xsh" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "${TMP_DIR}/xsh" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    rm -rf "$TMP_DIR"

    print_success "xsh compiled and installed"
}

# Welcome message
show_welcome() {
    echo ""
    echo "╔══════════════════════════════════════╗"
    echo "║                                      ║"
    echo "║   xsh - Twitter/X CLI Installer      ║"
    echo "║                                      ║"
    echo "╚══════════════════════════════════════╝"
    echo ""
}

# Post-installation message
show_post_install() {
    echo ""
    print_success "Installation complete!"
    echo ""
    echo "Quick commands:"
    echo "  xsh auth login       # Authentication"
    echo "  xsh feed             # Timeline"
    echo "  xsh --help           # Help"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo ""
}

# Main
main() {
    show_welcome

    check_deps
    detect_os

    # Installation method
    METHOD="${1:-release}"

    case "$METHOD" in
        release)
            install_from_release
            ;;
        go)
            install_from_go
            ;;
        source)
            install_from_source
            ;;
        *)
            print_error "Unknown method: $METHOD"
            print_info "Usage: $0 [release|go|source]"
            exit 1
            ;;
    esac

    show_post_install
}

main "$@"
