package cli

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// A query that matches nothing must say so on stderr while stdout keeps `null`.
//
// Lago wraps every response, so `--query lago_id` against `{"customers": [...]}` is a
// valid expression that matches nothing. QA read the resulting `null` as "no data" twice.
// stdout is unchanged so a script parsing it does not have to care that a human was told
// something; the two streams are asserted separately for exactly that reason.
func TestEmptyQueryMatchPrintsAHintWithoutChangingStdout(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, `{"customers":[{"lago_id":"cus_1","external_id":"acme"}],"meta":{"total_pages":1}}`)
	})
	profileAt(t, server.URL)

	stdout, stderr, err := execute(t, "", "--output", "json", "--query", "lago_id", "customers", "list")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if strings.TrimSpace(stdout) != "null" {
		t.Errorf("stdout = %q, want null so scripts parse unchanged", stdout)
	}
	if !strings.Contains(stderr, "query matched nothing") {
		t.Errorf("stderr carries no hint: %q", stderr)
	}
	for _, key := range []string{"customers", "meta"} {
		if !strings.Contains(stderr, key) {
			t.Errorf("hint does not name the available key %q: %q", key, stderr)
		}
	}
}

// The hint must not fire when the query worked, nor when the response itself was null:
// a null answer is an answer, and a hint on every empty result trains people to ignore it.
func TestQueryHintStaysQuietWhenThereIsNothingToExplain(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		response string
		query    string
	}{
		{name: "successful query", response: `{"customers":[{"lago_id":"cus_1"}]}`, query: "customers[].lago_id"},
		{name: "query with no flag", response: `{"customers":[]}`, query: ""},
		{name: "legitimately empty list", response: `{"customers":[],"meta":{}}`, query: "customers"},
		{name: "response is null", response: `null`, query: "anything"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(response, testCase.response)
			})
			profileAt(t, server.URL)

			argv := []string{"--output", "json", "customers", "list"}
			if testCase.query != "" {
				argv = append([]string{"--query", testCase.query}, argv...)
			}
			_, stderr, err := execute(t, "", argv...)
			if err != nil {
				t.Fatalf("failed: %v", err)
			}
			if strings.Contains(stderr, "query matched nothing") {
				t.Errorf("hint fired when it should not have: %q", stderr)
			}
		})
	}
}

// --query without an explicit --output switches to JSON and announces it. An empty table
// is not an answer, and a silent format change is not one either.
func TestQueryWithoutAnOutputFlagSwitchesToJSON(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, `{"customers":[{"lago_id":"cus_1","external_id":"acme"}],"meta":{}}`)
	})
	profileAt(t, server.URL)

	stdout, stderr, err := execute(t, "", "--query", "customers[].external_id", "customers", "list")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Errorf("stdout is not JSON: %q", stdout)
	}
	if !strings.Contains(stdout, "acme") {
		t.Errorf("stdout lost the queried value: %q", stdout)
	}
	if !strings.Contains(stderr, "--query implies --output json") {
		t.Errorf("the format switch was not announced: %q", stderr)
	}
}

// An explicit --output is always honoured, including --output table. The switch adds a
// sensible default, it does not take the table renderer away.
func TestExplicitOutputBeatsTheQueryDefault(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, `{"customers":[{"lago_id":"cus_1","external_id":"acme"}],"meta":{}}`)
	})
	profileAt(t, server.URL)

	stdout, stderr, err := execute(t, "", "--output", "table", "--query", "customers[].external_id", "customers", "list")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Errorf("--output table was overridden: %q", stdout)
	}
	if !strings.Contains(stdout, "acme") {
		t.Errorf("table output lost the queried value: %q", stdout)
	}
	if strings.Contains(stderr, "--query implies") {
		t.Errorf("the switch notice fired despite an explicit --output: %q", stderr)
	}

	yamlOut, _, err := execute(t, "", "--output", "yaml", "--query", "customers[].external_id", "customers", "list")
	if err != nil {
		t.Fatalf("yaml query failed: %v", err)
	}
	if !strings.Contains(yamlOut, "- acme") {
		t.Errorf("--output yaml was overridden: %q", yamlOut)
	}
}

// A wrong ID species must surface the API's 404 as exit 4 with a message naming the
// resource type and the value. QA passed a plan code as --external-subscription-id and
// got a bare "Not Found" that read as "this subscription has no usage".
func TestWrongIDSpeciesNamesWhatWasNotFound(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		argv     []string
		wantKind string
		wantID   string
	}{
		{
			name:     "plan code passed as a subscription external id",
			argv:     []string{"events", "send", "--external-subscription-id", "ai_plan_gpt4_tokens", "--code", "api_calls"},
			wantKind: "subscription",
			wantID:   "ai_plan_gpt4_tokens",
		},
		{
			name:     "plan code passed as a subscription path argument",
			argv:     []string{"subscriptions", "get", "ai_plan_gpt4_tokens"},
			wantKind: "subscription",
			wantID:   "ai_plan_gpt4_tokens",
		},
		{
			name:     "unknown customer external id",
			argv:     []string{"customers", "get", "cus_does_not_exist"},
			wantKind: "customer",
			wantID:   "cus_does_not_exist",
		},
		{
			name:     "unknown plan code",
			argv:     []string{"plans", "get", "no_such_plan"},
			wantKind: "plan",
			wantID:   "no_such_plan",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
				response.Header().Set("X-Request-Id", "req_not_found")
				response.WriteHeader(http.StatusNotFound)
				// This is all Lago sends. Everything useful in the message has to come
				// from what the CLI knows it asked for.
				_, _ = io.WriteString(response, `{"status":404,"error":"Not Found","code":"not_found"}`)
			})
			profileAt(t, server.URL)

			_, _, err := execute(t, "", testCase.argv...)
			if err == nil {
				t.Fatal("a 404 was reported as success")
			}
			if apperr.ExitCode(err) != apperr.ExitNotFound {
				t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitNotFound)
			}
			var appErr *apperr.Error
			if !errorsAs(err, &appErr) {
				t.Fatalf("not a structured error: %v", err)
			}
			if !strings.Contains(appErr.Message, testCase.wantKind) {
				t.Errorf("message does not name the resource type %q: %s", testCase.wantKind, appErr.Message)
			}
			if !strings.Contains(appErr.Message, testCase.wantID) {
				t.Errorf("message does not name the value %q: %s", testCase.wantID, appErr.Message)
			}
			if strings.TrimSpace(appErr.Message) == "Not Found" {
				t.Errorf("the bare API message survived: %s", appErr.Message)
			}
			if appErr.Suggestion == "" {
				t.Error("a wrong-identifier 404 carries no suggestion")
			}
			if appErr.RequestID != "req_not_found" {
				t.Errorf("request ID lost: %q", appErr.RequestID)
			}
		})
	}
}

// Only identifier-shaped fields become part of the message. A create that 404s must not
// echo the customer's name or the amount back as "not found".
func TestOnlyIdentifiersAppearInANotFoundMessage(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(response, `{"status":404,"error":"Not Found"}`)
	})
	profileAt(t, server.URL)

	_, _, err := execute(t, "", "customers", "create", "--external-id", "cus_1", "--name", "Sensitive Customer Name")
	if err == nil {
		t.Fatal("a 404 was reported as success")
	}
	if !strings.Contains(err.Error(), "cus_1") {
		t.Errorf("the identifier is missing from the message: %v", err)
	}
	if strings.Contains(err.Error(), "Sensitive Customer Name") {
		t.Errorf("a non-identifier field leaked into the not-found message: %v", err)
	}
}

// Statuses other than 404 keep the API's own message: a 422's validation detail is far
// more useful than a list of identifiers.
func TestNonNotFoundStatusesKeepTheAPIMessage(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(response, `{"status":422,"error":"code is invalid","code":"validation_errors"}`)
	})
	profileAt(t, server.URL)

	_, _, err := execute(t, "", "customers", "get", "cus_1")
	if err == nil {
		t.Fatal("a 422 was reported as success")
	}
	if !strings.Contains(err.Error(), "code is invalid") {
		t.Errorf("the API's validation message was replaced: %v", err)
	}
	if apperr.ExitCode(err) != apperr.ExitValidation {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitValidation)
	}
}

// subjectKind is what turns a flag name into the noun in the message. Getting it wrong
// is how "no subscription ai_plan_x exists" becomes "no external subscription id exists".
func TestSubjectKindNamesTheResource(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ field, resource, want string }{
		{"external_subscription_id", "events", "subscription"},
		{"subscription_external_id", "subscriptions", "subscription"},
		{"external_customer_id", "customers", "customer"},
		{"lago_customer_id", "wallets", "customer"},
		{"plan_code", "plans", "plan"},
		{"billable_metric_id", "billable-metrics", "billable metric"},
		{"code", "plans", "plan"},
		{"id", "customers", "customer"},
		{"external_id", "billable-metrics", "billable metric"},
		{"code", "billing-entities", "billing entity"},
		{"code", "taxes", "tax"},
		{"code", "credit-notes", "credit note"},
	} {
		if got := subjectKind(testCase.field, testCase.resource); got != testCase.want {
			t.Errorf("subjectKind(%q, %q) = %q, want %q", testCase.field, testCase.resource, got, testCase.want)
		}
	}
}

func TestIsIdentifierName(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]bool{
		"id": true, "code": true, "external_id": true, "lago_id": true,
		"external_subscription_id": true, "plan_code": true, "billable_metric_id": true,
		"name": false, "amount_cents": false, "currency": false, "interval": false,
		"aggregation_type": false, "identifier": false, "encoded": false,
	} {
		if got := isIdentifierName(name); got != want {
			t.Errorf("isIdentifierName(%q) = %v, want %v", name, got, want)
		}
	}
}

func errorsAs(err error, target **apperr.Error) bool {
	appErr, ok := err.(*apperr.Error)
	if ok {
		*target = appErr
	}
	return ok
}
