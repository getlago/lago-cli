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
)

// --all must walk every page reported by meta.total_pages and stop exactly there,
// never looping forever against a server that keeps answering.
func TestAllPagesWalksEveryPageAndStops(t *testing.T) {
	var mutex sync.Mutex
	var pages []string

	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		page := request.URL.Query().Get("page")
		mutex.Lock()
		pages = append(pages, page)
		mutex.Unlock()
		if request.URL.Query().Get("per_page") != "100" {
			t.Errorf("--all did not request a full page: %s", request.URL.RawQuery)
		}
		fmt.Fprintf(response, `{"customers":[{"lago_id":"cus_%s"}],"meta":{"current_page":%s,"total_pages":3}}`, page, page)
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "--output", "json", "customers", "list", "--all")
	if err != nil {
		t.Fatalf("customers list --all failed: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if strings.Join(pages, ",") != "1,2,3" {
		t.Fatalf("pages fetched = %v, want 1,2,3", pages)
	}
	for _, want := range []string{"cus_1", "cus_2", "cus_3"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("output is missing %s:\n%s", want, stdout)
		}
	}
}

// A single-page response must not trigger a second request.
func TestAllPagesStopsOnASinglePage(t *testing.T) {
	var mutex sync.Mutex
	requests := 0
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		requests++
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customers":[],"meta":{"current_page":1,"total_pages":1}}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "--output", "json", "customers", "list", "--all"); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

// A response with no usable pagination metadata must stop after one page rather
// than paging forever.
func TestAllPagesStopsWithoutPaginationMetadata(t *testing.T) {
	for _, body := range []string{
		`{"customers":[]}`,
		`{"customers":[],"meta":{}}`,
		`{"customers":[],"meta":{"total_pages":"many"}}`,
		`[]`,
	} {
		t.Run(body, func(t *testing.T) {
			var mutex sync.Mutex
			requests := 0
			server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
				mutex.Lock()
				requests++
				mutex.Unlock()
				_, _ = response.Write([]byte(body))
			})
			profileAt(t, server.URL)

			if _, _, err := execute(t, "", "--output", "json", "customers", "list", "--all"); err != nil {
				t.Fatal(err)
			}
			mutex.Lock()
			defer mutex.Unlock()
			if requests != 1 {
				t.Fatalf("requests = %d for %s, want 1", requests, body)
			}
		})
	}
}

// --all buffers nothing, so a whole-collection JMESPath query cannot be honoured.
// The refusal must teach the alternative rather than just saying no.
func TestAllPagesRefusesAQueryAndTeachesTheAlternative(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"customers":[]}`))
	})
	profileAt(t, server.URL)

	_, _, err := execute(t, "", "--output", "json", "--query", "customers[].lago_id", "customers", "list", "--all")
	if err == nil {
		t.Fatal("--all accepted a full-collection query")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	var appErr *apperr.Error
	if !asAppError(err, &appErr) {
		t.Fatalf("refusal is not a typed error: %v", err)
	}
	if !strings.Contains(appErr.Suggestion, "jq") || !strings.Contains(appErr.Suggestion, "--query") {
		t.Errorf("refusal does not teach the alternative: %q", appErr.Suggestion)
	}
}

// An explicit --page must be honoured as the starting point, and a nonsense value
// must be rejected before any request is issued.
func TestAllPagesValidatesTheStartingPage(t *testing.T) {
	var mutex sync.Mutex
	var first string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		if first == "" {
			first = request.URL.Query().Get("page")
		}
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customers":[],"meta":{"current_page":2,"total_pages":2}}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "--output", "json", "customers", "list", "--all", "--page", "2"); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	if first != "2" {
		t.Errorf("--all started at page %q, want 2", first)
	}
	mutex.Unlock()

	if _, _, err := execute(t, "", "customers", "list", "--all", "--page", "0"); err == nil {
		t.Error("--page 0 was accepted")
	}
	if _, _, err := execute(t, "", "customers", "list", "--all", "--page", "later"); err == nil {
		t.Error("a non-numeric --page was accepted")
	}
}

func TestJSONIntegerAcceptsEveryNumericEncoding(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		value any
		want  int64
		ok    bool
	}{
		{json.Number("3"), 3, true},
		{float64(4), 4, true},
		{int64(5), 5, true},
		{json.Number("9007199254740993"), 9007199254740993, true},
		{float64(2.5), 0, false},
		{json.Number("nope"), 0, false},
		{"7", 0, false},
		{nil, 0, false},
	} {
		got, ok := jsonInteger(testCase.value)
		if ok != testCase.ok || (ok && got != testCase.want) {
			t.Errorf("jsonInteger(%#v) = %d,%v want %d,%v", testCase.value, got, ok, testCase.want, testCase.ok)
		}
	}
}

// --watch re-renders only when the response changes, and refuses an interval short
// enough to hammer the API.
func TestWatchRefusesTooShortAnInterval(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"customers":[]}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "customers", "list", "--watch", "--watch-interval", "10ms"); err == nil {
		t.Fatal("a 10ms watch interval was accepted")
	} else if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}

	if _, _, err := execute(t, "", "logs", "tail", "--interval", "10ms"); err == nil {
		t.Fatal("a 10ms log tail interval was accepted")
	}
}

// Plan import must diff against the server: unchanged plans are skipped, new plans
// created, and changed plans updated. Re-importing must not duplicate anything.
func TestPlanImportDiffsAgainstTheServer(t *testing.T) {
	var mutex sync.Mutex
	var mutations []string

	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			switch {
			case strings.HasSuffix(request.URL.Path, "/plans/unchanged"):
				_, _ = response.Write([]byte(`{"plan":{"code":"unchanged","name":"Same","amount_cents":1000}}`))
			case strings.HasSuffix(request.URL.Path, "/plans/changed"):
				_, _ = response.Write([]byte(`{"plan":{"code":"changed","name":"Old","amount_cents":500}}`))
			default:
				response.WriteHeader(http.StatusNotFound)
				_, _ = response.Write([]byte(`{"status":404,"error":"Not Found"}`))
			}
			return
		}
		mutex.Lock()
		mutations = append(mutations, request.Method+" "+request.URL.Path)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"plan":{"code":"ok"}}`))
	})
	profileAt(t, server.URL)

	path := filepath.Join(t.TempDir(), "plans.json")
	body := `{"plans":[
	  {"code":"unchanged","name":"Same","amount_cents":1000},
	  {"code":"changed","name":"New","amount_cents":1500},
	  {"code":"brand-new","name":"Fresh","amount_cents":2000}
	]}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := execute(t, "", "--output", "json", "plans", "import", path)
	if err != nil {
		t.Fatalf("plans import failed: %v", err)
	}

	var report struct {
		Plans []struct {
			Code   string `json:"code"`
			Action string `json:"action"`
		} `json:"plans"`
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("import output is not JSON: %q", stdout)
	}
	actions := map[string]string{}
	for _, plan := range report.Plans {
		actions[plan.Code] = plan.Action
	}
	if actions["unchanged"] != "unchanged" {
		t.Errorf("an identical plan was not skipped: %v", actions)
	}
	if actions["changed"] != "update" {
		t.Errorf("a changed plan was not updated: %v", actions)
	}
	if actions["brand-new"] != "create" {
		t.Errorf("a new plan was not created: %v", actions)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if len(mutations) != 2 {
		t.Fatalf("import issued %d mutations, want 2: %v", len(mutations), mutations)
	}
}

// Every shape the importer accepts must round-trip, and malformed input must be a
// usage error rather than a partial import.
func TestPlanImportInputShapes(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			response.WriteHeader(http.StatusNotFound)
			_, _ = response.Write([]byte(`{"status":404}`))
			return
		}
		_, _ = response.Write([]byte(`{"plan":{"code":"ok"}}`))
	})
	profileAt(t, server.URL)

	for _, testCase := range []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"single object", `{"code":"one","name":"One"}`, false},
		{"array", `[{"code":"a"},{"code":"b"}]`, false},
		{"named envelope", `{"plans":[{"code":"c"}]}`, false},
		{"not json", `{{{`, true},
		{"plan without a code", `{"plans":[{"name":"no code"}]}`, true},
		{"code is not a string", `{"plans":[{"code":123}]}`, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "plans.json")
			if err := os.WriteFile(path, []byte(testCase.body), 0o600); err != nil {
				t.Fatal(err)
			}
			_, _, err := execute(t, "", "--output", "json", "plans", "import", path)
			if testCase.wantErr && err == nil {
				t.Fatalf("%s was accepted", testCase.name)
			}
			if !testCase.wantErr && err != nil {
				t.Fatalf("%s was rejected: %v", testCase.name, err)
			}
		})
	}

	if _, _, err := execute(t, "", "plans", "import", filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("a missing import file was accepted")
	}
}

func TestSubsetEqualComparesOnlyDeclaredFields(t *testing.T) {
	t.Parallel()
	existing := map[string]any{
		"code": "p", "name": "Plan", "amount_cents": json.Number("1000"),
		"lago_id": "server-only", "created_at": "2026-01-01",
	}
	if !subsetEqual(existing, map[string]any{"code": "p", "name": "Plan"}) {
		t.Error("server-only fields made an unchanged plan look different")
	}
	if subsetEqual(existing, map[string]any{"code": "p", "name": "Renamed"}) {
		t.Error("a changed field was reported as equal")
	}
	if subsetEqual(existing, map[string]any{"code": "p", "new_field": "x"}) {
		t.Error("a field absent from the server was reported as equal")
	}
	if subsetEqual(nil, map[string]any{"code": "p"}) {
		t.Error("a missing existing plan was reported as equal")
	}
	nested := map[string]any{"charges": []any{map[string]any{"code": "c", "extra": 1}}}
	if !subsetEqual(nested, map[string]any{"charges": []any{map[string]any{"code": "c", "extra": 1}}}) {
		t.Error("identical nested structures were reported as different")
	}
}

func TestUnwrapNamedObject(t *testing.T) {
	t.Parallel()
	inner := map[string]any{"code": "p"}
	if got := unwrapNamedObject(map[string]any{"plan": inner}, "plan"); got["code"] != "p" {
		t.Errorf("named envelope was not unwrapped: %v", got)
	}
	// Anything that is not the expected envelope yields nil, which subsetEqual
	// then reports as "different". That fails toward re-sending an update rather
	// than toward silently skipping a plan the operator asked to change.
	for _, unexpected := range []any{inner, "not an object", nil, map[string]any{"other": inner}} {
		if got := unwrapNamedObject(unexpected, "plan"); got != nil {
			t.Errorf("unwrapNamedObject(%#v) = %v, want nil", unexpected, got)
		}
	}
	if subsetEqual(unwrapNamedObject("unexpected", "plan"), map[string]any{"code": "p"}) {
		t.Error("an unrecognised response shape was treated as unchanged")
	}
}

// asAppError is errors.As specialised to the CLI error type.
func asAppError(err error, target **apperr.Error) bool {
	for err != nil {
		if typed, ok := err.(*apperr.Error); ok {
			*target = typed
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}
