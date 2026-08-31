# Decisions

## 2026-08-31 — Go and generated API commands

Use Go, Cobra, and pflag for one cross-platform static binary. Generate every API command from the pinned OpenAPI document. `x-lago-cli-action` is the upstream escape hatch for naming collisions; no CLI-local operation mapping is allowed.

## 2026-08-31 — Backend-independent 1.0 gate

The 1.0 GA gate is install, init, full API parity, golden billing flow, safety, and all harnesses. Browser device login, webhook listen/trigger, activity-log tailing, and dashboard deep links require server contracts and ship through the 1.1 beta channel as those contracts land. Capability probes must return actionable unsupported-version errors.

## 2026-08-31 — Controlled distribution surfaces

The GA Homebrew promise is `brew install getlago/tap/lago`. Homebrew Core is a fast-follow because its third-party review queue is outside Lago's control. The canonical shell installer is `https://getlago.com/install.sh`; no artifact or documentation may use `get.lago.com`, which Lago does not control.

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

## Deferred beyond 1.x

Plugin/extension system, TUI dashboard, and `lago scaffold` sample-app generation.
