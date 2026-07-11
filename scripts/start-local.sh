#!/bin/bash
# Build and run yarr locally (no Docker): backend API + Vue3/Vite frontend together.
# Requires a reachable MySQL instance, Go, and Node.js/npm.
#
# Config is loaded from deploy/.env (same file used by deploy-docker.sh) if
# present; env vars already set in the shell take precedence. Copy
# deploy/.env.example to deploy/.env and fill in YARR_DB first.
#
# The backend serves the JSON API on 127.0.0.1:7070; the Vite dev server runs on
# localhost:5173 and proxies API/opml/page/logout requests to the backend (see
# frontend/vite.config.js). Open http://localhost:5173 in the browser.
#
# Usage:
#   ./scripts/start-local.sh          # backend + frontend dev server
#   ./scripts/start-local.sh --debug  # same, but backend serves its own embedded
#                                      # assets from disk (backend/src/assets)
#
# Press Ctrl+C to stop both processes.

set -euo pipefail
cd "$(dirname "$0")/.."

if [ -f deploy/.env ]; then
  set -a
  # shellcheck disable=SC1091
  source deploy/.env
  set +a
fi

if [ -z "${YARR_DB:-}" ]; then
  echo "YARR_DB is not set. Copy deploy/.env.example to deploy/.env and fill it in," >&2
  echo "or export YARR_DB yourself before running this script." >&2
  exit 1
fi

# Run the local backend unauthenticated. The cf-worker-auth SSO flow can't work
# through the Vite dev proxy (localhost:5173 is a different origin with no SSO
# session cookie, and the login redirect can't happen for XHR), so a configured
# YARR_AUTH_URL would make every /api call return 401 — which the frontend turns
# into a page reload, i.e. an infinite reload loop. Drop auth for local dev.
# Set LOCAL_AUTH=1 to keep it (only useful if you have another way to log in).
if [ "${LOCAL_AUTH:-}" != "1" ]; then
  unset YARR_AUTH_URL YARR_AUTH_SECRET
fi

# Build the backend (optionally with the debug tag for disk-served assets).
if [ "${1:-}" = "--debug" ]; then
  ( cd backend && go build -tags debug -o out/yarr ./cmd/yarr )
else
  ( cd backend && go build -o out/yarr ./cmd/yarr )
fi

# Install frontend dependencies on first run.
if [ ! -d frontend/node_modules ]; then
  echo "Installing frontend dependencies (first run)..."
  ( cd frontend && npm install )
fi

# Kill both child processes (and any workers they spawned) when this script
# exits — Ctrl+C, error, or one process dying on its own.
pids=()
cleanup() {
  trap - EXIT INT TERM
  echo ""
  echo "Shutting down..."
  for pid in "${pids[@]}"; do
    pkill -P "$pid" 2>/dev/null || true
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT INT TERM

# Start backend (exec so the subshell PID is the yarr binary itself).
( cd backend && exec ./out/yarr ) &
pids+=($!)

# Start the Vite dev server (invoke the binary directly, not via `npm run`, so
# killing this PID cleanly stops it and its esbuild workers).
( cd frontend && exec ./node_modules/.bin/vite ) &
pids+=($!)

echo ""
echo "  Backend API : http://127.0.0.1:7070"
echo "  Frontend    : http://localhost:5173   <- open this"
echo ""
echo "  Press Ctrl+C to stop both."
echo ""

# Stay alive while both processes run; exit (and let the trap tear down the
# survivor) as soon as either one dies. The short sleep keeps Ctrl+C responsive
# on older bash, where a pending trap runs as soon as `sleep` returns.
while kill -0 "${pids[0]}" 2>/dev/null && kill -0 "${pids[1]}" 2>/dev/null; do
  sleep 1
done
