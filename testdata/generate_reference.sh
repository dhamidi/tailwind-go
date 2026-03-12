#!/usr/bin/env bash
#
# generate_reference.sh — Generate reference CSS files using the official tailwindcss CLI.
#
# Usage:
#   ./testdata/generate_reference.sh [classfile]
#
# If classfile is omitted, reads from testdata/classes.txt.
# Each line in the class file is a single utility class name.
# Output goes to testdata/reference/<encoded-classname>.css
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLASSFILE="${1:-$SCRIPT_DIR/classes.txt}"
REFDIR="$SCRIPT_DIR/reference"

if [ ! -f "$CLASSFILE" ]; then
  echo "Error: class file not found: $CLASSFILE" >&2
  exit 1
fi

mkdir -p "$REFDIR"

# Ensure tailwindcss CLI is available
if ! command -v npx &>/dev/null; then
  echo "Error: npx not found. Install Node.js and npm first." >&2
  exit 1
fi

# Create a temporary working directory
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

# Install tailwindcss in the temp directory
echo "Installing tailwindcss..."
(cd "$TMPDIR" && npm init -y >/dev/null 2>&1 && npm install tailwindcss >/dev/null 2>&1)

# Create the input CSS file with @import
cat > "$TMPDIR/input.css" <<'CSSEOF'
@import "tailwindcss";
CSSEOF

# URL-encode a class name for safe filesystem usage
urlencode() {
  local s="$1"
  s="${s//\%/%25}"
  s="${s//\//%2F}"
  s="${s//\[/%5B}"
  s="${s//\]/%5D}"
  s="${s//#/%23}"
  s="${s// /%20}"
  s="${s//!/%21}"
  s="${s//:/%3A}"
  s="${s//@/%40}"
  s="${s//./%2E}"
  echo "$s"
}

echo "Generating reference CSS files..."

while IFS= read -r class || [ -n "$class" ]; do
  # Skip empty lines and comments
  [[ -z "$class" || "$class" == \#* ]] && continue

  encoded="$(urlencode "$class")"
  outfile="$REFDIR/$encoded.css"

  # Create minimal HTML with the class
  cat > "$TMPDIR/input.html" <<HTMLEOF
<div class="$class"></div>
HTMLEOF

  # Run tailwindcss CLI to generate CSS
  if npx @tailwindcss/cli --input "$TMPDIR/input.css" --content "$TMPDIR/input.html" --output "$TMPDIR/output.css" 2>/dev/null; then
    cp "$TMPDIR/output.css" "$outfile"
    echo "  OK: $class -> $encoded.css"
  else
    echo "  FAIL: $class" >&2
  fi
done < "$CLASSFILE"

echo "Done. Reference files written to $REFDIR/"
