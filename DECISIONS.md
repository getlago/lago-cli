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
the `x-lago-cli-action` precedent.
A contract test fails if any destructive operation is gated or retried incorrectly.

*Superseded in part on 2026-09-02: `--idempotency-key` is no longer offered. See "No
client-side idempotency key" below.*

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

*Superseded in part on 2026-09-02: the exclusion of deletes and state transitions is
withdrawn now that `status` is an identifier key. See "Every write prints identifiers plus
status" below. The exclusions for read-shaped actions and bulk ingestion stand.*

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

## 2026-09-02 — Reject what the API would mishandle: userinfo URLs and out-of-range pages

**Userinfo is refused, not stripped.** `https://api.getlago.com@evil.example` is a valid
URL whose host is `evil.example`; the part before `@` is a username. QA S-16 and N-9 showed
the CLI printing one host and dialling another. The normalizer now refuses any URL with
userinfo, at exit 2, with `url.URL.Redacted()` so a pasted password never lands in a
terminal or a log. Stripping the userinfo and continuing would be guessing which half the
operator meant.

**`--limit` is bounded at 1..1000 client-side.** `--limit 0` reached lago-api as
`per_page=0` and came back as a 500: `BaseQuery#paginate` hands `per_page` straight to
Kaminari (`scope.page(page).per(limit)`) with no `max_per_page`, and a zero page size
divides by zero in `total_pages`. The spec declares no `maximum` on `per_page` either. So
the lower bound is a server bug shield until lago-api validates the parameter, and the
upper bound is a sanity limit, not a contract: nobody reads a thousand-row table, and
`--all` exists for the rest. `--page` gets the same check on the single-page path that
`--all` already had. Both are handed to lago-api and lago-openapi as fixes to make.
## 2026-09-02 — Required means required and non-nullable; types come from the spec

Three generator findings from QA, one derivation pass.

**A body is required only when the spec says so.** The generator never read
`requestBody.required`, and the runtime demanded a body whenever one was declared. So
`invoices void`, which the spec documents as callable without a body, refused the call.
`Body.Required` now mirrors the spec, whose default is false. Ten operations become
bodiless-capable; two of them, `credit-notes estimate` and `fees update`, merely omit
`required` upstream and the server still validates them. That is a lago-openapi fix, not
a CLI workaround, and it is handed off rather than patched here.

**A nullable field is never a required flag.** `SubscriptionUpdateInput.subscription`
lists `ending_at` under `required` and types it `[string, 'null']`. In JSON Schema that
means the key must be present and null satisfies it; the CLI read it as "a value must be
given" and blocked every `subscriptions update` that did not set an end date. Required is
now `required && !nullable`, `Field.Nullable` records the union, and a contract test walks
every body so the trap cannot return through another endpoint. It was the only
flag-producing case in the spec; 82 others are in response schemas, which never become
flags.

**Types resolve through unions.** OpenAPI 3.1 writes nullability as `type: [integer,
'null']`. The generator read `type` as a plain string, failed on the list, and fell
through to "string", so `coupons create --amount-cents 1000` sent `"1000"` while
`add-ons create` sent `1000`. Unions resolve to their non-null member; a `oneOf`/`anyOf`
resolves when its scalar members agree and stays a string otherwise (`event.timestamp`
is `oneOf [integer, string]`, and the server accepts both). 26 integer, 8 boolean and 2
decimal fields change shape. A contract test runs every write with synthetic values and
asserts the JSON kind on the wire matches the spec type: the request-body counterpart of
`internal/moneycheck`, which only inspects Go source.
## 2026-09-02 — Raw and scripted requests share the generated danger classification

QA deleted a live customer through a fixture step and through `lago api DELETE` without
being asked once. Generated commands were gated; the two paths that bypass the command
tree were not. Both now classify a request by looking up its method and path in the
generated operation table and reading the `Dangerous` flag the generator derived from
the spec. That keeps one classification: `POST /invoices/{id}/void` is gated because the
spec says so, not because a second verb list happened to agree. A path no operation
claims falls back to the generator's default-deny rule, DELETE or a destructive segment,
so an unknown endpoint is never silently ungated. The vocabulary moved to
`internal/generated` so the generator, the runtime, and the contract test read one list.

`fixtures run` and `seed demo` are refused outside test profiles rather than offered a
live-mode confirmation. A scenario creates and deletes many objects; confirming the
first destructive step says nothing about the rest, and a refusal midway leaves the
account half-applied. So the scan runs before step one and the whole fixture is
confirmed by name, or not run at all.

`lago api` keeps test mode ungated. It exists as the raw escape hatch for endpoints and
shapes the generated commands do not cover yet, and a test organization is the place to
use it that way. Live mode goes through `--confirm <path>` or the typed prompt, the same
gate as a generated delete. Recorded because the asymmetry is deliberate.
## 2026-09-02 — Profiles: explicit switching, announced insecurity, warned permissions

Four QA findings about what the CLI persists and how quietly it does so.

**`init --profile X` no longer switches the current profile.** It did, so configuring a
staging profile silently pointed every later command at staging. The first profile ever
written becomes current, because there is nothing else to point at; after that, switching
is opt-in with `--use`, and stderr names the profile that remains current. Recorded as
breaking pre-1.0.

**`--insecure` is persisted only for the init that passes it, and announced whenever it
is true.** An init with the flag silently disabled TLS verification for every later command
on the profile. It is now announced on stderr. A re-init without the flag clears it and
says so, rather than inheriting it from the stored profile: a setting that weakens
security is re-asked every time, in the fail-safe direction. The per-command `--insecure`
warning stays.

**Loose config permissions warn on every command.** `doctor` already checked the mode;
QA asked why only `doctor`, since the file holds API keys and the operator who never runs
`doctor` is the one who needs telling. The check runs once per invocation in `App.Load`.
It warns rather than refuses, unlike ssh: a refusal would break every script on a shared
CI volume for a condition fixed by the one command the warning prints. Windows has no
POSIX mode bits and is skipped.

**Aliases may not carry credentials or TLS flags.** An alias with `--api-key` is a second
place a key lives, outside the 0600 profile table; one with `--insecure` disables TLS
checks for whoever runs it without seeing the flag. `--api-url`, `--api-key` and
`--insecure` are refused at `alias set` with the profile alternative named. `--profile` and
`--mode` stay allowed: a mode is a safety declaration, not a secret.
## 2026-09-02 — No client-side idempotency key

QA sent two creates with the same `--idempotency-key` and different bodies. Both
succeeded. The header reached the server and nothing read it: `IdempotencyRecord` in
lago-api is internal to invoice generation and no controller consults the request header.
A flag whose help promised "safe mutation retries" was a correctness lie in a billing
tool, so it is removed rather than kept as a no-op: from generated commands, from
`lago api`, from `plans import`, and from the fixture schema, which now rejects
`idempotency_key` by name so an old fixture fails loudly instead of silently losing a
safety it never had.

The removal changes the retry policy, because the key was the only thing that made a
mutation `Idempotent` for the transport. Mutations are now never auto-retried; a 429 or
5xx on a write is reported and the operator decides. The one exception is a usage event
that carries a `timestamp`: Lago deduplicates events on `transaction_id`, and on the
ClickHouse store on `transaction_id` plus `timestamp`, so replaying such an event with the
same body is provably safe. An event without a timestamp is sent once, which is the rule
the 2026-09-02 events warning already taught. `plans import` loses its retry too; a PUT
that races a concurrent edit is not idempotent in the billing sense, whatever RFC 9110
says.

The removal is a single commit so it can be reverted the day lago-api implements the
header. Until then the request to lago-api is: read `Idempotency-Key` on every write, or
document that it never will, so no client offers the flag again.
## 2026-09-02 — Table cells are display-safe, list envelopes unwrap, columns are declared

Four QA findings, one renderer, four rules.

**Every cell is escaped, never stripped.** A customer name holding `\e[31m` and a newline
recoloured the terminal and injected a fake row. All C0 and C1 control characters and
invalid UTF-8 bytes now print as visible escapes (`\x1b`, `\n`, `\r`, `\t`, `\xNN`) on the one
path every value takes to the terminal, `output.Sanitize`, and the text error printer uses
the same function. Replacing keeps the evidence: an operator sees the name contains an
escape instead of a silently shortened string. JSON output is untouched; `encoding/json`
escapes on its own.

**The list envelope unwraps; `meta` goes to stderr.** `{"customers": [...], "meta": {...}}`
has two keys, so the single-key unwrap left it as two key/value rows with the page JSON
in one cell. Exactly one array beside an optional `meta` object is now recognised as a
list and rendered one row per item. `meta` is omitted from stdout: a pagination object as
a table row is noise for a reader and useless to a script, which reads `--output json`
where `meta` is intact. When `total_pages > 1` a one-line stderr hint names the page and
the two ways to get the rest, the same channel the empty-query hint already uses. `--all`
renders page by page without buffering, so its header repeats per page and the hint is
suppressed; buffering rows to print one header would contradict the streaming rule.

**Columns are declared per resource, in one file.** Customers, invoices, subscriptions and
plans carry a fixed column list in `internal/output/columns.go`, verified against the
pinned spec by a contract test so a rename fails the build rather than printing an empty
column. Every other resource takes the heuristic: identifiers, then status, then each
money amount followed by its currency, then dates, then the rest, drawn from the union of
keys across all rows on the page (the old renderer read the first row only) and capped at
eight. The map is data, not command code, so adding a resource is one line.

**A cell never contains JSON.** A nested list summarises as `2 items: requests, storage`
(labels from code, external ID, name or Lago ID, capped at five), a nested object as its
identifier pairs, as `k=v` pairs when it is four scalars or fewer, or as `{12 fields}`.
The structured form is one flag away. The exception is `--dry-run`: its envelope carries
the request payload under `body`, which the summary would reduce to `{2 fields}`, so the
envelope prints as JSON in table mode. It is a request, not a resource, and JSON is the
form it will be sent in.

## 2026-09-02 — Every write prints identifiers plus status; read-shaped writes print in full

QA X-3: `delete`, `terminate`, `apply` and the wallet transaction commands dumped the
full resource while `create` and `update` printed the terse block, so the same tool had
two default shapes for a write. The 2026-09-01 rule excluded state transitions because
"the interesting output is the new state", which is true and is exactly why `status`
now sits in the identifier block, after `name`. With that, every non-GET operation is
terse: 109 of 217 commands. The generator rule is still one function.

The exclusions that stand are the ones where reduction would delete the answer: a POST
that is a question (`invoices preview`, `credit-notes estimate`, `events estimate-fees`,
`billable-metrics evaluate-expression`), a download or payment URL, and bulk ingestion
whose output is a summary (`events send`, `events batch`). Twelve actions, listed in the
generator and independently in the contract test so the two cannot drift silently.

QA M-empty: key/value tables drop null and blank values. `events send` is asynchronous and
answers with most fields unset; five blank rows tell the operator less than the three
that carry a value. When every value is blank the keys still print, so the output is
never empty. A terse array response renders one row per item with the identifier and
status columns rather than a header-less block.

QA C-3: `whoami` printed the organization as a JSON blob in one cell. It is now a short
identity block in reading order, name first and the resolved host last, because "which
organization and which host" is the whole question. The JSON form drops the double
nesting (`organization` holds the object directly), recorded as breaking pre-1.0.

## 2026-09-02 — Flag and action names: primary key unwrapped, scoped nouns kept, clean break

Two naming rules in the generator, both breaking pre-1.0 and taken now because the
command surface is not yet frozen.

**The required object of a multi-key envelope is the primary key, and its fields are
not prefixed.** `SubscriptionCreateInput` is `{subscription: {...}, authorization: {...}}`,
so the single-property unwrap did not fire and every flag carried the envelope key:
`--subscription-external-id`, `--subscription-plan-code`. When a multi-key envelope has
exactly one required object property, that property is the resource being created and
its children keep their own names; the siblings keep their prefix
(`--authorization-amount-cents`, and `--status` on update, which is a scalar). Paths are
unchanged, only flag spelling is, and generation fails if two fields would share a flag.
Only the two subscription operations qualify in the current spec.

**A path-scoped operation keeps the noun of what it acts on.** `wallets create-customer`
created a wallet; `entitlements destroy-subscription` deleted an entitlement. Both came
from stripping the resource noun off the operation ID. For operations whose path is
scoped under another resource, the noun is kept whenever the stripped remainder still
carries a qualifier: `create-customer-wallet`, `destroy-subscription-entitlement`,
`list-applied-coupons`. A bare verb (`applyCoupon` under `/applied_coupons`) is unambiguous
and stays short. Eleven commands rename. Nesting under the parent
(`customers wallets create`) was rejected: it moves forty operations into a three-level
tree and breaks the one-operation-one-command parity the contract test guards.

**Clean break, no aliases.** The repository is internal and the output shapes are
already changing in this release. A hidden alias table would preserve the misleading
names in the code for the life of 1.x to spare a rename nobody has scripted yet.

## Deferred beyond 1.x

Plugin/extension system, TUI dashboard, and `lago scaffold` sample-app generation.
