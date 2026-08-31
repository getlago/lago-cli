SHELL := /bin/sh

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || printf unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X github.com/getlago/lago-cli/internal/cli.version=$(VERSION)

.PHONY: build test test-e2e-compile coverage generate generate-check docs lint security release clean

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/lago ./cmd/lago

test:
	go test -race ./...

# The golden-invoice harness is build-tagged and only executes against trusted
# staging credentials. Compile it on every PR so it cannot rot silently.
test-e2e-compile:
	go vet -tags e2e ./test/...

coverage:
	./scripts/check-coverage.sh

generate:
	go run ./internal/gen -spec spec/openapi.yaml -out internal/generated/spec_gen.go -manifest spec/operations.json
	go run ./internal/docgen -markdown docs/reference -man man -completions completions

generate-check:
	go run ./internal/gen -check -spec spec/openapi.yaml -out internal/generated/spec_gen.go -manifest spec/operations.json
	go run ./internal/docgen -markdown docs/reference -man man -completions completions
	git diff --exit-code -- docs/reference man completions

docs:
	go run ./internal/docgen -markdown docs/reference -man man -completions completions

lint:
	test -z "$$(gofmt -l cmd internal pkg)"
	go vet ./...
	go run ./internal/moneycheck ./...
	go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12

security:
	./scripts/security-check.sh

release:
	goreleaser release --clean

clean:
	rm -f bin/lago coverage.out
