#!/bin/sh
set -eu

paths="testdata"
[ -d "$paths" ] || exit 0

if grep -REn --exclude='*.md' '(sk|pk|lago)_(live|test)_[A-Za-z0-9]{20,}|[A-Za-z0-9._%+-]+@(gmail|outlook|yahoo)\.com' $paths; then
  echo "fixture scan: possible secret or real-looking personal email detected" >&2
  exit 1
fi

echo "fixture scan: clean"
