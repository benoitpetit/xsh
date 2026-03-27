#!/bin/bash
#
# install.sh - Script d'installation de xsh (Linux & macOS uniquement)
# Usage: curl -sSL https://raw.githubusercontent.com/benoitpetit/xsh/master/scripts/install.sh | bash
#
# Pour Windows, téléchargez directement le .exe depuis:
# https://github.com/benoitpetit/xsh/releases/latest

set -e

# Configuration
REPO="benoitpetit/xsh"
BINARY_NAME="xsh"
INSTALL_DIR="/usr/local/bin"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Fonctions
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

# Détection de l'OS
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
            print_error "Windows détecté. Ce script bash ne supporte pas Windows."
            print_info "Veuillez télécharger directement le fichier .exe depuis:"
            print_info "https://github.com/benoitpetit/xsh/releases/latest"
            exit 1
            ;;
        *)
            print_error "OS non supporté: $OS"
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
            print_error "Architecture non supportée: $ARCH"
            exit 1
            ;;
    esac

    print_info "Plateforme détectée: ${PLATFORM}-${ARCH}"
}

# Vérification des dépendances
check_deps() {
    if ! command -v curl &> /dev/null; then
        print_error "curl est requis mais n'est pas installé"
        exit 1
    fi
}

# Installation depuis les releases GitHub
install_from_release() {
    print_info "Récupération de la dernière release..."

    # Récupérer la dernière version
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        print_error "Impossible de récupérer la dernière version"
        exit 1
    fi

    print_info "Version: ${LATEST_VERSION}"

    # Construction de l'URL - binaire direct (non compressé)
    ASSET_NAME="${BINARY_NAME}-${PLATFORM}-${ARCH}"
    if [ "$PLATFORM" = "windows" ]; then
        ASSET_NAME="${ASSET_NAME}.exe"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${ASSET_NAME}"

    # Téléchargement
    print_info "Téléchargement depuis: ${DOWNLOAD_URL}"
    TMP_DIR=$(mktemp -d)
    curl -sSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${BINARY_NAME}"

    if [ ! -f "${TMP_DIR}/${BINARY_NAME}" ]; then
        print_error "Échec du téléchargement"
        exit 1
    fi

    chmod +x "${TMP_DIR}/${BINARY_NAME}"

    # Installation
    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        print_warning "Permission requise pour installer dans ${INSTALL_DIR}"
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    rm -rf "$TMP_DIR"

    print_success "xsh installé dans ${INSTALL_DIR}/${BINARY_NAME}"
}

# Installation depuis Go
install_from_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go n'est pas installé"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go version: ${GO_VERSION}"

    print_info "Installation via go install..."
    go install "github.com/${REPO}@latest"

    # Vérifier que le binaire est dans le PATH
    if command -v xsh &> /dev/null; then
        print_success "xsh installé"
    else
        print_warning "xsh installé mais peut-être pas dans le PATH"
        print_info "Assurez-vous que \$GOPATH/bin ou \$HOME/go/bin est dans votre PATH"
    fi
}

# Installation depuis le source
install_from_source() {
    if ! command -v go &> /dev/null; then
        print_error "Go est requis pour compiler depuis le source"
        exit 1
    fi

    print_info "Clonage du dépôt..."
    TMP_DIR=$(mktemp -d)
    git clone "https://github.com/${REPO}.git" "${TMP_DIR}/xsh"

    print_info "Compilation..."
    cd "${TMP_DIR}/xsh"
    go build -o "${TMP_DIR}/xsh" .

    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMP_DIR}/xsh" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "${TMP_DIR}/xsh" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    rm -rf "$TMP_DIR"

    print_success "xsh compilé et installé"
}

# Message de bienvenue
show_welcome() {
    echo ""
    echo "╔══════════════════════════════════════╗"
    echo "║                                      ║"
    echo "║   xsh - Twitter/X CLI Installer      ║"
    echo "║                                      ║"
    echo "╚══════════════════════════════════════╝"
    echo ""
}

# Message post-installation
show_post_install() {
    echo ""
    print_success "Installation terminée !"
    echo ""
    echo "Commandes rapides:"
    echo "  xsh auth login       # Authentification"
    echo "  xsh feed             # Timeline"
    echo "  xsh --help           # Aide"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo ""
}

# Main
main() {
    show_welcome

    check_deps
    detect_os

    # Méthode d'installation
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
            print_error "Méthode inconnue: $METHOD"
            print_info "Usage: $0 [release|go|source]"
            exit 1
            ;;
    esac

    show_post_install
}

main "$@"
