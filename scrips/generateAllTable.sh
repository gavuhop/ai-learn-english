#!/usr/bin/env bash
set -euo pipefail

# Determine repo root (parent of this script directory)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT"

echo "[gen] Using repo: $REPO_ROOT"

# Basic checks
if [ ! -f "$REPO_ROOT/go.mod" ]; then
  echo "[gen] go.mod not found. Please run from the project root."
  exit 1
fi

if [ ! -f "$REPO_ROOT/config.yaml" ]; then
  echo "[gen] config.yaml not found in repo root. Create it before running."
  exit 1
fi

echo "[gen] Running generator: go run ./cmd/gen"
exec go run ./cmd/gen


