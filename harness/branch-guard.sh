#!/usr/bin/env bash

set -euo pipefail

MODE="${1:-local}"
BRANCH_REGEX='^(feat|fix|docs|chore|refactor|test|hotfix)/[A-Za-z0-9._-]+$'
DEFAULT_BRANCH="${DEFAULT_BRANCH:-main}"

fail() {
  echo "branch guard failed: $1" >&2
  exit 1
}

case "${MODE}" in
  local)
    if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
      fail "current directory is not a git repository; initialize or clone the new repository before running branch guard"
    fi
    branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
    if [[ -z "${branch}" || "${branch}" == "HEAD" ]]; then
      fail "unable to determine current git branch"
    fi
    if [[ "${branch}" == "${DEFAULT_BRANCH}" || "${branch}" == "main" || "${branch}" == "master" ]]; then
      fail "do not develop directly on ${branch}; create a feature branch from latest ${DEFAULT_BRANCH}"
    fi
    if [[ ! "${branch}" =~ ${BRANCH_REGEX} ]]; then
      fail "branch '${branch}' must match ${BRANCH_REGEX}"
    fi
    if ! git remote get-url origin >/dev/null 2>&1; then
      echo "branch guard warning: origin remote is not configured yet for this repository" >&2
    fi
    echo "branch guard passed: ${branch}"
    ;;
  ci)
    if [[ "${GITHUB_EVENT_NAME:-}" != "pull_request" ]]; then
      echo "branch guard skipped: non-pull_request event"
      exit 0
    fi
    base="${GITHUB_BASE_REF:-}"
    head="${GITHUB_HEAD_REF:-}"
    if [[ "${base}" != "${DEFAULT_BRANCH}" ]]; then
      fail "pull request base branch must be ${DEFAULT_BRANCH}, got '${base}'"
    fi
    if [[ -z "${head}" || "${head}" == "${DEFAULT_BRANCH}" ]]; then
      fail "pull request head branch must be a feature branch"
    fi
    if [[ ! "${head}" =~ ${BRANCH_REGEX} ]]; then
      fail "pull request head branch '${head}' must match ${BRANCH_REGEX}"
    fi
    echo "branch guard passed: ${head} -> ${base}"
    ;;
  *)
    echo "usage: ./harness/branch-guard.sh <local|ci>" >&2
    exit 1
    ;;
esac
