#!/usr/bin/env bash
# internal/cssdata/download.sh
# Run via: go generate (from package root)
VERSION=4.2.1
URL="https://github.com/tailwindlabs/tailwindcss/archive/refs/tags/v${VERSION}.tar.gz"
TMP=$(mktemp -d)
curl -sSL -o "$TMP/tailwind.tar.gz" "$URL"
tar -C "$TMP" -xzf "$TMP/tailwind.tar.gz"

# Copy the relevant CSS files
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cp "$TMP/tailwindcss-${VERSION}/packages/tailwindcss/theme.css"     "$SCRIPT_DIR/theme.css"
cp "$TMP/tailwindcss-${VERSION}/packages/tailwindcss/utilities.css" "$SCRIPT_DIR/utilities.css"
cp "$TMP/tailwindcss-${VERSION}/packages/tailwindcss/preflight.css" "$SCRIPT_DIR/preflight.css"
cp "$TMP/tailwindcss-${VERSION}/packages/tailwindcss/index.css"     "$SCRIPT_DIR/index.css"
rm -rf "$TMP"
echo "Downloaded Tailwind CSS v${VERSION}"
