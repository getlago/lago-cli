#!/bin/sh
# Verify every dependency carries a license compatible with this repository's MIT
# license. The repo shipped gosec and govulncheck gates but no license gate, so a
# copyleft dependency could arrive unnoticed in a PR.
set -eu

# Permissive licenses compatible with redistribution under MIT.
allowed='MIT|Apache-2.0|BSD-2-Clause|BSD-3-Clause|ISC|MPL-2.0|Unlicense|CC0-1.0'

report=$(mktemp)
trap 'rm -f "$report"' EXIT HUP INT TERM

go run github.com/google/go-licenses@v1.6.0 report ./... 2>/dev/null > "$report" || {
  echo "license audit: go-licenses failed to produce a report" >&2
  exit 1
}
[ -s "$report" ] || { echo "license audit: empty report" >&2; exit 1; }

failed=0
while IFS=, read -r module _ license; do
  [ -n "$module" ] || continue
  if ! printf '%s' "$license" | grep -qE "^($allowed)$"; then
    echo "license audit: $module is $license, which is not MIT-compatible" >&2
    failed=1
  fi
done < "$report"

if [ "$failed" -eq 0 ]; then
  echo "license audit: $(wc -l < "$report" | tr -d ' ') dependencies, all MIT-compatible"
fi
exit "$failed"
