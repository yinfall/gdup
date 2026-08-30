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
    TMP_BIN="${INSTALL_DIR}/gdup.exe.tmp"
    FINAL_BIN="${INSTALL_DIR}/gdup.exe"
    OLD_BIN="${INSTALL_DIR}/gdup.exe.old"

    if ! curl -fsSL -o "$TMP_BIN" "$DOWNLOAD_URL"; then
        echo "Error: Failed to download binary."
        rm -f "$TMP_BIN"
        exit 1
    fi
    
    rm -f "$OLD_BIN" || true
    RENAMED_OLD=0
    if [ -f "$FINAL_BIN" ]; then
        if mv "$FINAL_BIN" "$OLD_BIN" 2>/dev/null; then
            RENAMED_OLD=1
        else
            echo "Error: Failed to rename running binary. Is it currently locked by the OS?"
            rm -f "$TMP_BIN"
            exit 1
        fi
    fi
    
    if ! mv "$TMP_BIN" "$FINAL_BIN" 2>/dev/null; then
        echo "Error: Failed to place the new executable."
        if [ "$RENAMED_OLD" = 1 ]; then
            mv "$OLD_BIN" "$FINAL_BIN" 2>/dev/null || true
        fi
        rm -f "$TMP_BIN"
        exit 1
    fi
    
    chmod +x "$FINAL_BIN" 2>/dev/null || true
else
    TMP_BIN="${INSTALL_DIR}/gdup.tmp"
    FINAL_BIN="${INSTALL_DIR}/gdup"

    if ! curl -fsSL -o "$TMP_BIN" "$DOWNLOAD_URL"; then
        echo "Error: Failed to download binary."
        rm -f "$TMP_BIN"
        exit 1
    fi
    chmod +x "$TMP_BIN"

    if ! mv "$TMP_BIN" "$FINAL_BIN"; then
        echo "Error: Failed to update binary."
        rm -f "$TMP_BIN"
        exit 1
    fi
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
