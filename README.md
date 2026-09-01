# Lago CLI

The official command-line interface for [Lago](https://getlago.com), the open-source usage-based billing platform. The CLI is generated from Lago's OpenAPI specification and ships as one static Go binary.

> This repository is under active development. Until the first signed release, build from source and use a Lago test profile.

## Install

Until the first signed release, install from source:

```console
$ go install github.com/getlago/lago-cli/cmd/lago@latest
```

### Channels at GA

These channels are provisioned as part of the release sequence. Each becomes a
documented install command only after its endpoint is live and smoke-tested, so
none of them are printed above as if they already work.

| Platform | Command | Status |
| --- | --- | --- |
| macOS / Linux | `curl -fsSL https://getlago.com/install.sh \| sh` | endpoint not yet published |
| Homebrew tap | `brew install getlago/tap/lago` | tap not yet provisioned |
| Docker | `docker run --rm ghcr.io/getlago/lago-cli:latest version` | published on first tag |
| Windows | `winget install Lago.LagoCLI` or `scoop install lago` | repositories not yet provisioned |
| Go | `go install github.com/getlago/lago-cli/cmd/lago@latest` | available now |

Publishing order is not a checklist. `curl -fsSL … | sh` on a missing endpoint
fails silently: curl writes nothing, `sh` reads an empty stream, and the shell
exits 0 having installed nothing. So each channel is published, then verified,
then documented, in that order:

1. Publish `install.sh` at `https://getlago.com/install.sh` and confirm HTTP 200.
2. Provision the Homebrew tap and the Windows package repositories.
3. Tag the release; the pipeline signs artifacts and publishes checksums.
4. Run the post-release smoke job, which installs from the real endpoints.
5. Only then move a row above into the install table.

The canonical installer is served only from `getlago.com`, a Lago-controlled domain. It verifies the release checksum before installing. Release artifacts include checksums, signatures, provenance, and SBOMs; the installer verifies the checksum, and signature verification is tracked in [DECISIONS.md](DECISIONS.md).

## Five-minute billing flow

Initialize a test profile. `init` exits successfully only after a live `GET /api/v1/organizations` validates the URL and key.

```console
$ lago init --api-key "$LAGO_API_KEY" --region us --mode test
Connected to Lago as Example Organization.
Saved us profile "default" to ~/.config/lago/config.toml (mode: test).
```

Create a metric, plan, customer, and subscription; then send usage and preview the invoice. Every payload and enum below comes from the pinned OpenAPI document.

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

$ lago invoices preview --input '{"customer":{"external_id":"quickstart_customer"},"subscriptions":{"external_ids":["quickstart_subscription"]}}'
```

Creates and updates print identifiers, not the full resource: after a create the one
thing you do not already have is the ID Lago minted. The complete resource is one flag
away, and that is the form to script against.

```console
$ lago customers create --external-id quickstart_customer --name "Quickstart Customer" --output json
{
  "customer": {
    "lago_id": "1a901a90-1a90-1a90-1a90-1a901a901a90",
    "external_id": "quickstart_customer",
    "name": "Quickstart Customer",
    "currency": "USD",
    ...
  }
}
```

Reads are unchanged: `lago customers get` and `lago customers list` print every column.

`events send` creates a transaction ID when one is omitted. For bulk ingestion, stream newline-delimited JSON without loading the full file into memory:

```sh
lago events send --file events.ndjson --concurrency 8
cat events.ndjson | lago events send --file -
```

## Profiles and safety

Configuration precedence is flags, then environment, then `~/.config/lago/config.toml`. Select named profiles with `--profile staging`. Profiles declare `mode = "live"` or `mode = "test"`; credential overrides without an explicit mode deliberately resolve to live.

Live commands print `[LIVE]`. Destructive live operations require the resource identifier via `--confirm` or typed interactive confirmation. Plain HTTP and disabled TLS verification require the explicit `--insecure` flag and always print a warning.

```sh
lago whoami
lago doctor
lago customers list --all --output json
lago api GET /customers?page=2
lago api POST /events --data @event.json --idempotency-key event-42
```

Global scripting controls include `--output table|json|yaml`, `--query` for JMESPath, `--dry-run`, `--timing`, `--verbose`, `--timeout`, and `--no-retry`. `--timing` separates API round-trip, retry wait, and CLI overhead. The API key is redacted from dry runs, errors, and verbose logs.

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

Pull requests validate the checked-in OpenAPI snapshot. A daily trusted workflow fetches `https://swagger.getlago.com/openapi.yaml` and opens a spec-drift PR; `@getlago/developers` owns a one-business-day merge SLA. See [ARCHITECTURE.md](ARCHITECTURE.md), [CONTRIBUTING.md](CONTRIBUTING.md), and [SECURITY.md](SECURITY.md).

## Telemetry

The CLI sends no product telemetry. With explicit update-check consent, it makes at most one anonymous release metadata request per 24 hours; disable it with `LAGO_NO_UPDATE_CHECK=1`. Lago uses public release downloads, Homebrew tap analytics, and public documentation traffic as aggregate adoption proxies.

## License

MIT
