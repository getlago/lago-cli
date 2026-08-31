#!/bin/sh
# Enforce the per-package coverage floors in coverage.floors.
#
# Fails when a package drops below its floor, when a package has no floor entry
# (so new packages cannot arrive untested and unnoticed), and reports packages
# that have risen far enough above their floor to be ratcheted.
set -eu

floors=${1:-coverage.floors}
report=$(mktemp)
trap 'rm -f "$report"' EXIT HUP INT TERM

go test -count=1 -cover ./internal/... ./pkg/... 2>/dev/null \
  | sed -n 's#^ok[[:space:]]*github.com/getlago/lago-cli/\([^[:space:]]*\).*coverage: \([0-9.]*\)% of statements#\1 \2#p' \
  > "$report"

[ -s "$report" ] || { echo "coverage: no packages reported" >&2; exit 1; }

failed=0
while read -r package actual; do
  floor=$(awk -v p="$package" '$1 == p { print $2 }' "$floors")
  if [ -z "$floor" ]; then
    echo "coverage: $package has no floor in $floors; add one" >&2
    failed=1
    continue
  fi
  if awk -v a="$actual" -v f="$floor" 'BEGIN { exit !(a + 0 < f + 0) }'; then
    echo "coverage: $package at ${actual}% is below its ${floor}% floor" >&2
    failed=1
  elif awk -v a="$actual" -v f="$floor" 'BEGIN { exit !(a + 0 >= f + 5) }'; then
    echo "coverage: $package rose to ${actual}%; raise its floor from ${floor}%"
  else
    echo "coverage: $package ${actual}% (floor ${floor}%)"
  fi
done < "$report"

exit "$failed"
