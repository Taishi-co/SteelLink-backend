#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/backend}"
CONTAINER_NAME="${CONTAINER_NAME:-backend}"
IMAGE_NAME="${IMAGE_NAME:-my-backend}"
HOST_PORT="${HOST_PORT:-3000}"
CONTAINER_PORT="${CONTAINER_PORT:-8080}"
ENV_FILE="${ENV_FILE:-${APP_DIR}/.env}"

cd "${APP_DIR}"

if [ ! -f "${ENV_FILE}" ]; then
  echo "missing env file: ${ENV_FILE}" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "${ENV_FILE}"
set +a

required_env=(
  DATABASE_HOST
  DATABASE_PORT
  DATABASE_USERNAME
  DATABASE_PASSWORD
  DATABASE_NAME
  REDIS_HOST
  REDIS_PORT
  JWT_SECRET
  COS_SECRET_ID
  COS_SECRET_KEY
)

for name in "${required_env[@]}"; do
  if [ -z "${!name:-}" ]; then
    echo "missing required env: ${name}" >&2
    exit 1
  fi
done

if [ "${DATABASE_HOST}" = "localhost" ] || [ "${DATABASE_HOST}" = "127.0.0.1" ]; then
  echo "warning: DATABASE_HOST=${DATABASE_HOST} points inside the container."
  echo "if MySQL runs on the Docker host, set DATABASE_HOST=host.docker.internal in ${ENV_FILE}."
fi

if [ "${REDIS_HOST}" = "localhost" ] || [ "${REDIS_HOST}" = "127.0.0.1" ]; then
  echo "warning: REDIS_HOST=${REDIS_HOST} points inside the container."
  echo "if Redis runs on the Docker host, set REDIS_HOST=host.docker.internal in ${ENV_FILE}."
fi

git fetch origin
git switch main
git pull --ff-only

docker stop "${CONTAINER_NAME}" || true
docker rm "${CONTAINER_NAME}" || true

docker build -t "${IMAGE_NAME}" .

docker run -d \
  --name "${CONTAINER_NAME}" \
  --restart always \
  -p "${HOST_PORT}:${CONTAINER_PORT}" \
  --env-file "${ENV_FILE}" \
  --add-host host.docker.internal:host-gateway \
  "${IMAGE_NAME}"

echo "waiting for backend health check..."
for _ in $(seq 1 30); do
  if docker exec "${CONTAINER_NAME}" wget -qO- "http://127.0.0.1:${CONTAINER_PORT}/health" >/dev/null 2>&1; then
    docker image prune -f
    echo "backend deployed on host port ${HOST_PORT}"
    exit 0
  fi
  sleep 1
done

echo "backend failed health check; recent container logs:" >&2
docker logs --tail 120 "${CONTAINER_NAME}" >&2 || true
exit 1
