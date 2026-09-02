package cli

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/spf13/cobra"
)

// `lago api` is the escape hatch for endpoints the generated tree does not cover.
// It must accept a body from a flag, a file, or stdin, and reject anything else.
func TestAPICommandBodySources(t *testing.T) {
	var mutex sync.Mutex
	var received []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		raw := make([]byte, request.ContentLength)
		if request.ContentLength > 0 {
			_, _ = request.Body.Read(raw)
		}
		mutex.Lock()
		received = append(received, string(raw))
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt_1"}}`))
	})
	profileAt(t, server.URL)

	payload := `{"event":{"code":"requests"}}`
	file := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(file, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := execute(t, "", "--output", "json", "api", "POST", "/events", "--data", payload); err != nil {
		t.Fatalf("inline --data failed: %v", err)
	}
	if _, _, err := execute(t, "", "--output", "json", "api", "POST", "/events", "--data", "@"+file); err != nil {
		t.Fatalf("--data @file failed: %v", err)
	}
	if _, _, err := execute(t, payload, "--output", "json", "api", "POST", "/events", "--data", "-"); err != nil {
		t.Fatalf("--data - failed: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if len(received) != 3 {
		t.Fatalf("received %d bodies, want 3", len(received))
	}
	for index, body := range received {
		if !strings.Contains(body, "requests") {
			t.Errorf("body %d did not reach the server: %q", index, body)
		}
	}
}

func TestAPICommandRejectsBadInput(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{}`))
	})
	profileAt(t, server.URL)

	for _, testCase := range []struct {
		name string
		argv []string
	}{
		{"absolute URL", []string{"api", "GET", "https://evil.example.com/customers"}},
		{"malformed JSON body", []string{"api", "POST", "/events", "--data", "{not json"}},
		{"missing file after @", []string{"api", "POST", "/events", "--data", "@"}},
		{"header without a colon", []string{"api", "GET", "/customers", "--header", "NoColon"}},
		{"header with an empty name", []string{"api", "GET", "/customers", "--header", ": value"}},
		{"caller-supplied Authorization", []string{"api", "GET", "/customers", "--header", "Authorization: Bearer stolen"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, _, err := execute(t, "", testCase.argv...); err == nil {
				t.Fatalf("%s was accepted", testCase.name)
			} else if apperr.ExitCode(err) != apperr.ExitUsage {
				t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
			}
		})
	}

	// A body file that cannot be read is an I/O failure, not a malformed command,
	// so it carries the general exit code rather than the usage one.
	_, _, err := execute(t, "", "api", "POST", "/events", "--data", "@/nonexistent/payload.json")
	if err == nil {
		t.Fatal("an unreadable body file was accepted")
	}
	if apperr.ExitCode(err) != apperr.ExitGeneral {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitGeneral)
	}
}

// Custom headers must reach the server, and the query string in the path argument
// must be preserved rather than dropped.
func TestAPICommandForwardsHeadersAndQuery(t *testing.T) {
	var mutex sync.Mutex
	var header, query, idempotency string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		header = request.Header.Get("X-Trace")
		query = request.URL.RawQuery
		idempotency = request.Header.Get("Idempotency-Key")
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customers":[]}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "--output", "json", "api", "GET", "/customers?page=2&per_page=5",
		"--header", "X-Trace: abc"); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if header != "abc" {
		t.Errorf("custom header = %q", header)
	}
	if !strings.Contains(query, "page=2") || !strings.Contains(query, "per_page=5") {
		t.Errorf("query lost: %q", query)
	}
	if idempotency != "" {
		t.Errorf("Idempotency-Key header = %q, want none: lago-api does not read it", idempotency)
	}
}

func TestReadDataHandlesEverySource(t *testing.T) {
	t.Parallel()
	if body, err := readData(strings.NewReader(""), ""); err != nil || body != nil {
		t.Errorf("empty --data = %q %v, want nil", body, err)
	}
	if body, err := readData(strings.NewReader(""), "   "); err != nil || body != nil {
		t.Errorf("blank --data = %q %v, want nil", body, err)
	}
	if body, err := readData(strings.NewReader(`{"a":1}`), "-"); err != nil || !strings.Contains(string(body), `"a"`) {
		t.Errorf("stdin --data = %q %v", body, err)
	}
	if _, err := readData(strings.NewReader("{broken"), "-"); err == nil {
		t.Error("malformed stdin JSON was accepted")
	}
	if _, err := readData(strings.NewReader(""), "@"); err == nil {
		t.Error("a bare @ was accepted")
	}
}

func TestValidateJSONPreservesExactBytes(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"total_amount_cents":9007199254740993}`)
	validated, err := validateJSON(payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(validated) != string(payload) {
		t.Fatalf("validateJSON rewrote the body: %s", validated)
	}
	if body, err := validateJSON([]byte("  \n ")); err != nil || body != nil {
		t.Errorf("blank body = %q %v, want nil", body, err)
	}
	if _, err := validateJSON([]byte("{")); err == nil {
		t.Error("malformed JSON was accepted")
	}
}

// Generated flags are typed from the spec. A wrong type must be rejected before the
// request, and a monetary "number" must stay an exact decimal string.
func TestGeneratedValueParsingIsTypeSafe(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		raw, valueType string
		wantErr        bool
	}{
		{"true", "boolean", false},
		{"maybe", "boolean", true},
		{"42", "integer", false},
		{"4.2", "integer", true},
		{"not-a-number", "integer", true},
		{"10.50", "number", false},
		{"abc", "number", true},
		{`{"a":1}`, "object", false},
		{"{broken", "object", true},
		{`["a"]`, "array", false},
		{"plain", "string", false},
	} {
		_, err := parseGeneratedValue(testCase.raw, testCase.valueType, false)
		if testCase.wantErr && err == nil {
			t.Errorf("%q as %s was accepted", testCase.raw, testCase.valueType)
		}
		if !testCase.wantErr && err != nil {
			t.Errorf("%q as %s was rejected: %v", testCase.raw, testCase.valueType, err)
		}
	}

	// An exact decimal must survive as its original digits, never as a float.
	value, err := parseGeneratedValue("10.50", "number", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, isFloat := value.(float64); isFloat {
		t.Errorf("a monetary decimal was parsed into a float64: %#v", value)
	}
}

// Repeated query parameters accept both JSON arrays and comma-separated lists.
func TestQueryValuesAcceptsBothListForms(t *testing.T) {
	t.Parallel()
	if got := queryValues("single", "string"); len(got) != 1 || got[0] != "single" {
		t.Errorf("scalar = %v", got)
	}
	if got := queryValues(`["a","b"]`, "array"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("JSON array = %v", got)
	}
	if got := queryValues("a, b ,c", "array"); len(got) != 3 || got[2] != "c" {
		t.Errorf("comma list = %v", got)
	}
	if got := queryValues("a,,b", "array"); len(got) != 2 {
		t.Errorf("empty entries were not skipped: %v", got)
	}
	if got := queryValues("", "array"); len(got) != 0 {
		t.Errorf("empty list = %v", got)
	}
}

// --watch re-renders only when the payload changes and stops when the context is
// cancelled, so it never leaks a goroutine or spins after Ctrl-C.
func TestWatchRerendersOnlyOnChangeAndStopsOnCancel(t *testing.T) {
	var mutex sync.Mutex
	polls := 0
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		polls++
		current := polls
		mutex.Unlock()
		if current < 3 {
			_, _ = response.Write([]byte(`{"customers":[{"lago_id":"cus_1"}]}`))
			return
		}
		_, _ = response.Write([]byte(`{"customers":[{"lago_id":"cus_2"}]}`))
	})
	profileAt(t, server.URL)

	var stdout, stderr strings.Builder
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"--output", "json", "customers", "list", "--watch", "--watch-interval", "500ms"})
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 1600*time.Millisecond)
	defer cancel()
	if err := root.ExecuteContext(ctx); err != nil {
		t.Fatalf("watch exited with an error: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if polls < 3 {
		t.Fatalf("watch polled %d times, want at least 3", polls)
	}
	if strings.Count(stdout.String(), "cus_1") != 1 {
		t.Errorf("an unchanged response was rendered more than once:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "cus_2") {
		t.Errorf("the changed response was never rendered:\n%s", stdout.String())
	}
}

// A watch that hits an API error must surface it rather than polling forever.
func TestWatchSurfacesAPIErrors(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
		_, _ = response.Write([]byte(`{"status":401,"error":"Unauthorized"}`))
	})
	profileAt(t, server.URL)

	var stdout, stderr strings.Builder
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"customers", "list", "--watch", "--watch-interval", "500ms"})
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := root.ExecuteContext(ctx)
	if err == nil {
		t.Fatal("watch swallowed an authentication failure")
	}
	if apperr.ExitCode(err) != apperr.ExitAuth {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitAuth)
	}
}

// logs tail is runWatch behind a filter surface. The filters must reach the query.
func TestLogsTailBuildsItsFilterQuery(t *testing.T) {
	var mutex sync.Mutex
	var query string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		if query == "" {
			query = request.URL.RawQuery
		}
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"api_logs":[]}`))
	})
	profileAt(t, server.URL)

	var stdout, stderr strings.Builder
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"--output", "json", "logs", "tail", "--interval", "500ms",
		"--status", "4xx", "--status", "500", "--method", "post", "--resource", "/customers"})
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	if err := root.ExecuteContext(ctx); err != nil {
		t.Fatalf("logs tail failed: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	for _, want := range []string{"http_statuses%5B%5D=4xx", "http_statuses%5B%5D=500", "http_methods%5B%5D=POST", "request_paths=%2Fcustomers"} {
		if !strings.Contains(query, want) {
			t.Errorf("log filter query missing %s: %s", want, query)
		}
	}
}

func TestDefaultTransactionIDIsUniqueAndWellFormed(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for index := 0; index < 200; index++ {
		key := defaultTransactionID()
		if len(key) != 36 || strings.Count(key, "-") != 4 {
			t.Fatalf("generated transaction ID %q is not a UUID", key)
		}
		if seen[key] {
			t.Fatalf("generated a duplicate transaction ID: %s", key)
		}
		seen[key] = true
	}
}

func TestHelpersHandleEdgeCases(t *testing.T) {
	t.Parallel()
	if errorText(nil) != "" {
		t.Error("errorText(nil) is not empty")
	}
	if got := firstNonBlank("", "  ", "value"); got != "value" {
		t.Errorf("firstNonBlank = %q", got)
	}
	if got := firstNonBlank("", "   "); got != "" {
		t.Errorf("firstNonBlank = %q, want empty", got)
	}
	if findDirectChild(&cobra.Command{}, "absent") != nil {
		t.Error("findDirectChild found a command in an empty tree")
	}
	if len(generated.Operations) == 0 {
		t.Fatal("the generated operation table is empty")
	}
	if diagnostic("check", true, "detail")["name"] != "check" {
		t.Error("diagnostic dropped its name")
	}
}

func TestOrganizationIdentityToleratesUnexpectedShapes(t *testing.T) {
	t.Parallel()
	id, name := organizationIdentity([]byte(`{"organization":{"lago_id":"org_1","name":"Example"}}`))
	if id != "org_1" || name != "Example" {
		t.Fatalf("identity = %q %q", id, name)
	}
	for _, body := range []string{`{}`, `{"organization":null}`, `{"organization":{}}`, `not json`, ``} {
		if gotID, gotName := organizationIdentity([]byte(body)); gotID != "" || gotName != "" {
			t.Errorf("body %q produced %q %q, want empty", body, gotID, gotName)
		}
	}
}

func TestJSONErrorEnvelopeIsStable(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("X-Request-Id", "req_stable")
		response.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = response.Write([]byte(`{"status":422,"error":"Unprocessable","code":"value_is_invalid","error_details":{"code":["invalid"]}}`))
	})
	profileAt(t, server.URL)

	_, stderr, err := execute(t, "", "--output", "json", "api", "GET", "/customers")
	if err == nil {
		t.Fatal("a 422 did not surface as an error")
	}
	var payload struct {
		Error struct {
			Status    int    `json:"status"`
			Code      string `json:"code"`
			RequestID string `json:"request_id"`
			Message   string `json:"message"`
		} `json:"error"`
	}
	encoded := apperr.Encode(err)
	if decodeErr := json.Unmarshal(encoded, &payload); decodeErr != nil {
		t.Fatalf("error envelope is not JSON: %s", encoded)
	}
	if payload.Error.Status != 422 || payload.Error.Code != "value_is_invalid" {
		t.Errorf("error envelope lost the API fields: %s", encoded)
	}
	if payload.Error.RequestID != "req_stable" {
		t.Errorf("error envelope lost the request ID: %s", encoded)
	}
	_ = stderr
}

// QA S-28: `lago api DELETE /customers/x` bypassed the confirmation gate entirely. In
// live mode a raw destructive request now needs the same --confirm as a generated
// delete, with the request path as the identifier.
func TestQA_S28_APIDeleteInLiveModeRequiresConfirmation(t *testing.T) {
	var mutex sync.Mutex
	reached := 0
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		reached++
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1"}}`))
	})
	setCleanEnvironment(t)
	writeProfile(t, config.ModeLive, server.URL)

	_, _, err := execute(t, "", "api", "DELETE", "/customers/cus_1")
	if err == nil {
		t.Fatal("live DELETE ran without confirmation")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	if !strings.Contains(err.Error(), "confirmation required for /customers/cus_1") {
		t.Errorf("refusal does not name the path: %v", err)
	}
	mutex.Lock()
	if reached != 0 {
		t.Fatalf("request reached the API %d time(s) before confirmation", reached)
	}
	mutex.Unlock()

	if _, _, err := execute(t, "", "--output", "json", "api", "DELETE", "/customers/cus_1", "--confirm", "/customers/cus_1"); err != nil {
		t.Fatalf("confirmed DELETE failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if reached != 1 {
		t.Errorf("confirmed request reached the API %d time(s), want 1", reached)
	}
}

// The gate follows the spec classification, so a POST to a destructive path is gated
// even though the verb is not DELETE.
func TestQA_S28_APIDestructivePathInLiveModeRequiresConfirmation(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) {
		t.Error("unconfirmed void reached the API")
	})
	setCleanEnvironment(t)
	writeProfile(t, config.ModeLive, server.URL)

	_, _, err := execute(t, "", "api", "POST", "/invoices/inv_1/void")
	if err == nil {
		t.Fatal("live void ran without confirmation")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
}

// Test mode keeps `lago api` as the raw escape hatch it exists to be: no gate.
func TestQA_S28_APIInTestModeStaysUngated(t *testing.T) {
	var mutex sync.Mutex
	var method string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		method = request.Method
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1"}}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "--output", "json", "api", "DELETE", "/customers/cus_1"); err != nil {
		t.Fatalf("test-mode DELETE was gated: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if method != http.MethodDelete {
		t.Errorf("method reached = %q, want DELETE", method)
	}
}

// QA F-12: the flag is gone from every command, so a script still passing it fails at
// parse time instead of sending a header nothing reads.
func TestQA_F12_NoCommandOffersIdempotencyKeyFlag(t *testing.T) {
	t.Parallel()
	root := NewRoot(NewApp(strings.NewReader(""), io.Discard, io.Discard, "test"))
	var walk func(cmd *cobra.Command)
	count := 0
	walk = func(cmd *cobra.Command) {
		count++
		if cmd.Flags().Lookup("idempotency-key") != nil || cmd.PersistentFlags().Lookup("idempotency-key") != nil {
			t.Errorf("%s still offers --idempotency-key", cmd.CommandPath())
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
	if count < 200 {
		t.Fatalf("walked only %d commands", count)
	}
}

// The header must not come back through any code path. Comments may mention it; a
// string literal that would be handed to http.Header may not.
func TestQA_F12_IdempotencyKeyHeaderIsAbsentFromSource(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if name := entry.Name(); name == ".git" || name == "dist-channels" || name == "man" || name == "docs" || name == "completions" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), `"Idempotency-Key"`) {
			t.Errorf("%s still names the Idempotency-Key header as a string literal", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
