# Decisions

## 2026-08-31 — Go and generated API commands

Use Go, Cobra, and pflag for one cross-platform static binary. Generate every API command from the pinned OpenAPI document. `x-lago-cli-action` is the upstream escape hatch for naming collisions; no CLI-local operation mapping is allowed.

## 2026-08-31 — Backend-independent 1.0 gate

The 1.0 GA gate is install, init, full API parity, golden billing flow, safety, and all harnesses. Browser device login, webhook listen/trigger, activity-log tailing, and dashboard deep links require server contracts and ship through the 1.1 beta channel as those contracts land. Capability probes must return actionable unsupported-version errors.

## 2026-08-31 — Controlled distribution surfaces

The GA Homebrew promise is `brew install getlago/tap/lago`. Homebrew Core is a fast-follow because its third-party review queue is outside Lago's control. No artifact or documentation may use `get.lago.com`, which Lago does not control.

*Superseded in part on 2026-09-01: the shell installer this originally named was never published and is now parked. See "Two install channels for 1.0" below.*

## 2026-08-31 — Repository publication sequence

Before any INTERNAL-to-PUBLIC change: scan all Git history with gitleaks, scan the worktree and fixtures, enable secret scanning and push protection, activate branch/ruleset protection, and verify fork-PR secret isolation. Only then may an authorized maintainer change visibility. The repository remains INTERNAL until explicitly approved.

## 2026-08-31 — Spec drift without flaky PRs

PRs use the checked-in spec. A scheduled trusted workflow fetches live and opens a drift PR. `@getlago/developers` owns those PRs with a one-business-day SLA. No unresolved drift PR is the default release-candidate gate, accepting that upstream churn can intentionally pause a billing-tool release. A narrower CLI maintainers team can replace this existing team after it is created and granted repository write access.

## 2026-08-31 — Environment mode is a client safety declaration

The organization API does not expose an authoritative live/test field. Mode is therefore explicit profile state. Any flag or environment credential override without an explicit mode resolves to live, failing toward caution.

## 2026-08-31 — Exact money and no telemetry

Minor currency units use `int64`; fractional values use validated decimal strings. An AST check rejects floating-point monetary fields. The CLI sends no product telemetry. Aggregate release downloads, tap installs, and public docs traffic are the adoption proxies.

## 2026-08-31 — Device-auth key scope and auditability

Lago's current key model does not provide a CLI-specific scope, so 1.1 device auth may mint a regular organization key only after explicit admin approval. The endpoint must name it `Lago CLI — approved by {admin}, {date}` so administrators can audit and revoke it. Scoped keys are separate permission-model work.

## 2026-08-31 — Danger and retry safety come from the spec, not the HTTP verb

RFC method semantics do not describe billing semantics. `PUT /invoices/{id}/finalize`
is idempotent by RFC 9110 and irreversible by accounting; `POST /invoices/{id}/retry_payment`
charges a card. Classification is therefore derived from the spec path and defaults to
deny: an operation matching the destructive vocabulary (`void`, `finalize`, `retry`,
`terminate`, `delete`, `destroy`, `refund`) or using DELETE is confirmation-gated, and
only GET, HEAD, and OPTIONS are auto-retried without an operator-supplied idempotency key.
`x-lago-cli-dangerous` and `x-lago-cli-retryable` are the upstream escape hatches, matching
the `x-lago-cli-action` precedent. `--idempotency-key` is offered on every mutation.
A contract test fails if any destructive operation is gated or retried incorrectly.

## 2026-08-31 — Per-package coverage floors, never a repo-wide average

A single average lets a well-covered package mask an untested one. The gate this
replaced measured only `./pkg/...`, which is 111 lines, and reported 90% while the
packages holding money, credentials, and the live-mode guardrails went unmeasured.
`coverage.floors` carries one floor per package, enforced on every PR; floors ratchet
upward only and a package with no floor entry fails the build. The packages that touch
money, credentials, or safety gates meet the 85% target: cli 86%, config 85%, transport
93%, output 96%, apperr 95%, redact and permissions 100%, pkg/api 90%. The generators
and doc tooling carry ratchet floors instead, because a defect there fails
`make generate-check` visibly rather than moving money.

## 2026-08-31 — Every harness carries a proof that it fails

A linter, scanner, or golden file that has never been seen to fail is not a gate.
`internal/moneycheck/canary` holds one deliberate monetary-float defect per class the
checker must catch, and `TestCanaryIsRejected` fails if the checker accepts it.
`test/billing` runs offline on every PR against an httptest server, so billing accuracy
no longer depends solely on `test/e2e`, which is build-tagged and skips without trusted
staging credentials. `scripts/check-fixtures.sh` treats a malformed pattern as a failure
rather than reporting a clean tree.

## 2026-09-01 — Default mutation output is identifiers; full detail is `--output json`

QA returned that a create printing 40 attributes buries the one thing the caller does
not already have: the identifier Lago minted. Default table output for every generated
create and update is therefore a terse block of `lago_id`, `external_id`, `code` and
`name`, in that order, modelled on `gh`'s terse-success pattern. `--output json` and
`--output yaml` are unchanged and always carry the complete resource, so scripts read
the structured form and humans read the identifiers.

The classification is a generator rule, not 48 hand edits: a POST, PUT or PATCH whose
action is `create`/`update` or begins with `create-`/`update-`. It deliberately excludes
read-shaped mutations (`invoices preview`, `credit-notes estimate`), state transitions
whose interesting output is the new state (`invoices finalize`, `invoices void`,
`orders execute`), and bulk ingestion whose output is a summary (`events send`). Three
behaviours keep the reduction safe: a response with no recognisable identifier falls
back to the full table rather than printing nothing, an explicit `--query` is honoured
as written rather than reduced first, and `--dry-run` always prints the request envelope
in full.

Decided pre-1.0, while output shapes are not yet frozen. After 1.0 this is a breaking
change requiring a major version, which is exactly why it is being made now.

## 2026-09-01 — Two install channels for 1.0

QA returned that the CLI should ship through Homebrew and `go install` only. Both are
now documented, smoke-tested on every release, and the only channels that exist. The
shell installer, the PowerShell installer, the GHCR image, Scoop and Winget are removed
from the README, `.goreleaser.yml`, and the release workflow.

The installer scripts and the release Dockerfile are parked under `dist-channels/parked/`
rather than deleted, with a README recording why each was parked and the four conditions
that bring one back: demonstrated demand, a live Lago-controlled endpoint, a post-release
smoke job that fails the release, and documentation only after that job has passed once.
Publish, verify, document, in that order. Scoop and Winget had no files of their own and
are recovered from the commit that removed their `.goreleaser.yml` blocks.

The CI jobs went with them rather than being skipped: a skipped job is a muted test that
reads green. What did **not** change is platform support. The release still builds darwin,
linux and windows on amd64 and arm64, CI still compiles and smoke-tests that matrix, and
`go install` works anywhere Go runs. Re-adding a channel is a publish-and-docs change,
not a port.

`lago upgrade` no longer replaces the running binary. With no script channel there is no
install the CLI itself owns: Homebrew owns its Cellar and `go install` rebuilds from
source, so replacing a Homebrew-managed binary in place would leave brew reporting a
version it no longer has. `upgrade` now checks for a newer release and prints the command
matching how the binary was installed, or both commands when it cannot tell. The download,
checksum-verify and atomic-replace path was removed with the channel that needed it.

The guardrail is a docs test: `test/docs` fails if any documented surface names a parked
channel, or if the README stops documenting either supported one.

## 2026-09-01 — Go 1.27 and the gates the bump moved

`go.mod` declares `go 1.27.0`. Every workflow reads `go-version-file: go.mod`, so CI
followed from one line; the Dockerfile and README are the only places a Go version is
written as a literal, and a test asserts all three agree.

Three things the bump moved, recorded because none of them is obvious from the diff:

**gosec had to move with it.** gosec 2.22 cannot read Go 1.27 export data and fails with
`internal error: package "flag" without types was imported`. The pin is now 2.29.0.
Verifying the tool version against the tool version, not against the toolchain, showed
all eight new findings come from rules added in 2.23-2.29, not from Go 1.27: 2.22.10 on
the 1.27 tree and 2.29.0 on the 1.25 tree produce the same eight. G703 (path traversal by
taint) and G122 (WalkDir symlink TOCTOU) fire only in the maintainer-run generators, where
the taint source is the flag the maintainer typed; each site carries that reason.

**One finding was a real hardening, so it was fixed rather than annotated.** G204 flagged
`openBrowser` handing a URL to `open`/`xdg-open`/`rundll32`. The URL comes from the pinned
spec, but those are general-purpose handlers that act on `file://` and custom schemes, so a
spec-drift PR changing one URL would have become code execution on the reader's machine.
`lago docs` now refuses anything that is not an absolute https URL, and a test asserts
every documentation URL in the manifest satisfies that.

**`golang.org/x/sys` went to v0.47.0** to clear GO-2026-5024. govulncheck now reports zero
vulnerabilities in required modules, not just zero reachable ones.

**Coverage floors were re-measured, not lowered.** Go 1.27 instruments more statements per
function. With an unchanged suite, `internal/gen` reads 9.7% where it read 11.7%, and
`internal/moneycheck` 15.5% where it read 16.0%, while config, diagnostics, docgen and
transport rose. Both directions are recorded in `coverage.floors` with the date and the
old value, so a future reader can tell a re-measurement from a regression. The
ratchet-upward rule stands for everything else.

Re-verified after the bump: full race suite, govulncheck, gosec, gitleaks, actionlint,
the license audit, byte-identical repeated builds, and the cold `--help` budget
(p50 42ms, p90 57ms against the 100ms budget).

## 2026-09-01 — One URL normalizer, and errors that name what is wrong

Four findings from live QA, each fixed once and tested on cloud US, cloud EU, and
self-hosted alike. Nothing here special-cases a hostname.

**One normalizer, and it is idempotent.** `transport.NormalizeBaseURL` is the only place
a base URL is decided. It strips any `/api/v1` the operator pasted and appends exactly
one, so the base form and the pasted form reach byte-identical URLs on every deployment,
including custom ports and sub-paths behind a proxy. Stripping runs in a loop because a
doubled prefix was already written into config files. `init` now saves the normalized
value, which is why `whoami` reported one host while requests went to another. A second
normalizer covers the request path: `lago api GET /api/v1/customers`, the form anyone
copying from the API reference will type, no longer becomes `/api/v1/api/v1/customers`.
`--region us|eu` resolves through the same function, and `USAPI`/`EUAPI` lost their
`/api/v1` so the shorthand and the explicit base URL are provably one code path.

**A region and an explicit URL that disagree is an error, not a precedence question.**
`--region us --api-url https://api.eu.getlago.com` used to prefer one silently. That is
how someone writes to the wrong continent. They are accepted together only when they
normalize to the same URL, which is merely redundant.

**The app URL is refused by name.** `app.getlago.com` answers, so sending an API key
there produced a confusing parse error rather than "wrong host". The rule is the `app.`
subdomain rather than a list of two cloud hosts, because self-hosted deployments follow
the same split; a single-domain deployment is unaffected. It fails before any request,
so the credential never leaves the machine.

**`whoami` and `doctor` both report the resolved URL.** `api_url` is what the profile
holds, `resolved_api_url` is what the client calls. Two lines instead of one, because the
difference is the bug.

**A query that matches nothing says so, on stderr.** Lago wraps every response, so
`--query lago_id` against `{"customers": [...]}` is valid and matches nothing. `null`
still goes to stdout unchanged; the hint and the available top-level keys go to stderr.
A script's parsing does not change because a human was told something.

**`--query` without `--output` switches to JSON, and announces it.** A JMESPath result is
structured data and the table renderer has nothing useful to do with a projection: QA's
query rendered an empty table that read as "no results". The alternative, exiting 2 and
demanding `--output json`, costs a round trip to teach a rule the tool can simply follow.
An explicit `--output` always wins, `--output table` included, so nothing is taken away,
and the switch is printed rather than silent because a changed output format is something
a script author needs to know about.

**A 404 names the resource type and the value.** Lago sends a bare `Not Found`. The CLI
carries the identifiers a request addressed into the error, so a plan code passed as
`--external-subscription-id` reports `no subscription "ai_plan_..." exists in this
organization` at exit 4 instead of something that reads as "no usage". Only
identifier-shaped fields qualify (`id`, `code`, `*_id`, `*_code`), so a create's `name`
never appears; other statuses keep the API's own message, because a 422's validation
detail beats a list of identifiers.

`transport.Config.DialContext` exists so these tests run against the production
hostnames served locally, rather than proving URL handling only for `127.0.0.1`. CI runs
the affected packages once per deployment target.

## Deferred beyond 1.x

Plugin/extension system, TUI dashboard, and `lago scaffold` sample-app generation.
