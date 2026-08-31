#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmp_file=$(mktemp)
trap 'rm -f "$tmp_file"' EXIT HUP INT TERM

curl -fsSL --proto '=https' --tlsv1.2 https://swagger.getlago.com/openapi.yaml -o "$tmp_file"
grep -q '^openapi: 3\.' "$tmp_file"
grep -q '^paths:' "$tmp_file"
mv "$tmp_file" "$repo_dir/spec/openapi.yaml"
trap - EXIT HUP INT TERM
cd "$repo_dir"
go run ./internal/gen
