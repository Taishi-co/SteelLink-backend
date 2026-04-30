#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FEATURE_DIR=""
PR_URL="${PR_URL:-}"
PR_TITLE="${PR_TITLE:-}"
DRY_RUN=0

INDEX_DOC="${LARK_PR_INDEX_DOC:-https://hwdz01k44ey.feishu.cn/wiki/OC6Owz2b3ijTuOkeW8NciN6onBe}"
PARENT_WIKI_NODE="${LARK_PR_WIKI_NODE:-OC6Owz2b3ijTuOkeW8NciN6onBe}"

usage() {
  cat <<'USAGE'
Usage:
  ./harness/lark-pr-doc.sh [docs/features/<feature-id>] [--pr-url <url>] [--pr-title <title>] [--dry-run]

Creates a Feishu/Lark PR document from the feature PRD artifacts, then inserts
the generated link under the repository section in the shared PR index.

Environment overrides:
  LARK_PR_INDEX_DOC   Shared PR index document URL or token
  LARK_PR_WIKI_NODE   Parent wiki node token for newly created PR docs
  PR_URL              GitHub PR URL, if not passed as --pr-url
  PR_TITLE            GitHub PR title, if not passed as --pr-title
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --pr-url)
      PR_URL="${2:-}"
      shift 2
      ;;
    --pr-title)
      PR_TITLE="${2:-}"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      if [[ -z "${FEATURE_DIR}" ]]; then
        FEATURE_DIR="$1"
        shift
      else
        echo "unexpected argument: $1" >&2
        usage >&2
        exit 1
      fi
      ;;
  esac
done

if ! command -v lark-cli >/dev/null 2>&1; then
  echo "lark-cli is required but was not found" >&2
  exit 1
fi

if ! command -v node >/dev/null 2>&1; then
  echo "node is required to parse lark-cli output" >&2
  exit 1
fi

if [[ -z "${FEATURE_DIR}" ]]; then
  FEATURE_DIR="$("${ROOT_DIR}/harness/resolve-feature-dir.sh")"
fi

cd "${ROOT_DIR}"

if [[ ! -d "${FEATURE_DIR}" ]]; then
  echo "feature directory not found: ${FEATURE_DIR}" >&2
  exit 1
fi

for file in spec.md test-cases.md implementation-plan.md; do
  if [[ ! -f "${FEATURE_DIR}/${file}" ]]; then
    echo "missing required PRD artifact: ${FEATURE_DIR}/${file}" >&2
    exit 1
  fi
done

REPO_NAME="$(basename "${ROOT_DIR}")"
case "${REPO_NAME}" in
  SteelLink-admin)
    INDEX_SECTION="Admin"
    ;;
  SteelLink-backend)
    INDEX_SECTION="Backend"
    ;;
  SteelLink-frontend)
    INDEX_SECTION="Frontend"
    ;;
  *)
    echo "unable to map repository to Feishu section: ${REPO_NAME}" >&2
    exit 1
    ;;
esac

FEATURE_ID="$(basename "${FEATURE_DIR}")"
BRANCH_NAME="$(git branch --show-current 2>/dev/null || true)"

if [[ -z "${PR_TITLE}" ]]; then
  PR_TITLE="${FEATURE_ID}"
fi

DOC_TITLE="${REPO_NAME} ${FEATURE_ID} PRD / PR Record"
TMP_MARKDOWN="$(mktemp "${ROOT_DIR}/.lark-pr-doc.XXXXXX.md")"
TMP_MARKDOWN_ARG="@./$(basename "${TMP_MARKDOWN}")"
trap 'rm -f "${TMP_MARKDOWN}"' EXIT

append_file() {
  local title="$1"
  local path="$2"

  {
    printf '\n## %s\n\n' "${title}"
    cat "${path}"
    printf '\n'
  } >>"${TMP_MARKDOWN}"
}

{
  printf '# %s\n\n' "${DOC_TITLE}"
  printf '## PR Metadata\n\n'
  printf -- '- Repository: `%s`\n' "${REPO_NAME}"
  printf -- '- Feature ID: `%s`\n' "${FEATURE_ID}"
  printf -- '- Branch: `%s`\n' "${BRANCH_NAME:-unknown}"
  if [[ -n "${PR_URL}" ]]; then
    printf -- '- GitHub PR: [%s](%s)\n' "${PR_TITLE}" "${PR_URL}"
  else
    printf -- '- GitHub PR: not provided\n'
  fi
  printf -- '- Source feature directory: `%s`\n' "${FEATURE_DIR}"
} >"${TMP_MARKDOWN}"

append_file "Spec" "${FEATURE_DIR}/spec.md"
append_file "Test Cases" "${FEATURE_DIR}/test-cases.md"
append_file "Implementation Plan" "${FEATURE_DIR}/implementation-plan.md"

if [[ -f "${FEATURE_DIR}/release-checklist.md" ]]; then
  append_file "Release Checklist" "${FEATURE_DIR}/release-checklist.md"
fi

CREATE_ARGS=(docs +create --title "${DOC_TITLE}" --markdown "${TMP_MARKDOWN_ARG}" --wiki-node "${PARENT_WIKI_NODE}")

if [[ "${DRY_RUN}" -eq 1 ]]; then
  echo "== Dry run: create Feishu PR document =="
  lark-cli "${CREATE_ARGS[@]}" --dry-run
  echo
  echo "== Dry run: insert link into shared PR index =="
  lark-cli docs +update \
    --doc "${INDEX_DOC}" \
    --mode insert_after \
    --selection-by-title "# ${INDEX_SECTION}" \
    --markdown "- [${DOC_TITLE}](https://example.feishu.cn/wiki/generated-doc-url) - PR: ${PR_URL:-not provided}" \
    --dry-run
  exit 0
fi

CREATE_OUTPUT="$(lark-cli "${CREATE_ARGS[@]}")"
DOC_URL="$(
  CREATE_OUTPUT="${CREATE_OUTPUT}" node <<'NODE'
const input = process.env.CREATE_OUTPUT || "";
let data;
try {
  data = JSON.parse(input);
} catch {
  process.exit(1);
}

const urls = [];
function walk(value) {
  if (typeof value === "string") {
    if (/https:\/\/[^\s"]*(feishu|larksuite)[^\s"]*/i.test(value)) {
      urls.push(value);
    }
    return;
  }
  if (Array.isArray(value)) {
    value.forEach(walk);
    return;
  }
  if (value && typeof value === "object") {
    Object.values(value).forEach(walk);
  }
}
walk(data);
if (urls.length === 0) {
  process.exit(1);
}
console.log(urls[0]);
NODE
)" || {
  echo "failed to parse created Feishu document URL from lark-cli output" >&2
  echo "${CREATE_OUTPUT}" >&2
  exit 1
}

INDEX_LINE="- [${DOC_TITLE}](${DOC_URL})"
if [[ -n "${PR_URL}" ]]; then
  INDEX_LINE="${INDEX_LINE} - PR: [${PR_TITLE}](${PR_URL})"
fi

lark-cli docs +update \
  --doc "${INDEX_DOC}" \
  --mode insert_after \
  --selection-by-title "# ${INDEX_SECTION}" \
  --markdown "${INDEX_LINE}"

echo "created Feishu PR document: ${DOC_URL}"
echo "updated shared PR index section: ${INDEX_SECTION}"
