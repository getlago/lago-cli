#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

gitleaks_bin=${GITLEAKS_BIN:-gitleaks}
if ! command -v "$gitleaks_bin" >/dev/null 2>&1; then
  echo "security: gitleaks is required (set GITLEAKS_BIN to an audited binary)" >&2
  exit 1
fi

"$gitleaks_bin" git --redact --no-banner .
"$gitleaks_bin" dir --redact --no-banner .
go run ./internal/moneycheck ./...
go vet ./...
go run github.com/securego/gosec/v2/cmd/gosec@v2.22.10 -quiet -exclude-generated ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
./scripts/check-ci-policy.sh
./scripts/check-fixtures.sh
./scripts/check-licenses.sh

echo "security: local gates passed"
