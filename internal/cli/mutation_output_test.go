package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/generated"
)

// fullCustomer, fullPlan and fullSubscription are the complete resources Lago returns
// from a create. They are deliberately verbose: the whole point of the terse default is
// that an operator does not have to read all of this to learn the ID they just minted.
const (
	fullCustomer = `{"customer":{"lago_id":"1a901a90-1a90-1a90-1a90-1a901a901a90",` +
		`"external_id":"quickstart_customer","name":"Quickstart Customer","sequential_id":1,` +
		`"slug":"EXA-QUI-001","created_at":"2026-09-01T10:00:00Z","updated_at":"2026-09-01T10:00:00Z",` +
		`"country":null,"address_line1":null,"address_line2":null,"state":null,"zipcode":null,` +
		`"email":null,"city":null,"url":null,"phone":null,"logo_url":null,"legal_name":null,` +
		`"legal_number":null,"currency":"USD","tax_identification_number":null,"timezone":null,` +
		`"applicable_timezone":"UTC","net_payment_term":null,"finalize_zero_amount_invoice":"inherit",` +
		`"billing_configuration":{"invoice_grace_period":0,"payment_provider":null,"document_locale":null},` +
		`"metadata":[],"taxes":[]}}`
	fullPlan = `{"plan":{"lago_id":"2b902b90-2b90-2b90-2b90-2b902b902b90","name":"Quickstart",` +
		`"invoice_display_name":null,"created_at":"2026-09-01T10:00:00Z","code":"quickstart",` +
		`"interval":"monthly","description":null,"amount_cents":0,"amount_currency":"USD",` +
		`"trial_period":0.0,"pay_in_advance":false,"bill_charges_monthly":null,"active_subscriptions_count":0,` +
		`"draft_invoices_count":0,"parent_id":null,"charges":[],"taxes":[],"minimum_commitment":null,` +
		`"usage_thresholds":[],"entitlements":[]}}`
	fullSubscription = `{"subscription":{"lago_id":"3c903c90-3c90-3c90-3c90-3c903c903c90",` +
		`"external_id":"quickstart_subscription","lago_customer_id":"1a901a90-1a90-1a90-1a90-1a901a901a90",` +
		`"external_customer_id":"quickstart_customer","name":null,"plan_code":"quickstart",` +
		`"status":"active","billing_time":"anniversary","subscription_at":"2026-09-01T10:00:00Z",` +
		`"started_at":"2026-09-01T10:00:00Z","trial_ended_at":null,"ending_at":null,` +
		`"terminated_at":null,"canceled_at":null,"created_at":"2026-09-01T10:00:00Z",` +
		`"previous_plan_code":null,"next_plan_code":null,"downgrade_plan_date":null,` +
		`"current_billing_period_started_at":"2026-09-01T10:00:00Z",` +
		`"current_billing_period_ending_at":"2026-10-01T10:00:00Z","on_termination_credit_note":"credit",` +
		`"on_termination_invoice":"generate","plan":{"lago_id":"2b902b90-2b90-2b90-2b90-2b902b902b90",` +
		`"code":"quickstart","name":"Quickstart","amount_cents":0}}}`
)

// The three creates from the README quickstart are pinned as snapshots. A change to any
// of them is a change to the documented public contract, so it has to be reviewed in the
// diff rather than re-recorded by hand.
func TestMutationDefaultOutputSnapshots(t *testing.T) {
	for _, testCase := range []struct {
		snapshot string
		response string
		argv     []string
	}{
		{
			snapshot: "customers_create_table.txt",
			response: fullCustomer,
			argv:     []string{"customers", "create", "--external-id", "quickstart_customer", "--name", "Quickstart Customer"},
		},
		{
			snapshot: "plans_create_table.txt",
			response: fullPlan,
			argv: []string{"plans", "create", "--name", "Quickstart", "--code", "quickstart",
				"--interval", "monthly", "--amount-cents", "0", "--amount-currency", "USD",
				"--pay-in-advance", "false"},
		},
		{
			snapshot: "subscriptions_create_table.txt",
			response: fullSubscription,
			argv: []string{"subscriptions", "create",
				"--external-customer-id", "quickstart_customer",
				"--external-id", "quickstart_subscription",
				"--plan-code", "quickstart"},
		},
	} {
		t.Run(testCase.snapshot, func(t *testing.T) {
			server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(response, testCase.response)
			})
			profileAt(t, server.URL)

			stdout, _, err := execute(t, "", testCase.argv...)
			if err != nil {
				t.Fatalf("%v failed: %v", testCase.argv, err)
			}
			assertSnapshot(t, testCase.snapshot, stdout)

			// The same command with --output json must still carry the complete
			// resource: the terse default reduces the table, never the payload.
			jsonOut, _, err := execute(t, "", append([]string{"--output", "json"}, testCase.argv...)...)
			if err != nil {
				t.Fatalf("%v --output json failed: %v", testCase.argv, err)
			}
			var full, want any
			if err := json.Unmarshal([]byte(jsonOut), &full); err != nil {
				t.Fatalf("--output json is not JSON: %q", jsonOut)
			}
			if err := json.Unmarshal([]byte(testCase.response), &want); err != nil {
				t.Fatal(err)
			}
			if encoded, _ := json.Marshal(full); string(encoded) != mustCompact(t, testCase.response) {
				t.Errorf("--output json did not return the complete resource:\n got %s\nwant %s", encoded, mustCompact(t, testCase.response))
			}
		})
	}
}

// Every mutation in the manifest must actually render terse, not just the three that
// are snapshotted. This is the guardrail that a new create arriving from a spec bump
// cannot ship with the old full-dump default.
func TestEveryMutationRendersOnlyIdentifiers(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		// A resource carrying one identifier plus attributes that must not appear.
		_, _ = io.WriteString(response, `{"resource":{"lago_id":"res_1","code":"res_code",`+
			`"amount_cents":123456,"created_at":"2026-09-01T10:00:00Z","secret_attribute":"leaked"}}`)
	})
	profileAt(t, server.URL)

	mutations := 0
	for _, operation := range generated.Operations {
		if !operation.Mutation {
			continue
		}
		mutations++
		t.Run(operation.OperationID, func(t *testing.T) {
			app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
			var stdout strings.Builder
			app.Out = &stdout
			root := NewRoot(app)
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			arguments := append([]string{"--api-url", server.URL, "--api-key", "lago_test_FAKE000000000000000000000000",
				"--mode", "test", "--insecure", operation.Resource, operation.Action}, syntheticArguments(operation)...)
			root.SetArgs(arguments)
			if err := root.Execute(); err != nil {
				t.Fatalf("%s failed: %v; args=%v", operation.OperationID, err, arguments)
			}
			out := stdout.String()
			for _, want := range []string{"LAGO_ID", "res_1", "CODE", "res_code"} {
				if !strings.Contains(out, want) {
					t.Errorf("terse output lacks %q:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{"AMOUNT_CENTS", "123456", "CREATED_AT", "secret_attribute", "leaked"} {
				if strings.Contains(out, unwanted) {
					t.Errorf("terse output leaked the full attribute dump (%q):\n%s", unwanted, out)
				}
			}
		})
	}
	if mutations < 100 {
		t.Fatalf("only %d operations are classified as mutations; the generator rule regressed", mutations)
	}
}

// A read is not a mutation. `customers list` and `customers get` must keep printing
// every column, or the terse renderer has been wired to the wrong half of the tree.
func TestReadsKeepTheFullTable(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, fullCustomer)
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "customers", "get", "quickstart_customer")
	if err != nil {
		t.Fatalf("customers get failed: %v", err)
	}
	if !strings.Contains(stdout, "CURRENCY") || !strings.Contains(stdout, "APPLICABLE_TIMEZONE") {
		t.Errorf("a read lost columns to the terse renderer:\n%s", stdout)
	}
}

// A mutation whose response carries no identifier must fall back to the full table.
// Printing nothing at all would be the worse failure the fallback exists to prevent.
func TestMutationWithoutIdentifiersFallsBackToFullTable(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, `{"customer":{"sequential_id":7,"slug":"EXA-QUI-007"}}`)
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "customers", "create", "--external-id", "cus_1", "--name", "Example")
	if err != nil {
		t.Fatalf("customers create failed: %v", err)
	}
	if !strings.Contains(stdout, "SEQUENTIAL_ID") || !strings.Contains(stdout, "EXA-QUI-007") {
		t.Errorf("identifier-less mutation printed nothing useful:\n%q", stdout)
	}
}

// An explicit --query is the operator naming the fields they want. The terse renderer
// must not run first and delete the keys their expression addresses.
func TestExplicitQueryOverridesTheTerseDefault(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, fullCustomer)
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "--output", "json", "--query", "customer.applicable_timezone",
		"customers", "create", "--external-id", "cus_1", "--name", "Example")
	if err != nil {
		t.Fatalf("query on a mutation failed: %v", err)
	}
	if !strings.Contains(stdout, "UTC") {
		t.Errorf("--query on a mutation was reduced away: %q", stdout)
	}
}

// --dry-run prints the request envelope, which has no identifiers. It must render in
// full or the flag would print an empty block for every create.
func TestMutationDryRunStillPrintsTheRequest(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	stdout, _, err := execute(t, "",
		"--api-url", "https://api.getlago.com", "--api-key", "lago_test_FAKE000000000000000000000000",
		"--mode", "test", "--dry-run", "customers", "create", "--external-id", "cus_1", "--name", "Example")
	if err != nil {
		t.Fatalf("dry-run create failed: %v", err)
	}
	if !strings.Contains(stdout, "POST") || !strings.Contains(stdout, "cus_1") {
		t.Errorf("dry-run create printed no request envelope:\n%q", stdout)
	}
}

// syntheticArguments builds the minimum flag set that satisfies a generated command.
func syntheticArguments(operation generated.Operation) []string {
	arguments := make([]string, 0)
	pathParameters := filterParameters(operation.Parameters, "path")
	for _, parameter := range pathParameters {
		arguments = append(arguments, "fake_"+parameter.Name)
	}
	if operation.Dangerous {
		identifier := operation.Resource
		if len(pathParameters) > 0 {
			identifier = "fake_" + pathParameters[len(pathParameters)-1].Name
		}
		arguments = append(arguments, "--confirm", identifier)
	}
	for _, parameter := range operation.Parameters {
		if parameter.In == "query" && parameter.Required {
			arguments = append(arguments, "--"+parameter.Flag, syntheticValue(parameter.Type, parameter.Enum))
		}
	}
	if operation.Body == nil {
		return arguments
	}
	required := false
	for _, field := range operation.Body.Fields {
		if field.Required {
			arguments = append(arguments, "--"+field.Flag, syntheticValue(field.Type, field.Enum))
			required = true
		}
	}
	if !required {
		payload := map[string]any{}
		if operation.Body.Wrapper != "" {
			payload[operation.Body.Wrapper] = map[string]any{}
		}
		encoded, _ := json.Marshal(payload)
		arguments = append(arguments, "--input", string(encoded))
	}
	return arguments
}

func assertSnapshot(t *testing.T, name, actual string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if os.Getenv("LAGO_UPDATE_SNAPSHOTS") == "1" {
		if err := os.WriteFile(path, []byte(actual), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing snapshot %s: %v\n--- got ---\n%s", path, err, actual)
	}
	if actual != string(expected) {
		t.Fatalf("%s changed; this is a public output contract, review the diff before re-recording\n--- want ---\n%s\n--- got ---\n%s", path, expected, actual)
	}
}

func mustCompact(t *testing.T, raw string) string {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}
