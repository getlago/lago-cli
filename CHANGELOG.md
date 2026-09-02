# Changelog

All notable changes are generated from conventional commits at release time. This project follows [Semantic Versioning](https://semver.org/).

## Unreleased

- Bootstrap the generated Lago CLI, secure profiles, resilient transport, raw API access, output formats, and golden billing commands.
- Pasting the full API path, a trailing slash, a custom port, or a sub-path now all normalize to one base URL, on cloud US, cloud EU, and self-hosted. `lago api GET /api/v1/customers` no longer requests `/api/v1/api/v1/customers`.
- Pasting a Lago **app** URL is refused before any request, naming the `api.*` host to use instead.
- `lago init` saves the normalized URL; `lago whoami` and `lago doctor` report the resolved host requests actually go to.
- `--region` and an explicit `--api-url` that name different deployments is now a usage error rather than a silent override.
- A `--query` that matches nothing prints a stderr hint naming the response's top-level keys, keeping `null` on stdout.
- `--query` without an explicit `--output` switches to JSON and says so. An explicit `--output` always wins.
- A 404 from a wrong identifier now names the resource type and the value, at exit 4, instead of a bare "Not Found".
- Build on Go 1.27.0. gosec moves to 2.29.0 (2.22 cannot read 1.27 export data), govulncheck to 1.7.0, and `golang.org/x/sys` to v0.47.0, clearing GO-2026-5024.
- `lago docs` now refuses to hand anything but an absolute https URL to the platform browser opener.
- **Breaking (pre-1.0):** install channels are Homebrew (`brew install getlago/tap/lago`) and `go install` only. The shell and PowerShell installers, the GHCR image, Scoop, and Winget are parked in `dist-channels/parked/`. Platform support is unchanged: all six os/arch targets are still built and smoke-tested.
- **Breaking (pre-1.0):** `lago upgrade` prints the upgrade command for how the binary was installed instead of replacing it in place.
- `lago upgrade` on a development build (`dev`, a local `VERSION=` override, a commit hash) explains that the binary is not a release and prints the rebuild command, without contacting GitHub. A failing release-metadata request (404 on a private repository, 403 from a proxy or rate limit, 5xx) now exits 8 (network) and names GitHub, instead of exit 7, which the exit-code table reserves for Lago server errors.
- `events send` warns on stderr when a caller-chosen `--transaction-id` (or a `--input`/`--file` event with a `transaction_id`) has no `timestamp`: on the ClickHouse event store the timestamp is part of the deduplication key and defaults to the time of reception, so such an event is not safe to resend. Streams report one warning with a count.
- A Lago API URL with credentials before the host (`https://api.getlago.com@evil.example`) is refused at exit 2 wherever a URL is accepted (`init`, profiles, `--api-url`, `LAGO_API_URL`); the printed host is always the dialled host. The password is never echoed.
- `--limit` must be an integer between 1 and 1000 and `--page` a positive integer, checked client-side at exit 2 before any request. `--limit 0` used to reach the API as `per_page=0` and come back as a 500.
- **Breaking (pre-1.0):** `create` and `update` commands now print a terse identifier block (`lago_id`, `external_id`, `code`, `name`) in default table output. `--output json` and `--output yaml` are unchanged and still return the complete resource.
