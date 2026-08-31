# Architecture

Lago CLI is a single static Go binary. Its boundary packages are deliberately small: `internal/config` resolves and securely persists profiles; `internal/transport` owns TLS, retries, redaction, errors, and timing; `internal/output` owns stable table/JSON/YAML rendering; `internal/cli` composes built-ins and generated operations; and `pkg/api` contains public exact-value types.

## OpenAPI generation

`spec/openapi.yaml` is a pinned copy of `https://swagger.getlago.com/openapi.yaml`. It is the only source for API paths, methods, parameters, body fields, types, enums, and descriptions.

```text
spec/openapi.yaml
       |
       v
internal/gen -----> internal/generated/spec_gen.go
       |----------> spec/operations.json
       |
command tree ------> docs/reference + man pages
```

`make generate` is the only regeneration entry point. The generator resolves local references and `allOf`, creates one descriptor for every path/verb operation, and embeds the spec version and SHA-256 in the binary. `make generate-check` reconstructs the output in memory and fails on drift.

An operation may declare `x-lago-cli-action` to choose its action name. When the extension is absent, the generator uses path, operation ID, and summary conventions deterministically. A remaining collision is a generation error and must be fixed in the upstream spec—not in a CLI mapping file.

PR CI never fetches a mutable network resource. It validates the pinned snapshot. A daily trusted workflow fetches the live document, regenerates, and opens a reviewable drift PR. Release candidates cannot ship with unresolved spec drift.

## Request lifecycle

Flags override `LAGO_*` environment variables, which override the selected TOML profile. A request then passes through live-mode guardrails, body generation, exact-value encoding, authorization injection, redacted diagnostics, bounded retries, response classification, optional JMESPath, and output rendering.

Retries are limited to network errors, 429, and 5xx. They occur only for an idempotent method or when an idempotency key exists. `Retry-After` takes precedence over exponential full-jitter backoff. The HTTP client has a five-second connect/TLS budget and a 30-second total default.

Bulk events use a bounded worker queue. NDJSON lines are decoded independently, receive a transaction ID if missing, and are sent with matching idempotency keys. The full input is never retained in memory.

## Trust boundaries

The config file is the sole location where an API key may be persisted and uses mode 0600 on Unix. API keys are never included in generated docs, fixtures, diagnostics, or dry-run output. Cross-origin redirects are refused. TLS 1.2+ is required unless a user explicitly selects `--insecure` for a self-hosted instance.
