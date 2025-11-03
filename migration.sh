#!/usr/bin/env bash
set -euo pipefail

# Determine repo root (directory of this script)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"

cd "$REPO_ROOT"

PYTHON_BIN="python3"
VENV_DIR="$REPO_ROOT/.venv"

echo "[migration] Using repo: $REPO_ROOT"

if [ ! -d "$VENV_DIR" ]; then
  echo "[migration] Creating virtualenv at $VENV_DIR"
  "$PYTHON_BIN" -m venv "$VENV_DIR"
fi

source "$VENV_DIR/bin/activate"

echo "[migration] Installing requirements"
pip install --upgrade pip >/dev/null
pip install -r "$REPO_ROOT/requirements.txt"

# Allow override; default to sqlite database in repo root
export DATABASE_URL="${DATABASE_URL:-sqlite:///$REPO_ROOT/test.db}"
echo "[migration] DATABASE_URL=$DATABASE_URL"

ALEMBIC_CFG="$REPO_ROOT/migration/alembic.ini"
VERSIONS_DIR="$REPO_ROOT/migration/alembic/versions"

mkdir -p "$VERSIONS_DIR"

# If arguments are provided, pass them through to Alembic
if [ $# -gt 0 ]; then
  echo "[migration] Running: alembic $*"
  exec alembic -c "$ALEMBIC_CFG" "$@"
fi

# Default behavior: ensure initial revision and upgrade head
if [ -z "$(ls -A "$VERSIONS_DIR" 2>/dev/null)" ]; then
  echo "[migration] Creating initial autogenerate revision"
  alembic -c "$ALEMBIC_CFG" revision --autogenerate -m "init schema"
fi

echo "[migration] Applying migrations"
alembic -c "$ALEMBIC_CFG" upgrade head

echo "[migration] Done. DB at: ${DATABASE_URL#sqlite:///}"


