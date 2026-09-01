package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/generated"
)

// jsonAPI serves the supplied handler over TLS with a JSON content type.
func jsonAPI(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		handler(response, request)
	}))
	t.Cleanup(server.Close)
	return server
}

// profileAt writes a test-mode profile pointing at serverURL and returns its path.
func profileAt(t *testing.T, serverURL string) string {
	t.Helper()
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	file := config.NewFile()
	file.CurrentProfile = "default"
	file.Profiles["default"] = config.Profile{
		Region: "self-hosted", APIKey: "lago_test_FAKE000000000000000000000000",
		APIURL: serverURL, Mode: config.ModeTest, Insecure: true,
	}
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAGO_CONFIG_FILE", path)
	return path
}

func execute(t *testing.T, stdin string, argv ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	app := NewApp(strings.NewReader(stdin), &stdout, &stderr, "test")
	root := NewRoot(app)
	root.SetArgs(argv)
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

// whoami is the first command an operator runs to confirm which environment they
// are pointed at. It must report the profile, mode, and organization together.
func TestWhoamiReportsProfileAndOrganization(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/organizations" {
			t.Errorf("whoami called %s", request.URL.Path)
		}
		_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1","name":"Example Organization"}}`))
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "--output", "json", "whoami")
	if err != nil {
		t.Fatalf("whoami failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("whoami output is not JSON: %q", stdout)
	}
	for _, field := range []string{"profile", "region", "mode", "api_url", "organization"} {
		if _, ok := payload[field]; !ok {
			t.Errorf("whoami output is missing %q: %s", field, stdout)
		}
	}
	if payload["mode"] != config.ModeTest {
		t.Errorf("mode = %v, want test", payload["mode"])
	}
}

// doctor must report every check it ran, exit non-zero when one fails, and still
// print the report so the operator can see which step broke.
func TestDoctorReportsChecksAndFailsLoudly(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
			response.Header().Set("X-Request-Id", "req_doctor")
			_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1"}}`))
		})
		profileAt(t, server.URL)

		stdout, _, err := execute(t, "", "--output", "json", "doctor")
		if err != nil {
			t.Fatalf("doctor failed against a healthy API: %v", err)
		}
		var report struct {
			OK     bool `json:"ok"`
			Checks []struct {
				Name string `json:"name"`
				OK   bool   `json:"ok"`
			} `json:"checks"`
		}
		if err := json.Unmarshal([]byte(stdout), &report); err != nil {
			t.Fatalf("doctor output is not JSON: %q", stdout)
		}
		if !report.OK || len(report.Checks) < 3 {
			t.Fatalf("doctor reported %+v", report)
		}
		names := map[string]bool{}
		for _, check := range report.Checks {
			names[check.Name] = check.OK
		}
		for _, required := range []string{"config_path", "configuration", "api"} {
			if !names[required] {
				t.Errorf("check %q missing or failed: %+v", required, report.Checks)
			}
		}
	})

	t.Run("unreachable API exits with the API error code", func(t *testing.T) {
		server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusUnauthorized)
			_, _ = response.Write([]byte(`{"status":401,"error":"Unauthorized"}`))
		})
		profileAt(t, server.URL)

		stdout, _, err := execute(t, "", "--output", "json", "doctor")
		if err == nil {
			t.Fatal("doctor succeeded against a rejecting API")
		}
		if apperr.ExitCode(err) != apperr.ExitAuth {
			t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitAuth)
		}
		if !strings.Contains(stdout, `"ok": false`) {
			t.Errorf("doctor did not print a failing report: %q", stdout)
		}
	})

	t.Run("no profile fails on configuration", func(t *testing.T) {
		setCleanEnvironment(t)
		t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "absent.toml"))
		stdout, _, err := execute(t, "", "--output", "json", "doctor")
		if err == nil {
			t.Fatal("doctor succeeded with no configuration")
		}
		if !strings.Contains(stdout, "configuration") {
			t.Errorf("doctor did not name the failing check: %q", stdout)
		}
	})
}

// The diagnostic bundle is what an operator emails to support. It must be written
// where asked, and must contain only the sanitized allowlist, never a credential.
func TestDoctorBundleContainsNoCredentials(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1"}}`))
	})
	profileAt(t, server.URL)
	bundlePath := filepath.Join(t.TempDir(), "diagnostics.tar.gz")

	if _, stderr, err := execute(t, "", "--output", "json", "doctor", "--bundle-path", bundlePath); err != nil {
		t.Fatalf("doctor --bundle-path failed: %v", err)
	} else if !strings.Contains(stderr, bundlePath) {
		t.Errorf("bundle path was not reported to the operator: %q", stderr)
	}

	contents, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("bundle was not written: %v", err)
	}
	for _, forbidden := range []string{"lago_test_FAKE", server.URL} {
		if bytes.Contains(contents, []byte(forbidden)) {
			t.Errorf("diagnostic bundle contains %q", forbidden)
		}
	}
}

// `lago docs` prints the documentation URL when not attached to a terminal, so it
// stays pipeable, and underscores must resolve the same as hyphens.
func TestDocsResolvesResourcesAndStaysPipeable(t *testing.T) {
	setCleanEnvironment(t)
	for _, resource := range []string{"customers", "billable_metrics", "billable-metrics"} {
		stdout, _, err := execute(t, "", "docs", resource)
		if err != nil {
			t.Fatalf("docs %s failed: %v", resource, err)
		}
		if !strings.HasPrefix(strings.TrimSpace(stdout), "https://") {
			t.Errorf("docs %s printed %q, want a URL", resource, stdout)
		}
	}

	if _, _, err := execute(t, "", "docs", "not-a-resource"); err == nil {
		t.Fatal("docs accepted an unknown resource")
	} else if apperr.ExitCode(err) != apperr.ExitNotFound {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitNotFound)
	}
}

func TestNormalizeCommandNameMapsUnderscores(t *testing.T) {
	t.Parallel()
	for input, want := range map[string]string{
		"billable_metrics": "billable-metrics",
		"credit_notes":     "credit-notes",
		"customers":        "customers",
		"":                 "",
	} {
		if got := normalizeCommandName(input); got != want {
			t.Errorf("normalizeCommandName(%q) = %q, want %q", input, got, want)
		}
	}
}

// Aliases are user-defined shortcuts. They must round-trip through the config and
// must never be able to shadow a real command.
func TestAliasLifecycle(t *testing.T) {
	path := profileAt(t, "https://api.example.test")

	if _, _, err := execute(t, "", "alias", "set", "recent", "invoices list --limit 5"); err != nil {
		t.Fatalf("alias set failed: %v", err)
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Aliases["recent"]) == 0 {
		t.Fatalf("alias was not persisted: %+v", saved.Aliases)
	}

	stdout, _, err := execute(t, "", "--output", "json", "alias", "list")
	if err != nil {
		t.Fatalf("alias list failed: %v", err)
	}
	if !strings.Contains(stdout, "recent") {
		t.Errorf("alias list did not include the new alias: %s", stdout)
	}

	if _, _, err := execute(t, "", "alias", "delete", "recent"); err != nil {
		t.Fatalf("alias delete failed: %v", err)
	}
	after, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := after.Aliases["recent"]; exists {
		t.Error("alias survived deletion")
	}
}

func TestAliasCannotShadowABuiltinCommand(t *testing.T) {
	profileAt(t, "https://api.example.test")
	for _, name := range []string{"customers", "doctor", "init", "version", "api", "whoami"} {
		if _, _, err := execute(t, "", "alias", "set", name, "invoices list"); err == nil {
			t.Errorf("alias %q was allowed to shadow a real command", name)
		}
	}
}

// Malformed alias definitions must be rejected rather than written to the config
// where they would break every later invocation.
func TestAliasRejectsMalformedDefinitions(t *testing.T) {
	profileAt(t, "https://api.example.test")
	for _, testCase := range []struct{ name, expansion string }{
		{"", "invoices list"},
		{"-flagish", "invoices list"},
		{"two words", "invoices list"},
		{"empty", "   "},
	} {
		if _, _, err := execute(t, "", "alias", "set", testCase.name, testCase.expansion); err == nil {
			t.Errorf("alias set %q %q was accepted", testCase.name, testCase.expansion)
		}
	}
	if _, _, err := execute(t, "", "alias", "delete", "never-defined"); err == nil {
		t.Error("deleting a missing alias succeeded")
	} else if apperr.ExitCode(err) != apperr.ExitNotFound {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitNotFound)
	}
}

// Completion scripts are consumed by shells; a malformed one breaks a user's
// terminal, so every advertised shell must produce non-empty output.
func TestCompletionScriptsGenerateForEveryShell(t *testing.T) {
	setCleanEnvironment(t)
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		stdout, _, err := execute(t, "", "completion", shell)
		if err != nil {
			t.Errorf("completion %s failed: %v", shell, err)
			continue
		}
		if len(strings.TrimSpace(stdout)) == 0 {
			t.Errorf("completion %s produced nothing", shell)
		}
	}
	if _, _, err := execute(t, "", "completion", "tcsh"); err == nil {
		t.Error("completion accepted an unsupported shell")
	}
}

// version must be machine-readable and carry the pinned spec identity, so a bug
// report names exactly which contract the binary was built against.
func TestVersionCarriesSpecIdentity(t *testing.T) {
	setCleanEnvironment(t)
	stdout, _, err := execute(t, "", "version", "--output", "json")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("version output is not JSON: %q", stdout)
	}
	for _, field := range []string{"version", "spec_version", "spec_sha256", "platform", "go_version"} {
		if value, _ := payload[field].(string); value == "" {
			t.Errorf("version output is missing %q: %s", field, stdout)
		}
	}
	if Version() == "" {
		t.Error("Version() returned an empty string")
	}

	plain, _, err := execute(t, "", "version")
	if err != nil || strings.TrimSpace(plain) == "" {
		t.Errorf("plain version output failed: %q %v", plain, err)
	}
}

// `lago docs` hands a URL from the pinned spec to the platform opener. xdg-open and
// rundll32 act on any scheme, so a spec-drift PR that changed one URL to a file:// path
// or a custom handler would become code execution on the reader's machine. The opener
// takes absolute https URLs only.
func TestBrowserOpenerRefusesNonHTTPSTargets(t *testing.T) {
	t.Parallel()
	for _, target := range []string{
		"",
		"not-a-url",
		"file:///etc/passwd",
		"javascript:alert(1)",
		"http://docs.example.test/customers",
		"https://",
		"/relative/path",
	} {
		if err := openBrowser(target); err == nil {
			t.Errorf("openBrowser accepted %q", target)
		}
	}
}

// Every documentation URL compiled into the binary must satisfy that rule, so the
// guardrail cannot be tripped by the spec the CLI actually ships with.
func TestEveryGeneratedDocsURLIsAbsoluteHTTPS(t *testing.T) {
	t.Parallel()
	seen := 0
	for _, operation := range generated.Operations {
		if operation.DocsURL == "" {
			continue
		}
		seen++
		parsed, err := url.Parse(operation.DocsURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			t.Errorf("%s has documentation URL %q, which `lago docs` would refuse to open", operation.Resource, operation.DocsURL)
		}
	}
	if seen == 0 {
		t.Fatal("no operation declares a documentation URL; `lago docs` has nothing to open")
	}
}
