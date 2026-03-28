#!/bin/bash
#
# verify.sh - Release verification and preparation script for xsh
# Usage: ./core/scripts/verify.sh [VERSION]
# Example: ./core/scripts/verify.sh 0.0.2
#

set -e

# Colors
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

# Ask for version if not provided
if [ -z "$1" ]; then
    print_info "Current version detected: $(grep 'var Version' "$CORE_DIR/cmd/version.go" | sed 's/.*"\(.*\)".*/\1/')"
    echo ""
    read -p "Enter new version (e.g., 0.0.2): " TARGET_VERSION
else
    TARGET_VERSION="$1"
fi

# Version validation
if [ -z "$TARGET_VERSION" ]; then
    print_error "Version not specified"
    exit 1
fi

# Clean 'v' prefix if present
TARGET_VERSION="${TARGET_VERSION#v}"

print_info "Preparing version: v$TARGET_VERSION"
echo ""

# ============================================
# STEP 1: Update versions in all files
# ============================================
print_step "Updating versions in files..."

# 1.1 Core version.go
CURRENT_CORE_VERSION=$(grep 'var Version' "$CORE_DIR/cmd/version.go" | sed 's/.*"\(.*\)".*/\1/')
if [ "$CURRENT_CORE_VERSION" != "$TARGET_VERSION" ]; then
    sed -i "s/var Version = \"$CURRENT_CORE_VERSION\"/var Version = \"$TARGET_VERSION\"/" "$CORE_DIR/cmd/version.go"
    print_success "core/cmd/version.go: $CURRENT_CORE_VERSION → $TARGET_VERSION"
else
    print_success "core/cmd/version.go: already at $TARGET_VERSION"
fi

# 1.2 Web +page.svelte
CURRENT_WEB_VERSION=$(grep -o 'v0\.[0-9]\+\.[0-9]\+' "$WEB_DIR/src/routes/+page.svelte" | head -1 | sed 's/v//')
if [ "$CURRENT_WEB_VERSION" != "$TARGET_VERSION" ]; then
    # Updates all occurrences (v0.x.x)
    sed -i "s/v${CURRENT_WEB_VERSION}/v${TARGET_VERSION}/g" "$WEB_DIR/src/routes/+page.svelte"
    print_success "web/src/routes/+page.svelte: v$CURRENT_WEB_VERSION → v$TARGET_VERSION"
else
    print_success "web/src/routes/+page.svelte: already at v$TARGET_VERSION"
fi

# 1.3 Web Logo.svelte
CURRENT_LOGO_VERSION=$(grep 'export let version' "$WEB_DIR/src/lib/components/Logo.svelte" | sed 's/.*"\(v[^"]*\)".*/\1/' | sed 's/v//')
if [ "$CURRENT_LOGO_VERSION" != "$TARGET_VERSION" ]; then
    sed -i "s/export let version: string = \"v${CURRENT_LOGO_VERSION}\"/export let version: string = \"v${TARGET_VERSION}\"/" "$WEB_DIR/src/lib/components/Logo.svelte"
    print_success "web/src/lib/components/Logo.svelte: v$CURRENT_LOGO_VERSION → v$TARGET_VERSION"
else
    print_success "web/src/lib/components/Logo.svelte: already at v$TARGET_VERSION"
fi

# 1.4 web package.json (if version present)
if [ -f "$WEB_DIR/package.json" ]; then
    CURRENT_PKG_VERSION=$(grep '"version"' "$WEB_DIR/package.json" | head -1 | sed 's/.*"\([0-9]\+\.[0-9]\+\.[0-9]\+\)".*/\1/')
    if [ "$CURRENT_PKG_VERSION" != "$TARGET_VERSION" ]; then
        sed -i "s/\"version\": \"$CURRENT_PKG_VERSION\"/\"version\": \"$TARGET_VERSION\"/" "$WEB_DIR/package.json"
        print_success "web/package.json: $CURRENT_PKG_VERSION → $TARGET_VERSION"
    else
        print_success "web/package.json: already at $TARGET_VERSION"
    fi
fi

echo ""

# ============================================
# STEP 2: Verify Core Build
# ============================================
print_step "Verifying Core build (Go)..."
cd "$CORE_DIR"
if go build -o xsh . 2>/dev/null; then
    ACTUAL_VERSION=$(./xsh version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+')
    if [ "$ACTUAL_VERSION" = "$TARGET_VERSION" ]; then
        print_success "Build OK - Version confirmed: $ACTUAL_VERSION"
    else
        print_error "Build OK but incorrect version: $ACTUAL_VERSION (expected: $TARGET_VERSION)"
        ERRORS=$((ERRORS + 1))
    fi
    rm -f xsh
else
    print_error "Core build failed"
    ERRORS=$((ERRORS + 1))
fi

# ============================================
# STEP 3: Go vet
# ============================================
print_step "Running go vet..."
cd "$CORE_DIR"
if go vet ./... 2>/dev/null; then
    print_success "go vet passed"
else
    print_warning "go vet found issues (non-blocking)"
fi

# ============================================
# STEP 4: Verify critical files
# ============================================
print_step "Verifying critical files..."
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
        print_error "$file missing"
        ERRORS=$((ERRORS + 1))
    fi
done

# ============================================
# STEP 5: Web Build
# ============================================
print_step "Building Web (SvelteKit)..."
cd "$WEB_DIR"
if npm run build 2>/dev/null >/dev/null; then
    print_success "Web build successful"
else
    print_warning "Web build failed or has warnings (verify manually)"
fi

# ============================================
# STEP 6: Final consistency check
# ============================================
print_step "Final consistency check..."
echo ""
echo "Current versions in files:"
echo "  core/cmd/version.go:      $(grep 'var Version' "$CORE_DIR/cmd/version.go" | sed 's/.*"\(.*\)".*/\1/')"
echo "  web/+page.svelte:         $(grep -o 'v0\.[0-9]\+\.[0-9]\+' "$WEB_DIR/src/routes/+page.svelte" | head -1)"
echo "  web/Logo.svelte:          $(grep 'export let version' "$WEB_DIR/src/lib/components/Logo.svelte" | sed 's/.*"\(v[^"]*\)".*/\1/')"

echo ""

# ============================================
# SUMMARY
# ============================================
echo "╔══════════════════════════════════════════════════════════════╗"
if [ $ERRORS -eq 0 ]; then
    echo "║                  ✅ READY FOR RELEASE                        ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    print_success "Version v$TARGET_VERSION ready!"
    echo ""
    echo "Next steps:"
    echo "  1. git add -A"
    echo "  2. git commit -m \"chore: release v$TARGET_VERSION\""
    echo "  3. git push origin prod"
    echo "  4. git tag v$TARGET_VERSION"
    echo "  5. git push origin v$TARGET_VERSION"
    echo ""
    exit 0
else
    echo "║                  ❌ $ERRORS ERROR(S) FOUND                   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    print_error "Fix errors before releasing"
    exit 1
fi
