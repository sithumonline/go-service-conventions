#!/usr/bin/env bash
# Installs the go-service-conventions project playbook into a repo that uses pstack.
# Usage: pstack/install.sh <project-root>
set -euo pipefail
root="${1:?usage: install.sh <project-root>}"
here="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$root/.agents/playbooks"
cp "$here/go-service.md" "$root/.agents/playbooks/go-service.md"
checker="$(ls -d "$HOME"/.claude/plugins/cache/*/pstack/*/skills/poteto-mode/scripts/check-playbooks.mjs 2>/dev/null | sort -V | tail -1 || true)"
if [ -n "$checker" ]; then
  node "$checker" "$root"
else
  echo "pstack's check-playbooks.mjs not found; run it from your pstack install to validate."
fi
