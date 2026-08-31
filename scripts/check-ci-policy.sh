#!/bin/sh
set -eu

failed=0
for workflow in .github/workflows/*.yml .github/workflows/*.yaml; do
  [ -f "$workflow" ] || continue
  if grep -q 'pull_request_target:' "$workflow"; then
    echo "$workflow: pull_request_target is forbidden" >&2
    failed=1
  fi
  if grep -q 'pull_request:' "$workflow" && grep -q '\${{[[:space:]]*secrets\.' "$workflow"; then
    echo "$workflow: pull-request workflow references repository secrets" >&2
    failed=1
  fi
done

if [ -f .github/workflows/staging-e2e.yml ] && grep -q 'pull_request:' .github/workflows/staging-e2e.yml; then
  echo ".github/workflows/staging-e2e.yml: staging credentials must never run on pull requests" >&2
  failed=1
fi

exit "$failed"
