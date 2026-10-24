#!/usr/bin/env bash
# Adds the go-service-conventions block to ~/.claude/CLAUDE.md, or replaces it
# if already there, so every Claude Code session applies the conventions.
# Usage: pstack/install-global.sh [path-to-CLAUDE.md]
set -euo pipefail
target="${1:-$HOME/.claude/CLAUDE.md}"
snippet="$(cd "$(dirname "$0")" && pwd)/claude-md-snippet.md"
start="<!-- go-service-conventions:start -->"
end="<!-- go-service-conventions:end -->"
mkdir -p "$(dirname "$target")"
touch "$target"
tmp="$(mktemp)"
awk -v s="$start" -v e="$end" '$0==s{skip=1} !skip{print} $0==e{skip=0}' "$target" > "$tmp"
if [ -s "$tmp" ] && [ -n "$(tail -c1 "$tmp")" ]; then echo >> "$tmp"; fi
if [ -s "$tmp" ]; then echo >> "$tmp"; fi
cat "$snippet" >> "$tmp"
mv "$tmp" "$target"
echo "go-service-conventions block installed in $target"
