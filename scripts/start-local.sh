#!/bin/bash
# Build and run yarr locally (no Docker). Requires a reachable MySQL instance.
#
# Config is loaded from deploy/.env (same file used by deploy-docker.sh) if
# present; env vars already set in the shell take precedence. Copy
# deploy/.env.example to deploy/.env and fill in YARR_DB first.
#
# Usage:
#   ./scripts/start-local.sh          # normal build, embedded assets
#   ./scripts/start-local.sh --debug  # serve assets from disk (backend/src/assets)
#                                      # for frontend iteration without rebuilding

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

cd backend
if [ "${1:-}" = "--debug" ]; then
  go build -tags debug -o out/yarr ./cmd/yarr
else
  go build -o out/yarr ./cmd/yarr
fi
exec ./out/yarr
