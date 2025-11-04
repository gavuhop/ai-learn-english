#!/usr/bin/env bash
set -euo pipefail

# Determine repo root (parent of this script directory)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)" 

cd "$REPO_ROOT"

PYTHON_BIN="python3"
VENV_DIR="$REPO_ROOT/.venv"

echo "[install] Using repo: $REPO_ROOT"

if [ ! -d "$VENV_DIR" ]; then
  echo "[install] Creating virtualenv at $VENV_DIR"
  "$PYTHON_BIN" -m venv "$VENV_DIR"
fi

# shellcheck disable=SC1090
source "$VENV_DIR/bin/activate"

echo "[install] Upgrading pip and installing requirements"
pip install --upgrade pip
pip install -r "$REPO_ROOT/requirements.txt"

echo "[install] Done. You can now run ./migration.sh"

