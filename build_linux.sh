#!/usr/bin/env bash
set -euo pipefail

LDFLAGS="-s -w"
OUTPUT="godot"

echo "Building ${OUTPUT}..."
go build -ldflags "$LDFLAGS" -trimpath -o "$OUTPUT" ./cmd/godot

ls -lh "$OUTPUT"
echo "Build complete."
