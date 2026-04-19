#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}"

module_name="$(awk '/^module / { print $2; exit }' go.mod 2>/dev/null || true)"
go_version="$(go version 2>/dev/null || echo 'go: not installed')"
make_version="$(make --version 2>/dev/null | head -n 1 || echo 'make: not installed')"
docker_version="$(
  (docker compose version 2>/dev/null || docker --version 2>/dev/null) | head -n 1 || true
)"

if [[ -z "${docker_version}" ]]; then
  docker_version="docker: not installed"
fi

echo "== SteelLink Backend Harness Bootstrap =="
echo "root: ${ROOT_DIR}"
if [[ -n "${module_name}" ]]; then
  echo "module: ${module_name}"
fi
echo "${go_version}"
echo "${make_version}"
echo "${docker_version}"

if git rev-parse --show-toplevel >/dev/null 2>&1; then
  branch="$(git symbolic-ref --quiet --short HEAD 2>/dev/null || echo 'DETACHED')"
  remote_url="$(git remote get-url origin 2>/dev/null || echo 'origin: not configured')"
  echo "git branch: ${branch}"
  echo "git remote: ${remote_url}"
else
  echo "git: current directory is not an initialized repository"
fi

echo
echo "Read in order:"
echo "1. AGENTS.md"
echo "2. docs/ai/entrypoint.md"
echo "3. docs/ai/project-map.md"
echo "4. docs/ai/engineering-rules.md"
echo "5. docs/ai/git-workflow.md"
echo "6. docs/ai/review-rules.md"
echo "7. docs/ai/release-rules.md"
echo
echo "Key commands:"
echo "- ./harness/branch-guard.sh local"
echo "- ./harness/verify.sh"
echo "- ./harness/review.sh docs/features/<feature-id>"
echo "- ./harness/release.sh docs/features/<feature-id>"
echo "- ./harness/deploy.sh staging"
echo
echo "Notes:"
echo "- verify follows the Go/Gin entrypoint at ./cmd/api"
echo "- branch-guard requires a valid Git worktree and current repository remote"
