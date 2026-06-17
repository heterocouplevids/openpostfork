#!/usr/bin/env bash
# Pre-push gate: runs the same lint checks the CI release workflow runs.
#
# This is intentionally redundant with the pre-commit hooks (which
# only fire on staged files matching the hook's `files` regex) and
# with the CI workflow. The point is to catch lint failures on the
# developer's machine before they reach CI, so a failing release
# never happens because of a stale branch.
#
# Installed automatically by devenv on shell entry. See devenv.nix
# `enterShell` and AGENTS.md for the rationale.

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

remote="${1:-}"
url="${2:-}"

# Skip when explicitly disabled. The release workflow sets this when
# pushing the tag for a release that has already been CI-gated.
if [ "${OPENPOST_SKIP_PRE_PUSH_LINT:-0}" = "1" ]; then
  echo "pre-push-lint: skipped (OPENPOST_SKIP_PRE_PUSH_LINT=1)"
  exit 0
fi

# Skip tag pushes: the release workflow already gated the commit
# being tagged with the same lint suite.
if [ -n "$url" ] && echo "$url" | grep -qE "tags/"; then
  echo "pre-push-lint: tag push detected, skipping (CI already gated)"
  exit 0
fi

echo "pre-push-lint: running full lint suite..."

# Same checks the CI release workflow runs.
denv_lint() {
  if command -v devenv >/dev/null 2>&1; then
    devenv shell --quiet -- lint
  else
    # Fallback: run the underlying commands directly. Used when the
    # developer hasn't entered the devenv shell (e.g. CI machines).
    (
      cd backend && gofmt -l . | { read -r line && [ -n "$line" ] && { echo "$line"; exit 1; } || true; }
      cd backend && golangci-lint run ./...
      cd cli && golangci-lint run ./...
      cd frontend && bun run lint
    )
  fi
}

if ! denv_lint; then
  echo ""
  echo "pre-push-lint: FAILED. Fix the issues above, then push again."
  echo "Bypass for tag releases with: OPENPOST_SKIP_PRE_PUSH_LINT=1 git push"
  exit 1
fi

echo "pre-push-lint: OK"
