#!/bin/sh
# Scan every tracked file for credentials and real-world identities before this
# repository can be made public.
#
# The previous version scanned only testdata/, missing internal/cli/fixtures/,
# internal/cli/testdata/ and docs/reference/, and its key pattern required a
# `lago_live_` prefix that real Lago API keys (UUID-shaped) never carry.
#
# Findings are compared against scripts/fixture-allowlist.txt, which records the
# accepted exceptions with a reason. Anything not on that list fails the build.
set -eu

allowlist=${1:-scripts/fixture-allowlist.txt}
findings=$(mktemp)
trap 'rm -f "$findings" "$findings.err"' EXIT HUP INT TERM

# Lago API keys are UUID-shaped, so a bare UUID rule matches every resource-id example
# in the vendored spec and drowns the signal. The rule below only fires on a UUID in a
# credential context (api_key:, token:, Bearer ...), which is what a real leak looks like.
#
# Tracked files plus new files that are not ignored. Scanning only tracked files
# let a newly written test pass this check locally and fail it in CI.
#
# Binary and vendored-lockfile noise would otherwise dominate the report.
files=$(git ls-files --cached --others --exclude-standard \
  | grep -vE '^(go\.sum|.*\.(png|jpg|jpeg|gif|ico|pdf|gz|zip))$')
[ -n "$files" ] || { echo "fixture scan: no tracked files" >&2; exit 1; }

# shellcheck disable=SC2086
grep -nEH \
  -e '(sk|pk|lago)_(live|test)_[A-Za-z0-9]{16,}' \
  -e '(api[_-]?key|secret|token|password|authorization|bearer)[^0-9a-f]{0,20}[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' \
  -e '\b(gh[pousr]_[A-Za-z0-9]{20,}|AKIA[0-9A-Z]{16}|xox[baprs]-[A-Za-z0-9-]{10,})\b' \
  -e '-----BEGIN [A-Z ]*PRIVATE KEY-----' \
  -e '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.(com|io|net|org|co|dev|ai)' \
  -e 'https?://[A-Za-z0-9.-]*\.(internal|local|corp|intra)\b' \
  $files > "$findings" 2>"$findings.err" || true

# grep exits 1 for "no matches" and 2 for a bad pattern. A malformed pattern must
# fail the scan loudly rather than reporting a clean tree.
if [ -s "$findings.err" ]; then
  echo "fixture scan: pattern error, scan did not run:" >&2
  cat "$findings.err" >&2
  exit 1
fi

if [ ! -s "$findings" ]; then
  echo "fixture scan: clean"
  exit 0
fi

unexplained=$(mktemp)
trap 'rm -f "$findings" "$findings.err" "$unexplained" "$unexplained.rules"' EXIT HUP INT TERM
if [ -f "$allowlist" ]; then
  # An allowlist entry is a `path:pattern` prefix; strip comments and blank lines.
  grep -vE '^[[:space:]]*(#|$)' "$allowlist" > "$unexplained.rules" || true
  grep -vFf "$unexplained.rules" "$findings" > "$unexplained" || true
else
  cp "$findings" "$unexplained"
fi

if [ -s "$unexplained" ]; then
  echo "fixture scan: unexplained credential or real-identity matches:" >&2
  cat "$unexplained" >&2
  echo "" >&2
  echo "Replace these with example.com / *.test values, or add them to $allowlist with a reason." >&2
  exit 1
fi

echo "fixture scan: clean (all matches allowlisted)"
