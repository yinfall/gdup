#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

LDFLAGS="-s -w"
INSTALL_DIR="${HOME}/.gdup/bin"

mkdir -p "$INSTALL_DIR"

OUTPUT="${INSTALL_DIR}/gdup"
OUTPUT_TMP="${INSTALL_DIR}/gdup.tmp"

echo "Building gdup to ${INSTALL_DIR}..."
if go build -ldflags "$LDFLAGS" -trimpath -o "$OUTPUT_TMP" ./cmd/gdup; then
    # Atomic replace to avoid 'Text file busy' on POSIX
    if ! mv "$OUTPUT_TMP" "$OUTPUT" 2>/dev/null; then
        # Fallback for Windows/MSYS where running files can't be directly overwritten
        rm -f "${OUTPUT}.old" 2>/dev/null || true
        mv "$OUTPUT" "${OUTPUT}.old" 2>/dev/null || true
        mv "$OUTPUT_TMP" "$OUTPUT"
    fi
else
    rm -f "$OUTPUT_TMP"
    exit 1
fi

ls -lh "$OUTPUT"
echo -e "\nBuild and install complete!"
echo "Make sure to add ${INSTALL_DIR} to your PATH."
echo "If you want to use the transparent 'godot' command, run: gdup shim install"
