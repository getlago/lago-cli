package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
)

func writeFixture(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// A fixture chains requests, feeding captured values into later steps. This is the
// mechanism behind `lago seed demo`, so a capture regression breaks the golden path.
func TestFixtureCapturesFlowIntoLaterSteps(t *testing.T) {
	var mutex sync.Mutex
	var bodies []string
	var paths []string

	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		body := make([]byte, request.ContentLength)
		if request.ContentLength > 0 {
			_, _ = request.Body.Read(body)
		}
		mutex.Lock()
		paths = append(paths, request.URL.Path)
		bodies = append(bodies, string(body))
		mutex.Unlock()

		switch {
		case strings.HasSuffix(request.URL.Path, "/billable_metrics"):
			_, _ = response.Write([]byte(`{"billable_metric":{"lago_id":"bm_generated"}}`))
		default:
			_, _ = response.Write([]byte(`{"plan":{"lago_id":"plan_1"}}`))
		}
	})
	profileAt(t, server.URL)

	path := writeFixture(t, `
version: 1
name: capture-chain
vars:
  prefix: demo
steps:
  - id: metric
    method: POST
    path: /billable_metrics
    idempotency_key: ${prefix}-metric
    body:
      billable_metric:
        code: ${prefix}-requests
    capture:
      metric_id: billable_metric.lago_id
  - id: plan
    method: POST
    path: /plans
    body:
      plan:
        code: ${prefix}-plan
        charges:
          - billable_metric_id: ${metric_id}
`)

	stdout, _, err := execute(t, "", "--output", "json", "fixtures", "run", path)
	if err != nil {
		t.Fatalf("fixture run failed: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if len(paths) != 2 {
		t.Fatalf("fixture issued %d requests, want 2: %v", len(paths), paths)
	}
	if !strings.Contains(bodies[1], "bm_generated") {
		t.Errorf("captured metric ID did not reach the second step: %s", bodies[1])
	}
	if !strings.Contains(bodies[0], "demo-requests") {
		t.Errorf("variable interpolation did not reach the first step: %s", bodies[0])
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("fixture output is not JSON: %q", stdout)
	}
	if result["fixture"] != "capture-chain" {
		t.Errorf("fixture name lost: %v", result["fixture"])
	}
	variables, _ := result["variables"].(map[string]any)
	if variables["metric_id"] != "bm_generated" {
		t.Errorf("captured variable not reported: %v", variables)
	}
}

// --var overrides a declared variable, which is how the same fixture is reused
// against different accounts.
func TestFixtureVariableOverrides(t *testing.T) {
	var mutex sync.Mutex
	var body string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		raw := make([]byte, request.ContentLength)
		if request.ContentLength > 0 {
			_, _ = request.Body.Read(raw)
		}
		mutex.Lock()
		body = string(raw)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1"}}`))
	})
	profileAt(t, server.URL)

	path := writeFixture(t, `
version: 1
name: override
vars:
  code: default-code
steps:
  - id: customer
    method: POST
    path: /customers
    body:
      customer:
        external_id: ${code}
`)
	if _, _, err := execute(t, "", "--output", "json", "fixtures", "run", path, "--var", "code=overridden"); err != nil {
		t.Fatalf("fixture run failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if !strings.Contains(body, "overridden") {
		t.Errorf("--var did not override the declared variable: %s", body)
	}
}

// Every malformed fixture must be a usage error naming what is wrong, not a panic
// or a partially-executed scenario against a real account.
func TestFixtureValidationRejectsBadScenarios(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"ok":true}`))
	})
	profileAt(t, server.URL)

	for _, testCase := range []struct{ name, contents string }{
		{"wrong version", "version: 2\nname: x\nsteps:\n  - id: a\n    path: /customers\n"},
		{"no steps", "version: 1\nname: x\nsteps: []\n"},
		{"missing id", "version: 1\nname: x\nsteps:\n  - path: /customers\n"},
		{"duplicate id", "version: 1\nname: x\nsteps:\n  - id: a\n    path: /a\n  - id: a\n    path: /b\n"},
		{"unknown field", "version: 1\nname: x\nsteps:\n  - id: a\n    path: /a\n    nonsense: true\n"},
		{"undefined variable", "version: 1\nname: x\nsteps:\n  - id: a\n    path: /customers/${missing}\n"},
		{"not yaml", "::::not yaml::::\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := writeFixture(t, testCase.contents)
			if _, _, err := execute(t, "", "fixtures", "run", path); err == nil {
				t.Fatalf("%s was accepted", testCase.name)
			}
		})
	}

	if _, _, err := execute(t, "", "fixtures", "run", filepath.Join(t.TempDir(), "absent.yaml")); err == nil {
		t.Error("a missing fixture file was accepted")
	}
}

func TestFixtureVariableFlagParsing(t *testing.T) {
	t.Parallel()
	parsed, err := parseFixtureVariables([]string{"a=1", "b=with=equals", "c="})
	if err != nil {
		t.Fatal(err)
	}
	if parsed["a"] != "1" || parsed["b"] != "with=equals" || parsed["c"] != "" {
		t.Fatalf("parsed variables = %+v", parsed)
	}
	for _, invalid := range []string{"novalue", "=novalue", "   =x"} {
		if _, err := parseFixtureVariables([]string{invalid}); err == nil {
			t.Errorf("--var %q was accepted", invalid)
		}
	}
}

// Interpolation must substitute whole-value variables with their original type so
// a captured integer stays an integer in the outgoing JSON body.
func TestInterpolationPreservesNonStringTypes(t *testing.T) {
	t.Parallel()
	variables := map[string]any{"count": json.Number("42"), "name": "example", "flag": true}

	whole, err := interpolateFixtureValue("${count}", variables)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := whole.(json.Number); !ok {
		t.Errorf("a whole-value substitution changed type: %T", whole)
	}

	embedded, err := interpolateFixtureValue("id-${count}-${name}", variables)
	if err != nil {
		t.Fatal(err)
	}
	if embedded != "id-42-example" {
		t.Errorf("embedded interpolation = %v", embedded)
	}

	nested, err := interpolateFixtureValue(map[string]any{
		"list": []any{"${name}", map[string]any{"deep": "${flag}"}},
		"kept": json.Number("7"),
	}, variables)
	if err != nil {
		t.Fatal(err)
	}
	tree := nested.(map[string]any)
	list := tree["list"].([]any)
	if list[0] != "example" {
		t.Errorf("list interpolation = %v", list[0])
	}
	// A whole-value substitution keeps the variable's own type, so a captured
	// boolean stays a JSON boolean rather than becoming the string "true".
	if deep := list[1].(map[string]any)["deep"]; deep != true {
		t.Errorf("nested map interpolation = %#v, want the bool true", deep)
	}
	if kept, ok := tree["kept"].(json.Number); !ok || kept.String() != "7" {
		t.Errorf("a non-string leaf was altered: %#v", tree["kept"])
	}

	if _, err := interpolateFixtureValue("${absent}", variables); err == nil {
		t.Error("an undefined whole-value variable was accepted")
	}
	if _, err := interpolateFixtureValue([]any{"${absent}"}, variables); err == nil {
		t.Error("an undefined variable inside a list was accepted")
	}
	if _, err := interpolateFixtureValue(map[string]any{"k": "${absent}"}, variables); err == nil {
		t.Error("an undefined variable inside a map was accepted")
	}
}

// A failing step must abort the scenario and name the step, so an operator knows
// exactly how far a partially-applied fixture got.
func TestFixtureStopsAtTheFailingStepAndNamesIt(t *testing.T) {
	var mutex sync.Mutex
	requests := 0
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		requests++
		current := requests
		mutex.Unlock()
		if current == 2 {
			response.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = response.Write([]byte(`{"status":422,"error":"Unprocessable","code":"value_is_invalid"}`))
			return
		}
		_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1"}}`))
	})
	profileAt(t, server.URL)

	path := writeFixture(t, `
version: 1
name: partial
steps:
  - id: first
    method: POST
    path: /customers
  - id: second
    method: POST
    path: /plans
  - id: third
    method: POST
    path: /subscriptions
`)
	_, _, err := execute(t, "", "fixtures", "run", path)
	if err == nil {
		t.Fatal("a failing step did not abort the fixture")
	}
	if !strings.Contains(err.Error(), `"second"`) {
		t.Errorf("error does not name the failing step: %v", err)
	}
	if apperr.ExitCode(err) != apperr.ExitValidation {
		t.Errorf("exit code = %d, want the API's %d", apperr.ExitCode(err), apperr.ExitValidation)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if requests != 2 {
		t.Errorf("fixture issued %d requests; it must stop at the failure", requests)
	}
}

// A capture expression that matches nothing is a scenario bug: continuing would
// send a literal "<nil>" into a later billing request.
func TestFixtureFailsWhenACaptureMatchesNothing(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1"}}`))
	})
	profileAt(t, server.URL)

	path := writeFixture(t, `
version: 1
name: bad-capture
steps:
  - id: customer
    method: POST
    path: /customers
    capture:
      missing: nothing.here
`)
	if _, _, err := execute(t, "", "fixtures", "run", path); err == nil {
		t.Fatal("a capture that matched nothing was accepted")
	}
}

// --dry-run must issue no requests at all, and must still produce a usable plan so
// an operator can review a scenario before running it against a live account.
func TestFixtureDryRunSendsNothing(t *testing.T) {
	var mutex sync.Mutex
	reached := false
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		reached = true
		mutex.Unlock()
		_, _ = response.Write([]byte(`{}`))
	})
	profileAt(t, server.URL)

	path := writeFixture(t, `
version: 1
name: dry
steps:
  - id: metric
    method: POST
    path: /billable_metrics
    body:
      billable_metric:
        code: demo
    capture:
      metric_id: billable_metric.lago_id
  - id: plan
    method: POST
    path: /plans
    body:
      plan:
        metric: ${metric_id}
`)
	stdout, _, err := execute(t, "", "--dry-run", "--output", "json", "fixtures", "run", path)
	if err != nil {
		t.Fatalf("dry-run fixture failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if reached {
		t.Fatal("--dry-run issued a real request")
	}
	if !strings.Contains(stdout, "dry-run") {
		t.Errorf("dry-run output does not mark captures as unresolved: %s", stdout)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("dry-run output is not JSON: %q", stdout)
	}
	if result["dry_run"] != true {
		t.Errorf("dry_run flag missing from output: %v", result["dry_run"])
	}
}

// `lago seed demo` writes real data, so it must refuse to run against a live
// profile even when the operator asks.
func TestSeedDemoRefusesLiveProfiles(t *testing.T) {
	var mutex sync.Mutex
	reached := false
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		reached = true
		mutex.Unlock()
		_, _ = response.Write([]byte(`{}`))
	})

	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	file := config.NewFile()
	file.CurrentProfile = "default"
	file.Profiles["default"] = config.Profile{
		Region: "self-hosted", APIKey: "lago_live_FAKE000000000000000000000000",
		APIURL: server.URL, Mode: config.ModeLive, Insecure: true,
	}
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAGO_CONFIG_FILE", path)

	_, _, err := execute(t, "", "seed", "demo")
	if err == nil {
		t.Fatal("seed demo ran against a live profile")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	if !strings.Contains(err.Error(), "test profiles") {
		t.Errorf("refusal does not explain the restriction: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if reached {
		t.Fatal("seed demo reached the API before refusing")
	}
}

// The bundled demo fixture is the golden path. It must parse and run end to end.
func TestSeedDemoRunsTheBundledFixture(t *testing.T) {
	var mutex sync.Mutex
	requests := 0
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		requests++
		mutex.Unlock()
		switch {
		case strings.Contains(request.URL.Path, "billable_metrics"):
			_, _ = response.Write([]byte(`{"billable_metric":{"lago_id":"bm_1","code":"c"}}`))
		case strings.Contains(request.URL.Path, "plans"):
			_, _ = response.Write([]byte(`{"plan":{"lago_id":"plan_1","code":"p"}}`))
		case strings.Contains(request.URL.Path, "customers"):
			_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1","external_id":"e"}}`))
		case strings.Contains(request.URL.Path, "subscriptions"):
			_, _ = response.Write([]byte(`{"subscription":{"lago_id":"sub_1","external_id":"s"}}`))
		case strings.Contains(request.URL.Path, "events"):
			_, _ = response.Write([]byte(`{"event":{"lago_id":"evt_1","transaction_id":"t"}}`))
		default:
			_, _ = response.Write([]byte(`{"invoice":{"lago_id":"inv_1","total_amount_cents":100,"currency":"USD","fees":[{"amount_cents":100}]}}`))
		}
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "--output", "json", "seed", "demo", "--prefix", "unit-test")
	if err != nil {
		t.Fatalf("seed demo failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if requests < 5 {
		t.Errorf("seed demo issued only %d requests", requests)
	}
	if !strings.Contains(stdout, "invoice-preview") {
		t.Errorf("seed demo output has no invoice preview step: %s", stdout)
	}
	if !strings.Contains(stdout, "unit-test") {
		t.Errorf("--prefix did not reach the fixture: %s", stdout)
	}
}

func TestFixtureStepErrorPreservesTheCauseExitCode(t *testing.T) {
	t.Parallel()
	cause := apperr.New(apperr.ExitRateLimit, "rate limited", "wait")
	wrapped := fixtureStepError("charge", cause)
	if apperr.ExitCode(wrapped) != apperr.ExitRateLimit {
		t.Fatalf("exit code = %d", apperr.ExitCode(wrapped))
	}
	if !strings.Contains(wrapped.Error(), "charge") {
		t.Fatalf("message does not name the step: %v", wrapped)
	}
	if !strings.Contains(fmt.Sprint(wrapped), "rate limited") {
		t.Fatalf("message lost the cause: %v", wrapped)
	}
}

// QA F-5: a fixture with a DELETE step ran against a live profile without any gate.
// `fixtures run` now refuses non-test profiles the same way `seed demo` does, before
// the file is even read.
func TestQA_F5_FixturesRunRefusesLiveProfiles(t *testing.T) {
	var mutex sync.Mutex
	reached := false
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		reached = true
		mutex.Unlock()
		_, _ = response.Write([]byte(`{}`))
	})
	setCleanEnvironment(t)
	writeProfile(t, config.ModeLive, server.URL)
	path := writeFixture(t, destructiveFixture)

	_, _, err := execute(t, "", "fixtures", "run", path, "--confirm", "cleanup")
	if err == nil {
		t.Fatal("fixtures run executed against a live profile")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	if !strings.Contains(err.Error(), "test profiles") {
		t.Errorf("refusal does not explain the restriction: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if reached {
		t.Fatal("fixtures run reached the API before refusing")
	}
}

const destructiveFixture = `
version: 1
name: cleanup
steps:
  - id: create
    method: POST
    path: /customers
    body:
      customer:
        external_id: doomed
    capture:
      customer_id: customer.external_id
  - id: remove
    method: DELETE
    path: /customers/${customer_id}
`

// QA S-21: a destructive step must be confirmed before step one runs, so a refusal
// never leaves the scenario half-applied. Without a TTY and without --confirm the run
// is refused at exit 2, exactly like a generated delete.
func TestQA_S21_DestructiveFixtureStepRequiresConfirmationBeforeStepOne(t *testing.T) {
	var mutex sync.Mutex
	var paths []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		paths = append(paths, request.Method+" "+request.URL.Path)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customer":{"external_id":"doomed"}}`))
	})
	profileAt(t, server.URL)
	path := writeFixture(t, destructiveFixture)

	_, stderr, err := execute(t, "", "fixtures", "run", path)
	if err == nil {
		t.Fatal("destructive fixture ran without confirmation")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	if !strings.Contains(err.Error(), `confirmation required for cleanup`) {
		t.Errorf("refusal does not name the fixture: %v", err)
	}
	if !strings.Contains(stderr, "remove (DELETE /customers/${customer_id})") {
		t.Errorf("stderr does not list the destructive step:\n%s", stderr)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if len(paths) != 0 {
		t.Fatalf("steps ran before the gate: %v", paths)
	}
}

// --confirm with the fixture name satisfies the gate and every step runs in order.
func TestQA_S21_DestructiveFixtureRunsWithConfirmName(t *testing.T) {
	var mutex sync.Mutex
	var paths []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		paths = append(paths, request.Method+" "+request.URL.Path)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customer":{"external_id":"doomed"}}`))
	})
	profileAt(t, server.URL)
	path := writeFixture(t, destructiveFixture)

	if _, _, err := execute(t, "", "--output", "json", "fixtures", "run", path, "--confirm", "cleanup"); err != nil {
		t.Fatalf("confirmed fixture failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	want := []string{"POST /api/v1/customers", "DELETE /api/v1/customers/doomed"}
	if fmt.Sprint(paths) != fmt.Sprint(want) {
		t.Errorf("requests = %v, want %v", paths, want)
	}
}

// A dry run sends nothing, so it needs no confirmation, matching generated commands.
func TestQA_S21_DestructiveFixtureDryRunNeedsNoConfirmation(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) {
		t.Error("dry run reached the API")
	})
	profileAt(t, server.URL)
	path := writeFixture(t, destructiveFixture)

	if _, _, err := execute(t, "", "--dry-run", "--output", "json", "fixtures", "run", path); err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
}
