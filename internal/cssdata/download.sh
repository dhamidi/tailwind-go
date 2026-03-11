#!/usr/bin/env bash
# internal/cssdata/download.sh
# Run via: go generate (from package root)
# Downloads theme.css and preflight.css from the tailwindcss npm package.
# utilities.css is project-owned and must NOT be overwritten by this script.
VERSION=4.2.1
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Download the npm package (much smaller than the GitHub tarball)
(cd "$TMP" && npm pack "tailwindcss@${VERSION}" --quiet)
tar -C "$TMP" -xzf "$TMP/tailwindcss-${VERSION}.tgz"

# Copy only the data-driven CSS files
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cp "$TMP/package/theme.css"     "$SCRIPT_DIR/theme.css"
cp "$TMP/package/preflight.css" "$SCRIPT_DIR/preflight.css"

echo "Downloaded Tailwind CSS v${VERSION} (theme.css, preflight.css)"
