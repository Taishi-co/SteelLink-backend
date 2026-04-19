#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_FEATURE_DIR="${DEFAULT_FEATURE_DIR:-}"

strip_wrapping() {
  local value="$1"
  value="${value#\`}"
  value="${value%\`}"
  value="${value#\"}"
  value="${value%\"}"
  printf '%s' "${value}"
}

print_if_valid() {
  local candidate="$1"
  if [[ -n "${candidate}" && -d "${ROOT_DIR}/${candidate}" ]]; then
    printf '%s\n' "${candidate}"
    exit 0
  fi
}

if [[ -n "${FEATURE_DIR:-}" ]]; then
  print_if_valid "$(strip_wrapping "${FEATURE_DIR}")"
fi

if [[ -n "${FEATURE_ID:-}" ]]; then
  print_if_valid "docs/features/$(strip_wrapping "${FEATURE_ID}")"
fi

if [[ -n "${GITHUB_EVENT_PATH:-}" && -f "${GITHUB_EVENT_PATH}" ]]; then
  parsed_feature_dir="$(grep -Eo 'docs/features/[A-Za-z0-9._/-]+' "${GITHUB_EVENT_PATH}" | head -n 1 || true)"
  print_if_valid "$(strip_wrapping "${parsed_feature_dir}")"
fi

mapfile -t feature_dirs < <(
  find "${ROOT_DIR}/docs/features" -mindepth 1 -maxdepth 1 -type d ! -name '_template' -exec basename {} \; | sort
)

if [[ "${#feature_dirs[@]}" -eq 1 ]]; then
  print_if_valid "docs/features/${feature_dirs[0]}"
fi

if [[ -n "${DEFAULT_FEATURE_DIR}" ]]; then
  print_if_valid "${DEFAULT_FEATURE_DIR}"
fi

echo "unable to resolve feature directory" >&2
exit 1
