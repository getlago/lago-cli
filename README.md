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

Pass the **base** URL. The CLI appends `/api/v1` itself. Pasting the full path works too: `https://api.getlago.com/api/v1` normalizes to the same host, and so do a trailing slash, a custom port, and a sub-path behind a proxy. Do not pass the **app** URL: `https://app.getlago.com` is the dashboard, not the API.

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

`lago whoami` reports the active profile, region, mode, API URL, and organization. `lago doctor` runs the configuration, permission, network, and authentication checks.

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

Plain HTTP and disabled TLS verification require the explicit `--insecure` flag and always print a warning. Use it for `http://localhost:3000` during development, not against a deployment that holds real money.

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
    --subscription-external-customer-id quickstart_customer \
    --subscription-external-id quickstart_subscription \
    --subscription-plan-code quickstart
LAGO_ID      3c903c90-3c90-3c90-3c90-3c903c903c90
EXTERNAL_ID  quickstart_subscription

$ lago events send --external-subscription-id quickstart_subscription --code quickstart_requests
CODE            quickstart_requests
LAGO_ID         5e905e90-5e90-5e90-5e90-5e905e905e90
TRANSACTION_ID  0f3c1d2e-4a5b-6c7d-8e9f-0a1b2c3d4e5f

$ lago invoices preview --input '{"customer":{"external_id":"quickstart_customer"},"subscriptions":{"external_ids":["quickstart_subscription"]}}'
CURRENCY            USD
FEES                [{"amount_cents":100,"units":"1.0"}]
ISSUING_DATE        2026-10-01
TOTAL_AMOUNT_CENTS  100
```

The `billable_metric_id` in the plan payload is the `LAGO_ID` the first command printed. Substitute yours.

### Creates print identifiers

Every `create` and `update` prints a terse identifier block by default, because after a create the one thing you do not already have is the ID Lago minted. `--output json` returns the complete resource, and that is the form to script against:

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

Reads are unchanged: `lago customers get` and `lago customers list` print every column.

`events send` generates a transaction ID when one is omitted. For bulk ingestion, stream newline-delimited JSON without loading the file into memory:

```console
$ lago events send --file events.ndjson --concurrency 8
$ cat events.ndjson | lago events send --file -
```

## Querying responses with `--query`

`--query` takes a [JMESPath](https://jmespath.org) expression. The trap is that **Lago wraps every response**, and the wrapper is part of the path. A query written against the unwrapped resource matches nothing and returns `null`.

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

A query that skips the wrapper matches nothing and yields `null`:

```console
$ lago customers list --output json --query 'lago_id'
null
```

`--all` and `--query` are mutually exclusive: `--all` streams pages without buffering, so a whole-collection expression has nothing to evaluate against. Use a per-page `--query`, or stream JSON pages into `jq`.

## Everyday commands

```console
$ lago whoami
$ lago doctor
$ lago customers list --all --output json
$ lago api GET /customers?page=2
$ lago api POST /events --data @event.json --idempotency-key event-42
```

Global scripting controls: `--output table|json|yaml`, `--query`, `--dry-run`, `--timing`, `--verbose`, `--timeout`, `--no-retry`. `--timing` separates API round-trip, retry wait, and CLI overhead. The API key is redacted from dry runs, errors, and verbose logs.

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

Go 1.25.13 or newer is required; older 1.25 patch releases contain reachable standard-library vulnerabilities and are rejected by the security gate.

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
