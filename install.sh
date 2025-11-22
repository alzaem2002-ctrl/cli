#!/bin/sh
# GitHub CLI Installation Script
# This script installs the GitHub CLI (gh) on Unix-like systems
# Usage: curl -fsSL https://cli.github.com/install.sh | sh
# or: wget -qO- https://cli.github.com/install.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default installation directory
PREFIX="${PREFIX:-$HOME/.local}"

# GitHub CLI repository information
REPO="cli/cli"
BINARY_NAME="gh"

# Helper functions
print_info() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

print_success() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

print_warning() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

print_error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
}

# Detect operating system and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case "$OS" in
        Linux*)     PLATFORM="linux" ;;
        Darwin*)    PLATFORM="macOS" ;;
        FreeBSD*)   PLATFORM="freebsd" ;;
        OpenBSD*)   PLATFORM="openbsd" ;;
        NetBSD*)    PLATFORM="netbsd" ;;
        *)          
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        armv6l)         ARCH="armv6" ;;
        i386|i686)      ARCH="386" ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    print_info "Detected platform: $PLATFORM ($ARCH)"
}

# Get the latest release version
get_latest_version() {
    print_info "Fetching latest version..."
    
    # Try to get version from GitHub API
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    # Fallback: try to scrape from releases page if API fails
    if [ -z "$VERSION" ]; then
        print_warning "GitHub API rate limit reached, trying alternative method..."
        if command -v curl >/dev/null 2>&1; then
            VERSION=$(curl -fsSL "https://github.com/$REPO/releases/latest" 2>/dev/null | grep -oP 'releases/tag/v\K[0-9.]+' | head -n1)
        elif command -v wget >/dev/null 2>&1; then
            VERSION=$(wget -qO- "https://github.com/$REPO/releases/latest" 2>/dev/null | grep -oP 'releases/tag/v\K[0-9.]+' | head -n1)
        fi
    fi
    
    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version. Please specify VERSION environment variable."
        print_info "Example: VERSION=2.62.0 $0"
        exit 1
    fi
    
    print_success "Latest version: v$VERSION"
}

# Download and extract the binary
download_and_install() {
    TARBALL="gh_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    URL="https://github.com/$REPO/releases/download/v$VERSION/$TARBALL"
    TMPDIR=$(mktemp -d)
    
    print_info "Downloading $TARBALL..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "$TMPDIR/$TARBALL"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$URL" -O "$TMPDIR/$TARBALL"
    fi
    
    if [ ! -f "$TMPDIR/$TARBALL" ]; then
        print_error "Failed to download $TARBALL"
        rm -rf "$TMPDIR"
        exit 1
    fi
    
    print_info "Extracting archive..."
    tar -xzf "$TMPDIR/$TARBALL" -C "$TMPDIR"
    
    # Find the extracted directory (format: gh_VERSION_PLATFORM_ARCH)
    EXTRACT_DIR="$TMPDIR/gh_${VERSION}_${PLATFORM}_${ARCH}"
    
    if [ ! -d "$EXTRACT_DIR" ]; then
        print_error "Extraction failed or unexpected directory structure"
        rm -rf "$TMPDIR"
        exit 1
    fi
    
    # Create installation directories
    mkdir -p "$PREFIX/bin"
    mkdir -p "$PREFIX/share/man/man1"
    
    # Install binary
    print_info "Installing to $PREFIX/bin..."
    cp "$EXTRACT_DIR/bin/gh" "$PREFIX/bin/"
    chmod +x "$PREFIX/bin/gh"
    
    # Install man pages if they exist
    if [ -d "$EXTRACT_DIR/share/man/man1" ]; then
        print_info "Installing man pages..."
        cp "$EXTRACT_DIR/share/man/man1"/* "$PREFIX/share/man/man1/" 2>/dev/null || true
    fi
    
    # Cleanup
    rm -rf "$TMPDIR"
    
    print_success "GitHub CLI installed successfully!"
}

# Check if gh is already installed
check_existing() {
    if command -v gh >/dev/null 2>&1; then
        INSTALLED_VERSION=$(gh version 2>/dev/null | head -n1 | awk '{print $3}' | sed 's/v//')
        print_warning "GitHub CLI is already installed (version: $INSTALLED_VERSION)"
        printf "Do you want to continue with installation? [y/N] "
        read -r response
        case "$response" in
            [yY][eE][sS]|[yY]) 
                return 0
                ;;
            *)
                print_info "Installation cancelled"
                exit 0
                ;;
        esac
    fi
}

# Verify installation
verify_installation() {
    if ! command -v "$PREFIX/bin/gh" >/dev/null 2>&1; then
        print_error "Installation verification failed"
        exit 1
    fi
    
    INSTALLED_VERSION=$("$PREFIX/bin/gh" version 2>/dev/null | head -n1 || echo "unknown")
    print_success "Verification successful!"
    print_info "Installed version: $INSTALLED_VERSION"
}

# Display post-installation instructions
show_instructions() {
    echo ""
    print_success "Installation complete!"
    echo ""
    print_info "Add the following to your shell profile (e.g., ~/.bashrc, ~/.zshrc):"
    echo ""
    echo "    export PATH=\"$PREFIX/bin:\$PATH\""
    echo ""
    
    # Check if PREFIX/bin is in PATH
    case ":$PATH:" in
        *:"$PREFIX/bin":*)
            print_success "$PREFIX/bin is already in your PATH"
            ;;
        *)
            print_warning "$PREFIX/bin is not in your PATH"
            print_info "Run the following command to add it to your current session:"
            echo ""
            echo "    export PATH=\"$PREFIX/bin:\$PATH\""
            echo ""
            ;;
    esac
    
    print_info "To get started with GitHub CLI, run:"
    echo ""
    echo "    gh auth login"
    echo ""
}

# Main installation flow
main() {
    echo ""
    print_info "GitHub CLI Installation Script"
    echo ""
    
    check_existing
    detect_platform
    get_latest_version
    download_and_install
    verify_installation
    show_instructions
}

# Run main installation
main
