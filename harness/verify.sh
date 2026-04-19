#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "${BUILD_DIR}"
}

trap cleanup EXIT

cd "${ROOT_DIR}"

echo "== Verify =="
echo "[1/3] go test ./..."
go test ./...
echo
echo "[2/3] CGO_ENABLED=0 go build ./cmd/api"
CGO_ENABLED=0 go build -o "${BUILD_DIR}/steellink-backend" ./cmd/api
echo
echo "[3/3] docker compose config (optional)"
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  docker compose config >/dev/null
  echo "docker compose config passed"
else
  echo "docker compose not available, skipped"
fi
echo
echo "verify passed"
