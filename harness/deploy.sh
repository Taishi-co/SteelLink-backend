#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_ENV="${1:-}"
ENV_FILE="${ENV_FILE:-.env.${TARGET_ENV}}"

if [[ "${TARGET_ENV}" != "staging" && "${TARGET_ENV}" != "prod" ]]; then
  echo "usage: ./harness/deploy.sh <staging|prod>" >&2
  exit 1
fi

cd "${ROOT_DIR}"

echo "== Deploy Plan =="
echo "target environment: ${TARGET_ENV}"

if [[ ! -f "Dockerfile" || ! -f "docker-compose.yml" ]]; then
  echo "deploy prerequisites missing: Dockerfile or docker-compose.yml" >&2
  exit 1
fi

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  if [[ -f "${ENV_FILE}" ]]; then
    docker compose --env-file "${ENV_FILE}" config >/dev/null
    echo "validated compose plan with ${ENV_FILE}"
  else
    docker compose config >/dev/null
    echo "validated compose plan with default environment"
  fi
else
  echo "docker compose not available; deployment plan not validated locally"
fi

if [[ "${DEPLOY_EXECUTE:-0}" != "1" ]]; then
  echo "dry-run mode"
  if [[ -f "${ENV_FILE}" ]]; then
    echo "planned command: docker compose --env-file ${ENV_FILE} up -d --build api"
  else
    echo "planned command: docker compose up -d --build api"
  fi
  echo "set DEPLOY_EXECUTE=1 to enable real deployment steps"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
  echo "docker compose is required when DEPLOY_EXECUTE=1" >&2
  exit 1
fi

if [[ -f "${ENV_FILE}" ]]; then
  docker compose --env-file "${ENV_FILE}" up -d --build api
else
  docker compose up -d --build api
fi

echo "deployment command executed"
