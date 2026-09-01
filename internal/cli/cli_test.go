package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/transport"
)

func TestFirstRunIsThreeFriendlyLines(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	var stdout, stderr bytes.Buffer
	if code := Execute(strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr.String())
	}
	if lines := strings.Count(strings.TrimSpace(stdout.String()), "\n") + 1; lines != 3 {
		t.Fatalf("first-run output has %d lines: %q", lines, stdout.String())
	}
	if !strings.Contains(stdout.String(), "lago init") {
		t.Fatalf("first-run output lacks init pointer: %q", stdout.String())
	}
}

func TestTimingReportsCLIOverheadSeparately(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	app := NewApp(strings.NewReader(""), io.Discard, &stderr, "test")
	app.timing = true
	app.Start = time.Now().Add(-20 * time.Millisecond)
	response := &transport.Response{Timing: transport.Timing{Total: 10 * time.Millisecond, RoundTrip: 8 * time.Millisecond}}
	if err := app.Render(map[string]any{}, response); err != nil {
		t.Fatal(err)
	}
	if response.Timing.CLIOverhead < 9*time.Millisecond || !strings.Contains(stderr.String(), "cli_overhead") {
		t.Fatalf("timing=%#v output=%s", response.Timing, stderr.String())
	}
}

func TestCLIOverheadP95Budget(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	durations := make([]time.Duration, 40)
	for index := range durations {
		started := time.Now()
		app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
		root := NewRoot(app)
		root.SetArgs([]string{"--api-url", "https://api.getlago.com", "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--dry-run", "billable-metrics", "create", "--name", "Requests", "--code", "requests", "--aggregation-type", "count_agg"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		durations[index] = time.Since(started)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[(len(durations)*95+99)/100-1]
	if p95 > 50*time.Millisecond {
		t.Fatalf("CLI p95 dry-run overhead=%s, budget=50ms", p95)
	}
}

func TestRootHelpSnapshot(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	root := NewRoot(NewApp(strings.NewReader(""), io.Discard, io.Discard, "test"))
	root.SetOut(&output)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join("testdata", "root_help.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != string(expected) {
		t.Fatalf("root help changed; review the public UX and update testdata/root_help.txt\n--- got ---\n%s", output.String())
	}
}

func TestGeneratedGoldenPathDryRunIsRedacted(t *testing.T) {
	setCleanEnvironment(t)
	var stdout, stderr bytes.Buffer
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{
		"--api-url", "https://api.getlago.com", "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--dry-run", "--output", "json",
		"billable-metrics", "create", "--name", "Requests", "--code", "requests", "--aggregation-type", "count_agg",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"method": "POST"`, `"aggregation_type": "count_agg"`, `Bearer [REDACTED]`} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("dry run %q lacks %q", stdout.String(), expected)
		}
	}
	if strings.Contains(stdout.String(), "abcdefghijklmnopqrstuv") {
		t.Fatal("dry-run output leaked the API key")
	}
}

func TestInitValidatesOrganizationBeforeSavingProfile(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/organizations" {
			t.Errorf("path = %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer lago_test_FAKEabcdefghijklmnopqrstuv" {
			t.Error("missing API key")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"organization":{"lago_id":"org_fake","name":"Test Organization"}}`)
	}))
	defer server.Close()
	var stdout, stderr bytes.Buffer
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"init", "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--region", "self-hosted", "--api-url", server.URL, "--insecure", "--mode", "test", "--profile", "staging"})
	if err := root.Execute(); err != nil {
		t.Fatalf("init failed: %v; stderr=%s", err, stderr.String())
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	profile := loaded.Profiles["staging"]
	if profile.OrganizationID != "org_fake" || profile.Organization != "Test Organization" || profile.Mode != config.ModeTest {
		t.Fatalf("saved profile = %#v", profile)
	}
	if runtime.GOOS != "windows" {
		mode, err := config.FileMode(path)
		if err != nil || mode != 0o600 {
			t.Fatalf("mode=%o err=%v", mode, err)
		}
	}
}

func TestEventNDJSONDryRunStreamsWithGeneratedIDs(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "events.ndjson")
	data := "{\"code\":\"requests\",\"external_subscription_id\":\"sub_1\"}\n{\"event\":{\"code\":\"requests\",\"external_subscription_id\":\"sub_2\",\"transaction_id\":\"event_2\"}}\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"--api-url", "https://api.getlago.com", "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--dry-run", "--output", "json", "events", "send", "--file", path, "--concurrency", "2"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"queued": 2`, `"sent": 2`, `"failed": 0`, `"dry_run": true`} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("summary %q lacks %q", stdout.String(), expected)
		}
	}
}

func TestEventStreamRetriesWithStablePerEventIdempotency(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "events.ndjson")
	data := "{\"code\":\"requests\",\"external_subscription_id\":\"sub_1\"}\n{\"code\":\"requests\",\"external_subscription_id\":\"sub_2\"}\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	var mutex sync.Mutex
	attempts := map[string]int{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		key := request.Header.Get("Idempotency-Key")
		if key == "" {
			t.Error("event request has no idempotency key")
		}
		var payload map[string]map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload["event"]["transaction_id"] != key {
			t.Errorf("transaction ID %v does not match key %s", payload["event"]["transaction_id"], key)
		}
		mutex.Lock()
		attempts[key]++
		attempt := attempts[key]
		mutex.Unlock()
		if attempt == 1 {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{}`)
	}))
	defer server.Close()
	var stdout, stderr bytes.Buffer
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"--api-url", server.URL, "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--insecure", "--output", "json", "events", "send", "--file", path, "--concurrency", "2"})
	if err := root.Execute(); err != nil {
		t.Fatalf("stream failed: %v stderr=%s", err, stderr.String())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if len(attempts) != 2 {
		t.Fatalf("transaction IDs = %#v", attempts)
	}
	for key, count := range attempts {
		if count != 2 {
			t.Errorf("key %s attempts=%d", key, count)
		}
	}
	if !strings.Contains(stdout.String(), `"retried": 2`) || !strings.Contains(stdout.String(), `"sent": 2`) {
		t.Fatalf("summary=%s", stdout.String())
	}
}

func TestRawAPIRejectsAbsolutePath(t *testing.T) {
	setCleanEnvironment(t)
	app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"api", "GET", "https://attacker.example.test/customers"})
	if err := root.Execute(); err == nil {
		t.Fatal("absolute path unexpectedly succeeded")
	}
}

func TestPlansImportDryRunReadsExistingStateButSendsNoMutation(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "plans.json")
	if err := os.WriteFile(path, []byte(`[{"name":"Test","code":"test-plan","interval":"monthly","amount_cents":0,"amount_currency":"USD","pay_in_advance":false}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		if request.Method != http.MethodGet || request.URL.Path != "/api/v1/plans/test-plan" {
			t.Errorf("unexpected mutation in dry run: %s %s", request.Method, request.URL.Path)
		}
		response.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(response, `{"error":"not found"}`)
	}))
	defer server.Close()
	var stdout bytes.Buffer
	app := NewApp(strings.NewReader(""), &stdout, io.Discard, "test")
	root := NewRoot(app)
	root.SetArgs([]string{"--api-url", server.URL, "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--insecure", "--dry-run", "--output", "json", "plans", "import", path})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("network requests=%d, want one read-only diff probe", requests)
	}
	if !strings.Contains(stdout.String(), `"action": "create"`) || !strings.Contains(stdout.String(), `"method": "POST"`) {
		t.Fatalf("dry-run diff=%s", stdout.String())
	}
}

func TestJSONErrorSnapshot(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	var stderr bytes.Buffer
	code := ExecuteArgs([]string{"--output", "json", "api", "GET", "https://attacker.example.test/customers"}, strings.NewReader(""), io.Discard, &stderr)
	if code != 2 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}
	expected, err := os.ReadFile(filepath.Join("testdata", "absolute_api_error.json"))
	if err != nil {
		t.Fatal(err)
	}
	if stderr.String() != string(expected) {
		t.Fatalf("JSON error contract changed\n--- got ---\n%s", stderr.String())
	}
}

func TestUsageErrorsUseFrozenExitCode(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	var stderr bytes.Buffer
	if code := ExecuteArgs([]string{"not-a-command"}, strings.NewReader(""), io.Discard, &stderr); code != 2 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr.String())
	}
}

func TestConfiguredAliasExpandsWithoutShadowingBuiltins(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)
	cfg := config.NewFile()
	cfg.Aliases["v"] = []string{"version", "--output", "json"}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if code := ExecuteArgs([]string{"v"}, strings.NewReader(""), &stdout, io.Discard); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(stdout.String(), `"spec_version": "1.52.1"`) {
		t.Fatalf("alias output=%s", stdout.String())
	}
}

func TestFrozenHTTPExitCodeContract(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		statusText := strings.TrimPrefix(request.URL.Path, "/api/v1/")
		status, err := strconv.Atoi(statusText)
		if err != nil {
			t.Error(err)
		}
		response.Header().Set("X-Request-Id", "req_fake_exit")
		response.WriteHeader(status)
		_, _ = io.WriteString(response, `{"error":"synthetic","code":"synthetic"}`)
	}))
	defer server.Close()
	tests := []struct {
		status int
		exit   int
	}{{401, 3}, {404, 4}, {422, 5}, {429, 6}, {500, 7}}
	for _, test := range tests {
		t.Run(strconv.Itoa(test.status), func(t *testing.T) {
			var stderr bytes.Buffer
			arguments := []string{"--api-url", server.URL, "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--insecure", "--no-retry", "--output", "json", "api", "GET", "/" + strconv.Itoa(test.status)}
			if code := ExecuteArgs(arguments, strings.NewReader(""), io.Discard, &stderr); code != test.exit {
				t.Fatalf("HTTP %d exit=%d want=%d stderr=%s", test.status, code, test.exit, stderr.String())
			}
			if !strings.Contains(stderr.String(), "req_fake_exit") {
				t.Fatalf("HTTP %d error omitted request ID: %s", test.status, stderr.String())
			}
		})
	}
}

func TestGenericAndNetworkExitCodes(t *testing.T) {
	setCleanEnvironment(t)
	directory := t.TempDir()
	malformed := filepath.Join(directory, "malformed.toml")
	if err := os.WriteFile(malformed, []byte("not = [toml"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAGO_CONFIG_FILE", malformed)
	if code := ExecuteArgs([]string{"version"}, strings.NewReader(""), io.Discard, io.Discard); code != 1 {
		t.Fatalf("malformed config exit=%d, want 1", code)
	}
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(directory, "missing.toml"))
	arguments := []string{"--api-url", "https://127.0.0.1:1", "--api-key", "lago_test_FAKEabcdefghijklmnopqrstuv", "--mode", "test", "--timeout", "50ms", "--no-retry", "api", "GET", "/organizations"}
	if code := ExecuteArgs(arguments, strings.NewReader(""), io.Discard, io.Discard); code != 8 {
		t.Fatalf("network exit=%d, want 8", code)
	}
}

func TestStartupBudget(t *testing.T) {
	if runtime.GOOS == "darwin" && os.Getenv("LAGO_TEST_SIGNED_BINARY") != "1" {
		t.Skip("fresh unsigned macOS binaries incur a one-time platform validation; signed release artifacts are measured post-release")
	}
	binary := filepath.Join("..", "..", "bin", "lago")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if _, err := os.Stat(binary); err != nil {
		t.Skip("build bin/lago before running the cold-start budget")
	}
	started := time.Now()
	command := exec.Command(binary, "--help")
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("cold lago --help took %s, budget is 100ms", elapsed)
	}
}

func setCleanEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"LAGO_API_KEY", "LAGO_API_URL", "LAGO_MODE", "LAGO_PROFILE", "LAGO_TIMEOUT", "LAGO_DEBUG", "LAGO_UPDATE_API_BASE"} {
		t.Setenv(name, "")
	}
}
