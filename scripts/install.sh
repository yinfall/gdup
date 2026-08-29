#!/usr/bin/env bash
set -euo pipefail

# REPLACE THIS with your actual GitHub username/repo if different
REPO="yinfall/gdup"
INSTALL_DIR="$HOME/.gdup/bin"

echo "Detecting OS and Architecture..."
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

if [ "$OS" = "darwin" ]; then
    BINARY="gdup-darwin-${ARCH}"
elif [ "$OS" = "linux" ]; then
    BINARY="gdup-linux-${ARCH}"
elif echo "$OS" | grep -q "mingw\|msys\|cygwin"; then
    OS="windows"
    BINARY="gdup-windows-${ARCH}.exe"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

echo "Fetching latest release version from github.com/$REPO..."
RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
LATEST_TAG=$(curl -s "$RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Failed to fetch latest release. Check if the repo is public and has a release."
    exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY}"

echo "Downloading gdup ${LATEST_TAG} for ${OS}-${ARCH}..."
mkdir -p "$INSTALL_DIR"

if [ "$OS" = "windows" ]; then
    curl -L -o "${INSTALL_DIR}/gdup.exe" "$DOWNLOAD_URL"
    chmod +x "${INSTALL_DIR}/gdup.exe" 2>/dev/null || true
else
    curl -L -o "${INSTALL_DIR}/gdup" "$DOWNLOAD_URL"
    chmod +x "${INSTALL_DIR}/gdup"
fi

echo ""
echo "=========================================================="
echo " Success! gdup has been installed to:"
echo "   $INSTALL_DIR"
echo ""
echo " PLEASE ADD THIS DIRECTORY TO YOUR SYSTEM PATH!"
echo "=========================================================="
echo ""
echo "After updating your PATH, run 'gdup help' to get started."
