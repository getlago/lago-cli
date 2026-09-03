package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// QA E-4, M-list, L-2h: `customers list` in table mode printed the page as one JSON
// string in a single cell. It now prints one row per customer with the declared
// columns, and tells the operator on stderr when there are more pages.
func TestQA_E4_CustomersListDefaultOutputIsATable(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		page := request.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		fmt.Fprintf(response, `{"customers":[{"lago_id":"cus_%[1]s_a","external_id":"acme_%[1]s","name":"Acme %[1]s","email":"ops@acme.test","currency":"USD","created_at":"2026-01-01T00:00:00Z","timezone":"UTC"},{"lago_id":"cus_%[1]s_b","external_id":"globex_%[1]s","name":"Globex \u001b[31m%[1]s\u001b[0m","currency":"EUR"}],"meta":{"current_page":%[1]s,"total_pages":3,"total_count":6}}`, page)
	})
	profileAt(t, server.URL)

	stdout, stderr, err := execute(t, "", "customers", "list")
	if err != nil {
		t.Fatalf("customers list failed: %v", err)
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want header + 2 rows, got %d lines:\n%s", len(lines), stdout)
	}
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "LAGO_ID EXTERNAL_ID NAME EMAIL CURRENCY CREATED_AT" {
		t.Errorf("header = %q", got)
	}
	if !strings.Contains(lines[1], "cus_1_a") || !strings.Contains(lines[2], "globex_1") {
		t.Errorf("rows are not one customer each:\n%s", stdout)
	}
	if strings.Contains(stdout, "\x1b") || !strings.Contains(stdout, `\x1b[31m1\x1b[0m`) {
		t.Errorf("a control byte reached the list table:\n%q", stdout)
	}
	if strings.Contains(stdout, "META") || strings.Contains(stdout, "[{") {
		t.Errorf("meta or raw JSON reached stdout:\n%s", stdout)
	}
	if !strings.Contains(stderr, "page 1 of 3 (6 total); use --page N or --all") {
		t.Errorf("stderr lacks the pagination hint:\n%s", stderr)
	}
}

// QA L-2h: --all renders page by page without buffering, so the header repeats per
// page and the per-page hint is suppressed. Documented trade-off; a script uses
// --output json.
func TestQA_L2h_AllPagesRepeatsHeaderPerPageWithoutHints(t *testing.T) {
	var mutex sync.Mutex
	pages := 0
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		pages++
		mutex.Unlock()
		page := request.URL.Query().Get("page")
		fmt.Fprintf(response, `{"customers":[{"lago_id":"cus_%s","name":"N"}],"meta":{"current_page":%s,"total_pages":3}}`, page, page)
	})
	profileAt(t, server.URL)

	stdout, stderr, err := execute(t, "", "customers", "list", "--all")
	if err != nil {
		t.Fatalf("customers list --all failed: %v", err)
	}
	if strings.Count(stdout, "LAGO_ID") != 3 {
		t.Errorf("want one header per page:\n%s", stdout)
	}
	for _, want := range []string{"cus_1", "cus_2", "cus_3"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("missing %s:\n%s", want, stdout)
		}
	}
	if strings.Contains(stderr, "use --page") {
		t.Errorf("--all printed a pagination hint:\n%s", stderr)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if pages != 3 {
		t.Errorf("fetched %d pages, want 3", pages)
	}
}

// QA M-nested: `plans get` rendered charges as a JSON blob in one cell.
func TestQA_MNested_PlansGetShowsChargesSummary(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"plan":{"lago_id":"pl_1","code":"quickstart","name":"Quickstart","amount_cents":0,"amount_currency":"USD","charges":[{"lago_id":"ch_1","billable_metric_code":"requests","charge_model":"standard","properties":{"amount":"1"}},{"lago_id":"ch_2","billable_metric_code":"storage","charge_model":"graduated","properties":{}}],"minimum_commitment":{"lago_id":"mc_1","amount_cents":1000,"invoice_display_name":"Minimum","plan_code":"quickstart","taxes":[]},"taxes":[]}}`))
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "plans", "get", "quickstart")
	if err != nil {
		t.Fatalf("plans get failed: %v", err)
	}
	if strings.Contains(stdout, "[{") || strings.Contains(stdout, `{"`) {
		t.Errorf("table still contains JSON:\n%s", stdout)
	}
	for _, want := range []string{"2 items: ch_1, ch_2", "lago_id=mc_1", "TAXES", "0 items", "AMOUNT_CENTS", "0"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("table missing %q:\n%s", want, stdout)
		}
	}
}

// A --dry-run envelope is a request, not a resource. With nested values summarised,
// the payload under `body` would collapse to `{2 fields}`, so it prints as JSON.
func TestDryRunEnvelopePrintsAsJSONInTableMode(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", t.TempDir()+"/missing.toml")
	stdout, _, err := execute(t, "",
		"--api-url", "https://api.getlago.com", "--api-key", "lago_test_FAKE000000000000000000000000",
		"--mode", "test", "--dry-run", "api", "POST", "/customers", "--data", `{"customer":{"external_id":"cus_1","name":"Example"}}`)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	for _, want := range []string{`"method": "POST"`, `"external_id": "cus_1"`, `"url":`} {
		if !strings.Contains(stdout, want) {
			t.Errorf("dry-run envelope missing %s:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "fields}") {
		t.Errorf("dry-run body was summarised away:\n%s", stdout)
	}
}

// QA S-22: an API error message is API data too.
func TestQA_S22_ErrorTextIsSanitized(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = response.Write([]byte("{\"status\":422,\"error\":\"bad\\u001b[31m\\nError: fake\",\"code\":\"x\\ry\"}"))
	})
	profileAt(t, server.URL)
	var stdout, stderr bytes.Buffer
	if code := ExecuteArgs([]string{"api", "GET", "/customers"}, strings.NewReader(""), &stdout, &stderr); code == 0 {
		t.Fatal("422 exited 0")
	}
	if strings.ContainsAny(stderr.String(), "\x1b\r") {
		t.Errorf("raw control byte reached stderr:\n%q", stderr.String())
	}
	errorLines := 0
	for _, line := range strings.Split(stderr.String(), "\n") {
		if strings.HasPrefix(line, "Error: ") {
			errorLines++
		}
	}
	if errorLines != 1 {
		t.Errorf("newline in the API message injected a fake error line:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), `Lago code: x\ry`) {
		t.Errorf("code was not escaped visibly:\n%s", stderr.String())
	}
}
