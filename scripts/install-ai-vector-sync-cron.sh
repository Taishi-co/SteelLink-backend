#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${AI_PROCUREMENT_VECTOR_SYNC_INSTALL_DIR:-/opt/steellink/bin}"
ENV_FILE="${AI_PROCUREMENT_VECTOR_SYNC_ENV_FILE:-/etc/steellink/aiprocurement-vector-sync.env}"
LOG_FILE="${AI_PROCUREMENT_VECTOR_SYNC_LOG_FILE:-/var/log/steellink/aiprocurement-vector-sync.log}"
LOCK_FILE="${AI_PROCUREMENT_VECTOR_SYNC_LOCK_FILE:-/tmp/steellink-aiprocurement-vector-sync.lock}"
BIN_PATH="${INSTALL_DIR}/aiprocurement-vector-sync"
RUNNER_PATH="${INSTALL_DIR}/run-aiprocurement-vector-sync.sh"
CRON_MARKER="# steellink ai procurement vector sync"

mkdir -p "${INSTALL_DIR}" "$(dirname "${ENV_FILE}")" "$(dirname "${LOG_FILE}")"

cd "${ROOT_DIR}"
CGO_ENABLED=0 go build -o "${BIN_PATH}" ./cmd/aiprocurement-vector-sync

if [[ ! -f "${ENV_FILE}" ]]; then
  cat >"${ENV_FILE}" <<EOF
# Required.
AI_PROCUREMENT_VECTOR_SYNC_ENDPOINT=http://127.0.0.1:8080/api/v1/internal/ai-procurements/vector-sync
AI_PROCUREMENT_VECTOR_SYNC_TOKEN=

# Optional.
# STEELLINK_BACKEND_BASE_URL=http://127.0.0.1:8080
AI_PROCUREMENT_VECTOR_SYNC_BATCH_SIZE=100
AI_PROCUREMENT_VECTOR_SYNC_TARGET=all
EOF
  chmod 600 "${ENV_FILE}"
  echo "created ${ENV_FILE}; fill AI_PROCUREMENT_VECTOR_SYNC_TOKEN before relying on cron" >&2
fi

cat >"${RUNNER_PATH}" <<EOF
#!/usr/bin/env bash
set -euo pipefail
set -a
source "${ENV_FILE}"
set +a

flock -n "${LOCK_FILE}" "${BIN_PATH}" \\
  --schedule=manual \\
  --target="\${AI_PROCUREMENT_VECTOR_SYNC_TARGET:-all}" \\
  --batch-size="\${AI_PROCUREMENT_VECTOR_SYNC_BATCH_SIZE:-100}" \\
  >>"${LOG_FILE}" 2>&1
EOF
chmod 755 "${RUNNER_PATH}"

tmp_cron="$(mktemp)"
trap 'rm -f "${tmp_cron}"' EXIT
crontab -l 2>/dev/null | grep -vF "${CRON_MARKER}" >"${tmp_cron}" || true
{
  cat "${tmp_cron}"
  echo "*/30 * * * * ${RUNNER_PATH} ${CRON_MARKER}"
} | crontab -

echo "installed cron: */30 * * * * ${RUNNER_PATH}"
