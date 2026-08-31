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

## Deferred beyond 1.x

Plugin/extension system, TUI dashboard, and `lago scaffold` sample-app generation.
