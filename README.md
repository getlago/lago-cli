# Lago CLI

The official command-line interface for [Lago](https://getlago.com), the open-source usage-based billing platform. The CLI is generated from Lago's OpenAPI specification and ships as one static Go binary.

> This repository is under active development. Until the first signed release, use a Lago test profile.

## Install

Two supported channels.

```console
$ brew install getlago/tap/lago
```

```console
$ go install github.com/getlago/lago-cli/cmd/lago@latest
```

<!-- TODO(public-repo): delete this subsection the day getlago/lago-cli goes public. -->
### While the repository is private

`go install` fetches through the Go module proxy, which cannot read a private repository. Without the two settings below it fails like this:

```
go: github.com/getlago/lago-cli/cmd/lago@latest: module github.com/getlago/lago-cli: git ls-remote -q origin in /Users/you/go/pkg/mod/cache/vcs/...: exit status 128:
	fatal: could not read Username for 'https://github.com': terminal prompts disabled
```

Tell Go to bypass the proxy for this module, and Git to reach GitHub over SSH:

```console
$ export GOPRIVATE=github.com/getlago/lago-cli
$ git config --global url."git@github.com:".insteadOf "https://github.com/"
$ go install github.com/getlago/lago-cli/cmd/lago@latest
```

The Homebrew tap is public and needs neither setting.

### What the two channels do not mean

This is a reduction in **channels**, not in platform support:

- Release archives are still built for macOS, Linux, and Windows on amd64 and arm64, and CI compiles and smoke-tests that full matrix on every pull request.
- `go install` works anywhere Go runs, Windows included.

Removed for 1.0: the shell and PowerShell installers, the container image, Scoop, and Winget. They are parked, with re-enable criteria, in [`dist-channels/parked/`](dist-channels/parked/README.md).

Neither channel self-updates. `lago upgrade` checks for a newer release and prints the command that matches how your binary was installed:

```console
$ lago upgrade
Lago CLI 1.1.0 is available (installed: 1.0.0).

    brew upgrade getlago/tap/lago
```

## Configure

Create an API key in the Lago app under **Developers → API keys**, then point the CLI at the right host.

| Deployment | `--api-url` |
| --- | --- |
| Cloud US | `https://api.getlago.com` |
| Cloud EU | `https://api.eu.getlago.com` |
| Self-hosted | your Lago API base URL, for example `https://lago.example.com` |

Pass the **base** URL. The CLI appends `/api/v1` itself. Pasting the full path works too: `https://api.getlago.com/api/v1` normalizes to the same host, and so do a trailing slash, a custom port (`https://lago.example.com:8443`), and a sub-path behind a proxy (`https://tools.example.com/lago` keeps the sub-path). `init` saves the normalized URL, so what the profile holds is what the CLI calls.

What does not work is the **app** URL. `https://app.getlago.com` is the dashboard, not the API, and it answers, which is what makes the mistake confusing. The CLI refuses it by name rather than sending your API key to the frontend:

```console
$ lago init --api-key "$LAGO_API_KEY" --api-url https://app.eu.getlago.com
Error: app.eu.getlago.com is the Lago dashboard, not the Lago API
Suggestion: Use https://api.eu.getlago.com instead. The API and the app are different hosts, and an API key sent to the app host will not authenticate.
```

`init` writes the profile only after a live `GET /api/v1/organizations` accepts the key, so a saved profile is a validated one.

```console
$ lago init --api-key "$LAGO_API_KEY" --region us --mode test
Connected to Lago as Example Organization.
Saved us profile "default" to ~/.config/lago/config.toml (mode: test).
```

```console
$ lago init --api-key "$LAGO_API_KEY" --region eu --mode test
```

```console
$ lago init --api-key "$LAGO_API_KEY" --region self-hosted --api-url https://lago.example.com --mode test
```

`--region us` and `--region eu` are shorthand for the two URLs in the table above; they resolve to exactly the same normalized host as passing `--api-url` explicitly.

Check which host you are actually hitting, on any deployment. `RESOLVED_API_URL` is the base URL the CLI calls; `API_URL` is what the profile holds:

```console
$ lago whoami
NAME              Example Organization
LAGO_ID           org_...
DEFAULT_CURRENCY  EUR
TIMEZONE          Europe/Paris
PROFILE           default
MODE              test
RESOLVED_API_URL  https://api.eu.getlago.com/api/v1
```

`lago whoami --output json` carries the full organization object under `organization`, plus `profile`, `region`, `mode`, `api_url`, and `resolved_api_url`. `lago organizations get` is the same object without the profile fields.

`lago doctor` reports the same resolved URL as its own check, next to the configuration, permission, and authentication checks. It is the first line to paste into a support ticket: it separates "wrong credentials" from "right credentials, wrong host".

```console
$ lago doctor --output json --query 'checks[].{check: name, detail: detail}'
```

### Precedence and the live-mode default

Configuration resolves flags → environment → `~/.config/lago/config.toml`. Select a named profile with `--profile staging`.

| Setting | Flag | Environment |
| --- | --- | --- |
| API URL | `--api-url` | `LAGO_API_URL` |
| API key | `--api-key` | `LAGO_API_KEY` |
| Mode | `--mode` | `LAGO_MODE` |
| Profile | `--profile` | `LAGO_PROFILE` |
| Timeout | `--timeout` | `LAGO_TIMEOUT` |

**Mode defaults to live.** A profile declares `mode = "live"` or `mode = "test"`, and a credential override (`--api-key`, `--api-url`, `LAGO_API_KEY`, `LAGO_API_URL`) without an explicit mode deliberately resolves to **live**, not test. Failing toward caution means a script that forgets `--mode` gets the confirmation gates, not a silent write to production. Live commands print `[LIVE]` on stderr, and destructive live operations require the resource identifier via `--confirm` or typed interactive confirmation.

The gate follows the spec, not the command surface. `lago api` classifies a raw request by the operation its method and path address: in a live profile, `lago api DELETE /customers/x` or `lago api POST /invoices/x/void` requires `--confirm <path>` or typed confirmation, while a test profile keeps `lago api` as an ungated escape hatch. `lago fixtures run` and `lago seed demo` run only against test profiles, and a fixture containing a destructive step must be confirmed with `--confirm <fixture name>` before its first step runs.

Plain HTTP and disabled TLS verification require the explicit `--insecure` flag and always print a warning. Use it for `http://localhost:3000` during development, not against a deployment that holds real money. `lago init --insecure` persists `insecure = true` into the profile and says so on stderr; the next `lago init` for that profile without the flag clears it.

`lago init --profile <name>` writes that profile without switching `current_profile`. The first profile you create is current by default; after that, pass `--use` to switch, or `--profile <name>` per command. An alias (`lago alias set`) may name a profile but may not carry `--api-key`, `--api-url`, or `--insecure`: credentials and TLS choices live in the 0600 config file, not in an alias someone runs without reading it.

The config file is checked on every command. If it is readable by anyone but you, stderr says so and prints the `chmod 600` to run; `lago doctor` reports the same check.

## Quickstart

A metric, a plan, a customer, a subscription, one usage event, and the invoice preview.

```console
$ lago billable-metrics create --name Requests --code quickstart_requests --aggregation-type count_agg
LAGO_ID  0b8a1e70-1a90-4b2c-9f3d-5e6a7b8c9d01
CODE     quickstart_requests
NAME     Requests

$ lago plans create --input '{"plan":{"name":"Quickstart","code":"quickstart","interval":"monthly","amount_cents":0,"amount_currency":"USD","pay_in_advance":false,"charges":[{"billable_metric_id":"0b8a1e70-1a90-4b2c-9f3d-5e6a7b8c9d01","charge_model":"standard","properties":{"amount":"1"}}]}}'
LAGO_ID  2b902b90-2b90-2b90-2b90-2b902b902b90
CODE     quickstart
NAME     Quickstart

$ lago customers create --external-id quickstart_customer --name "Quickstart Customer"
LAGO_ID      1a901a90-1a90-1a90-1a90-1a901a901a90
EXTERNAL_ID  quickstart_customer
NAME         Quickstart Customer

$ lago subscriptions create \
    --external-customer-id quickstart_customer \
    --external-id quickstart_subscription \
    --plan-code quickstart
LAGO_ID      3c903c90-3c90-3c90-3c90-3c903c903c90
EXTERNAL_ID  quickstart_subscription
STATUS       active

$ lago events send --external-subscription-id quickstart_subscription --code quickstart_requests
CODE            quickstart_requests
LAGO_ID         5e905e90-5e90-5e90-5e90-5e905e905e90
TRANSACTION_ID  0f3c1d2e-4a5b-6c7d-8e9f-0a1b2c3d4e5f

$ lago invoices preview --input '{"customer":{"external_id":"quickstart_customer"},"subscriptions":{"external_ids":["quickstart_subscription"]}}'
CURRENCY            USD
FEES                1 item
ISSUING_DATE        2026-10-01
TOTAL_AMOUNT_CENTS  100
```

The `billable_metric_id` in the plan payload is the `LAGO_ID` the first command printed. Substitute yours.

### Writes print identifiers

Every write (`create`, `update`, `delete`, `terminate`, `apply`, `finalize`, `void`, metadata operations) prints a terse identifier block by default: `LAGO_ID`, `EXTERNAL_ID`, `CODE`, `NAME`, and `STATUS`, whichever the resource carries. After a create the one thing you do not already have is the ID Lago minted; after a state transition it is the new status. The exceptions are read-shaped writes whose body is the answer (`invoices preview`, `credit-notes estimate`, downloads, payment URLs) and bulk `events send`/`events batch`, which print in full. `--output json` returns the complete resource, and that is the form to script against:

```console
$ lago customers create --external-id quickstart_customer --name "Quickstart Customer" --output json
{
  "customer": {
    "lago_id": "1a901a90-1a90-1a90-1a90-1a901a901a90",
    "external_id": "quickstart_customer",
    "name": "Quickstart Customer",
    "currency": "USD",
    "created_at": "2026-09-01T10:00:00Z"
  }
}
```

Reads print in full. `lago customers get` prints every field as key/value rows, with nested lists and objects summarised (`CHARGES  2 items: requests, storage`) rather than dumped as JSON. `--output json` is the structured form.

### List columns

`list` commands print one row per item. Four resources have a fixed, documented column set; every other resource picks identifiers, status, money amounts with their currency, and dates first, across all rows on the page, up to eight columns.

```console
$ lago customers list
LAGO_ID  EXTERNAL_ID  NAME  EMAIL  CURRENCY  CREATED_AT
$ lago invoices list
LAGO_ID  NUMBER  STATUS  PAYMENT_STATUS  CURRENCY  TOTAL_AMOUNT_CENTS  ISSUING_DATE
$ lago subscriptions list
LAGO_ID  EXTERNAL_ID  PLAN_CODE  STATUS  STARTED_AT  ENDING_AT
$ lago plans list
LAGO_ID  CODE  NAME  INTERVAL  AMOUNT_CENTS  AMOUNT_CURRENCY
```

A column absent from every row on the page is dropped. When the page is one of several, stderr says so: `page 1 of 3 (250 total); use --page N or --all`. The pagination `meta` object stays in `--output json`.

Table cells escape terminal control characters. A name containing an ANSI sequence or a newline prints as `\x1b[31m...` and `\n`, so a hostile value cannot recolour your terminal or inject a fake row. JSON output escapes on its own.

`events send` generates a transaction ID when one is omitted. When you pass your own `--transaction-id`, pass `--timestamp` with it and resend both unchanged on retry. `--timestamp` takes Unix seconds (`1788273288`, decimals allowed) or an RFC 3339 instant (`2026-09-01T14:34:48Z`, converted to Unix seconds before sending); the same conversion applies to a `timestamp` string in `--input` and `--file` events. A millisecond epoch (`1788273288000`) is refused with a hint rather than stored 3000 years out, and so is a number too small to be Unix seconds. Lago deduplicates on `transaction_id`, but on the ClickHouse event store the timestamp is part of the key and a missing one defaults to the time of reception, so a retry without a timestamp is a second billable event. The CLI warns on stderr when a command is not safe to retry for that reason.

For bulk ingestion, stream newline-delimited JSON without loading the file into memory:

```console
$ lago events send --file events.ndjson --concurrency 8
$ cat events.ndjson | lago events send --file -
```

## Querying responses with `--query`

`--query` takes a [JMESPath](https://jmespath.org) expression. The trap is that **Lago wraps every response**, and the wrapper is part of the path. A query written against the unwrapped resource matches nothing and returns `null`. Table mode unwraps `{"customers": [...], "meta": {...}}` for display; `--query` and `--output json` see the envelope as sent.

| Endpoint | Response shape | Query starts with |
| --- | --- | --- |
| `customers get` | `{"customer": {...}}` | `customer.` |
| `customers list` | `{"customers": [...], "meta": {...}}` | `customers[` |
| `subscriptions get` | `{"subscription": {...}}` | `subscription.` |
| `invoices list` | `{"invoices": [...], "meta": {...}}` | `invoices[` |

Two worked examples. Reach through the wrapper, then select:

```console
$ lago customers get quickstart_customer --output json --query 'customer.lago_id'
"1a901a90-1a90-1a90-1a90-1a901a901a90"

$ lago invoices list --output json --query 'invoices[?status==`"finalized"`].{id: lago_id, total: total_amount_cents}'
[
  {
    "id": "4d904d90-4d90-4d90-4d90-4d904d904d90",
    "total": 150000
  }
]
```

Note the quoting in that filter. A JMESPath backtick literal holds **JSON**, so a string
comparison needs the inner double quotes: `` `"finalized"` `` works, bare `` `finalized` ``
is not valid JSON and is rejected as an invalid query. The alternative is the raw-string
form `'finalized'`, which means double-quoting the whole expression for your shell:

```console
$ lago invoices list --output json --query "invoices[?status=='finalized'].[lago_id]"
```

When a query matches nothing, the CLI says so on stderr and still prints `null` on stdout, so a script's parsing does not change:

```console
$ lago customers list --output json --query 'lago_id'
query matched nothing; top-level keys: customers, meta
null
```

`--query` implies `--output json` when you have not chosen a format, because a JMESPath result is structured data and an empty table is not an answer. The switch is announced, not silent, and an explicit `--output` always wins:

```console
$ lago customers list --query 'customers[].external_id'
--query implies --output json; pass --output table explicitly to render the result as a table.
[
  "quickstart_customer"
]
```

### When an identifier is the wrong kind

Lago answers an unknown identifier with a bare `404 Not Found`, which reads as "no data" rather than "no such thing". The CLI names the resource type and the value instead, and exits 4:

```console
$ lago events send --external-subscription-id ai_plan_gpt4_tokens --code api_calls
Error: no subscription "ai_plan_gpt4_tokens" exists in this organization
Suggestion: Check that each identifier is the right kind for the flag it was passed to: a plan code is not a subscription external ID, and a Lago ID is not an external ID.
$ echo $?
4
```

`--all` and `--query` are mutually exclusive: `--all` streams pages without buffering, so a whole-collection expression has nothing to evaluate against. Use a per-page `--query`, or stream JSON pages into `jq`.

### Validation errors name the field

A 422 from Lago carries the failing field and reason in `error_details`. Default output prints each one under the summary line, so the fix is visible without switching to `--output json`, which carries the same map under `error.details`.

```console
$ lago customers create --external-id acme --name Acme
Error: Unprocessable Entity
  external_id: value_already_exist
HTTP status: 422
Lago code: validation_errors
Request ID: 3f1c…
Suggestion: Check the command flags and Lago API validation details.
```

## Everyday commands

```console
$ lago whoami
$ lago doctor
$ lago customers list --all --output json
$ lago api GET /customers?page=2
$ lago api POST /events --data @event.json
```

Global scripting controls: `--output table|json|yaml`, `--query`, `--dry-run`, `--timing`, `--verbose`, `--timeout`, `--no-retry`. `--timing` separates API round-trip, retry wait, and CLI overhead, and prints on failure too, including network errors, so retry behaviour is visible when it matters. The API key is redacted from dry runs, errors, and verbose logs.

## Stable exit codes

| Code | Meaning |
| ---: | --- |
| 0 | Success |
| 1 | Generic error |
| 2 | Usage error |
| 3 | Authentication or authorization error |
| 4 | Not found |
| 5 | Validation or other API 4xx error |
| 6 | Rate limited |
| 7 | Lago server 5xx error |
| 8 | Network or timeout error |

Changing this table is a breaking change.

## Development

Go 1.27.0 or newer is required. `go.mod` pins the minimum, so `GOTOOLCHAIN=auto` (the default) fetches it for you. Older toolchains are rejected by the security gate.

```sh
make generate
make build
make test
make coverage
make lint
make security
```

Pull requests validate the checked-in OpenAPI snapshot. A daily trusted workflow fetches `https://swagger.getlago.com/openapi.yaml` and opens a spec-drift PR; `@getlago/developers` owns a one-business-day merge SLA. See [ARCHITECTURE.md](ARCHITECTURE.md), [CONTRIBUTING.md](CONTRIBUTING.md), [DECISIONS.md](DECISIONS.md), and [SECURITY.md](SECURITY.md).

## Telemetry

The CLI sends no product telemetry. With explicit update-check consent, it makes at most one anonymous release metadata request per 24 hours; disable it with `LAGO_NO_UPDATE_CHECK=1`. Lago uses public release downloads, Homebrew tap analytics, and public documentation traffic as aggregate adoption proxies.

## License

MIT
