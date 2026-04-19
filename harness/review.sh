#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FEATURE_DIR="${1:-}"

if [[ -z "${FEATURE_DIR}" ]]; then
  FEATURE_DIR="$("${ROOT_DIR}/harness/resolve-feature-dir.sh")"
fi

cd "${ROOT_DIR}"

for file in spec.md test-cases.md implementation-plan.md; do
  if [[ ! -f "${FEATURE_DIR}/${file}" ]]; then
    echo "missing required review artifact: ${FEATURE_DIR}/${file}" >&2
    exit 1
  fi
done

echo "== Review Gate =="
echo "feature: ${FEATURE_DIR}"
./harness/verify.sh
echo
echo "review gate passed"
