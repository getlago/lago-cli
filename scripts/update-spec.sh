#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmp_file=$(mktemp)
tmp_file_v2=$(mktemp)
trap 'rm -f "$tmp_file" "$tmp_file_v2"' EXIT HUP INT TERM

curl -fsSL --proto '=https' --tlsv1.2 https://swagger.getlago.com/openapi.yaml -o "$tmp_file"
curl -fsSL --proto '=https' --tlsv1.2 https://swagger.getlago.com/openapi-v2.yaml -o "$tmp_file_v2"
for file in "$tmp_file" "$tmp_file_v2"; do
  grep -q '^openapi: 3\.' "$file"
  grep -q '^paths:' "$file"
done
mv "$tmp_file" "$repo_dir/spec/openapi.yaml"
mv "$tmp_file_v2" "$repo_dir/spec/openapi-v2.yaml"
trap - EXIT HUP INT TERM
cd "$repo_dir"
go run ./internal/gen
go run ./internal/gen -api-version v2 -spec spec/openapi-v2.yaml -out internal/generated/spec_v2_gen.go -manifest spec/operations-v2.json
