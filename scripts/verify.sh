#!/bin/bash
#
# verify.sh - Script de vérification et préparation de release pour xsh
# Usage: ./core/scripts/verify.sh [VERSION]
# Exemple: ./core/scripts/verify.sh 0.0.2
#

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_step() { echo -e "${CYAN}👉 $1${NC}"; }

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
CORE_DIR="$ROOT_DIR/core"
WEB_DIR="$ROOT_DIR/web"
ERRORS=0

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         xsh - Release Preparation Tool                       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Demander la version si non fournie
if [ -z "$1" ]; then
    print_info "Version actuelle détectée: $(grep 'var Version' "$CORE_DIR/cmd/version.go" | sed 's/.*"\(.*\)".*/\1/')"
    echo ""
    read -p "Entrez la nouvelle version (ex: 0.0.2): " TARGET_VERSION
else
    TARGET_VERSION="$1"
fi

# Validation de la version
if [ -z "$TARGET_VERSION" ]; then
    print_error "Version non spécifiée"
    exit 1
fi

# Nettoyer le prefix 'v' si présent
TARGET_VERSION="${TARGET_VERSION#v}"

print_info "Préparation de la version: v$TARGET_VERSION"
echo ""

# ============================================
# ÉTAPE 1: Mettre à jour les versions dans tous les fichiers
# ============================================
print_step "Mise à jour des versions dans les fichiers..."

# 1.1 Core version.go
CURRENT_CORE_VERSION=$(grep 'var Version' "$CORE_DIR/cmd/version.go" | sed 's/.*"\(.*\)".*/\1/')
if [ "$CURRENT_CORE_VERSION" != "$TARGET_VERSION" ]; then
    sed -i "s/var Version = \"$CURRENT_CORE_VERSION\"/var Version = \"$TARGET_VERSION\"/" "$CORE_DIR/cmd/version.go"
    print_success "core/cmd/version.go: $CURRENT_CORE_VERSION → $TARGET_VERSION"
else
    print_success "core/cmd/version.go: déjà à $TARGET_VERSION"
fi

# 1.2 Web +page.svelte
CURRENT_WEB_VERSION=$(grep -o 'v0\.[0-9]\+\.[0-9]\+' "$WEB_DIR/src/routes/+page.svelte" | head -1 | sed 's/v//')
if [ "$CURRENT_WEB_VERSION" != "$TARGET_VERSION" ]; then
    # Met à jour toutes les occurrences (v0.x.x)
    sed -i "s/v${CURRENT_WEB_VERSION}/v${TARGET_VERSION}/g" "$WEB_DIR/src/routes/+page.svelte"
    print_success "web/src/routes/+page.svelte: v$CURRENT_WEB_VERSION → v$TARGET_VERSION"
else
    print_success "web/src/routes/+page.svelte: déjà à v$TARGET_VERSION"
fi

# 1.3 Web Logo.svelte
CURRENT_LOGO_VERSION=$(grep 'export let version' "$WEB_DIR/src/lib/components/Logo.svelte" | sed 's/.*"\(v[^"]*\)".*/\1/' | sed 's/v//')
if [ "$CURRENT_LOGO_VERSION" != "$TARGET_VERSION" ]; then
    sed -i "s/export let version: string = \"v${CURRENT_LOGO_VERSION}\"/export let version: string = \"v${TARGET_VERSION}\"/" "$WEB_DIR/src/lib/components/Logo.svelte"
    print_success "web/src/lib/components/Logo.svelte: v$CURRENT_LOGO_VERSION → v$TARGET_VERSION"
else
    print_success "web/src/lib/components/Logo.svelte: déjà à v$TARGET_VERSION"
fi

# 1.4 TODO.md
CURRENT_TODO_VERSION=$(grep '^\*\*Version' "$ROOT_DIR/TODO.md" | sed 's/.*: \(.*\)/\1/')
if [ "$CURRENT_TODO_VERSION" != "$TARGET_VERSION" ]; then
    sed -i "s/^\*\*Version\*\*: .*/**Version**: $TARGET_VERSION/" "$ROOT_DIR/TODO.md"
    print_success "TODO.md: $CURRENT_TODO_VERSION → $TARGET_VERSION"
else
    print_success "TODO.md: déjà à $TARGET_VERSION"
fi

# 1.5 package.json du web (si version présente)
if [ -f "$WEB_DIR/package.json" ]; then
    CURRENT_PKG_VERSION=$(grep '"version"' "$WEB_DIR/package.json" | head -1 | sed 's/.*"\([0-9]\+\.[0-9]\+\.[0-9]\+\)".*/\1/')
    if [ "$CURRENT_PKG_VERSION" != "$TARGET_VERSION" ]; then
        sed -i "s/\"version\": \"$CURRENT_PKG_VERSION\"/\"version\": \"$TARGET_VERSION\"/" "$WEB_DIR/package.json"
        print_success "web/package.json: $CURRENT_PKG_VERSION → $TARGET_VERSION"
    else
        print_success "web/package.json: déjà à $TARGET_VERSION"
    fi
fi

echo ""

# ============================================
# ÉTAPE 2: Vérifier le build Core
# ============================================
print_step "Vérification du build Core (Go)..."
cd "$CORE_DIR"
if go build -o xsh . 2>/dev/null; then
    ACTUAL_VERSION=$(./xsh version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+')
    if [ "$ACTUAL_VERSION" = "$TARGET_VERSION" ]; then
        print_success "Build OK - Version confirmée: $ACTUAL_VERSION"
    else
        print_error "Build OK mais version incorrecte: $ACTUAL_VERSION (attendu: $TARGET_VERSION)"
        ERRORS=$((ERRORS + 1))
    fi
    rm -f xsh
else
    print_error "Build Core échoué"
    ERRORS=$((ERRORS + 1))
fi

# ============================================
# ÉTAPE 3: Go vet
# ============================================
print_step "Running go vet..."
cd "$CORE_DIR"
if go vet ./... 2>/dev/null; then
    print_success "go vet passed"
else
    print_warning "go vet a trouvé des problèmes (non bloquant)"
fi

# ============================================
# ÉTAPE 4: Vérifier les fichiers critiques
# ============================================
print_step "Vérification des fichiers critiques..."
CRITICAL_FILES=(
    "$CORE_DIR/main.go"
    "$CORE_DIR/go.mod"
    "$CORE_DIR/cmd/root.go"
    "$CORE_DIR/scripts/install.sh"
    "$CORE_DIR/.github/workflows/release.yml"
    "$WEB_DIR/package.json"
    "$WEB_DIR/src/routes/+page.svelte"
)

for file in "${CRITICAL_FILES[@]}"; do
    if [ -f "$file" ]; then
        print_success "$(basename $file)"
    else
        print_error "$file manquant"
        ERRORS=$((ERRORS + 1))
    fi
done

# ============================================
# ÉTAPE 5: Build Web
# ============================================
print_step "Build Web (SvelteKit)..."
cd "$WEB_DIR"
if npm run build 2>/dev/null >/dev/null; then
    print_success "Web build successful"
else
    print_warning "Web build a échoué ou warnings (vérifiez manuellement)"
fi

# ============================================
# ÉTAPE 6: Vérification finale cohérence
# ============================================
print_step "Vérification finale de cohérence..."
echo ""
echo "Versions actuelles dans les fichiers:"
echo "  core/cmd/version.go:      $(grep 'var Version' "$CORE_DIR/cmd/version.go" | sed 's/.*"\(.*\)".*/\1/')"
echo "  web/+page.svelte:         $(grep -o 'v0\.[0-9]\+\.[0-9]\+' "$WEB_DIR/src/routes/+page.svelte" | head -1)"
echo "  web/Logo.svelte:          $(grep 'export let version' "$WEB_DIR/src/lib/components/Logo.svelte" | sed 's/.*"\(v[^"]*\)".*/\1/')"
echo "  TODO.md:                  $(grep '^\*\*Version' "$ROOT_DIR/TODO.md" | sed 's/.*: \(.*\)/\1/')"
echo ""

# ============================================
# RÉSUMÉ
# ============================================
echo "╔══════════════════════════════════════════════════════════════╗"
if [ $ERRORS -eq 0 ]; then
    echo "║                  ✅ PRÊT POUR RELEASE                        ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    print_success "Version v$TARGET_VERSION prête!"
    echo ""
    echo "Prochaines étapes:"
    echo "  1. git add -A"
    echo "  2. git commit -m \"chore: release v$TARGET_VERSION\""
    echo "  3. git push origin prod"
    echo "  4. git tag v$TARGET_VERSION"
    echo "  5. git push origin v$TARGET_VERSION"
    echo ""
    exit 0
else
    echo "║                  ❌ $ERRORS ERREUR(S) TROUVÉE(S)              ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    print_error "Corrigez les erreurs avant de release"
    exit 1
fi
