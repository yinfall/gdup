#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

LDFLAGS="-s -w"
INSTALL_DIR="${HOME}/.gdup/bin"

mkdir -p "$INSTALL_DIR"

OUTPUT="${INSTALL_DIR}/gdup"

echo "Building gdup to ${INSTALL_DIR}..."
go build -ldflags "$LDFLAGS" -trimpath -o "$OUTPUT" ./cmd/gdup

ls -lh "$OUTPUT"
echo -e "\nBuild and install complete!"
echo "Make sure to add ${INSTALL_DIR} to your PATH."
echo "If you want to use the transparent 'godot' command, run: gdup shim install"
