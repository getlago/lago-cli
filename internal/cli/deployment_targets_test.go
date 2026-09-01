package cli

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/config"
)

// readEnv is a thin wrapper so the target selector reads clearly at its call site.
func readEnv(name string) string { return os.Getenv(name) }

// deploymentTargets are the three shapes every fix has to work on. Each is exercised at
// its real hostname, not at 127.0.0.1, so a fix that quietly assumed the cloud US shape
// cannot pass. The self-hosted entry carries a custom port on purpose.
//
// CI runs this file with LAGO_TEST_TARGET set to each name in turn as well as unset.
var deploymentTargets = []struct {
	name   string
	region string
	// base is what an operator configures, and resolved is the base URL the client must
	// end up calling. Both spellings of base -- with and without the API prefix -- have
	// to reach the same resolved URL.
	base     string
	pasted   string
	resolved string
	host     string
}{
	{
		name: "us", region: "us",
		base: "https://api.getlago.com", pasted: "https://api.getlago.com/api/v1",
		resolved: "https://api.getlago.com/api/v1", host: "api.getlago.com",
	},
	{
		name: "eu", region: "eu",
		base: "https://api.eu.getlago.com", pasted: "https://api.eu.getlago.com/api/v1/",
		resolved: "https://api.eu.getlago.com/api/v1", host: "api.eu.getlago.com",
	},
	{
		name: "self-hosted", region: "self-hosted",
		base: "https://lago.acme.test:8443", pasted: "https://lago.acme.test:8443/api/v1",
		resolved: "https://lago.acme.test:8443/api/v1", host: "lago.acme.test:8443",
	},
	{
		name: "self-hosted-subpath", region: "self-hosted",
		base: "https://tools.acme.com/lago", pasted: "https://tools.acme.com/lago/api/v1",
		resolved: "https://tools.acme.com/lago/api/v1", host: "tools.acme.com",
	},
}

// selectedTargets honours LAGO_TEST_TARGET so CI can run the same tests once per
// deployment shape. Unset means all of them.
func selectedTargets(t *testing.T) []int {
	t.Helper()
	wanted := strings.TrimSpace(strings.ToLower(readEnv("LAGO_TEST_TARGET")))
	indices := make([]int, 0, len(deploymentTargets))
	for index, target := range deploymentTargets {
		if wanted == "" || target.name == wanted || strings.HasPrefix(target.name, wanted+"-") {
			indices = append(indices, index)
		}
	}
	if len(indices) == 0 {
		t.Fatalf("LAGO_TEST_TARGET=%q matches no deployment target", wanted)
	}
	return indices
}

// localDialer routes every connection to addr, so a test can configure a production
// hostname and still be served locally. It is the only reason transport.Config exposes
// DialContext.
func localDialer(addr string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}
}

// runAt executes argv against server as if server were the deployment at apiURL.
func runAt(t *testing.T, server *httptest.Server, apiURL string, argv ...string) (string, string, error) {
	t.Helper()
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	var stdout, stderr strings.Builder
	app := NewApp(strings.NewReader(""), &stdout, &stderr, "test")
	app.dialContext = localDialer(server.Listener.Addr().String())
	root := NewRoot(app)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs(append([]string{
		"--api-url", apiURL,
		"--api-key", "lago_test_FAKE000000000000000000000000",
		"--mode", "test", "--insecure",
	}, argv...))
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

// A pasted /api/v1 must not produce a /api/v1/api/v1 request, and the base and pasted
// spellings must reach byte-identical URLs. This is the QA finding, checked on every
// deployment target rather than only on cloud US.
func TestPastedAPIPathReachesTheSameURLOnEveryTarget(t *testing.T) {
	for _, index := range selectedTargets(t) {
		target := deploymentTargets[index]
		t.Run(target.name, func(t *testing.T) {
			var paths []string
			var hosts []string
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				paths = append(paths, request.URL.Path)
				hosts = append(hosts, request.Host)
				response.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(response, `{"customers":[],"meta":{"total_pages":1}}`)
			}))
			defer server.Close()

			for _, configured := range []string{target.base, target.pasted} {
				if _, _, err := runAt(t, server, configured, "customers", "list"); err != nil {
					t.Fatalf("%s configured as %q failed: %v", target.name, configured, err)
				}
			}
			if len(paths) != 2 {
				t.Fatalf("expected two requests, got %d: %v", len(paths), paths)
			}
			if paths[0] != paths[1] {
				t.Errorf("base %q reached %q but pasted %q reached %q", target.base, paths[0], target.pasted, paths[1])
			}
			for _, path := range paths {
				if strings.Count(path, "/api/v1") != 1 {
					t.Errorf("%s produced a request path with a doubled prefix: %q", target.name, path)
				}
				if !strings.HasSuffix(path, "/api/v1/customers") {
					t.Errorf("%s reached %q, want it to end in /api/v1/customers", target.name, path)
				}
			}
			for _, host := range hosts {
				if host != target.host {
					t.Errorf("%s sent Host %q, want %q", target.name, host, target.host)
				}
			}
		})
	}
}

// A request path pasted with its /api/v1 prefix, through the raw `api` escape hatch,
// must land on the same endpoint as the relative form.
func TestRawAPICommandTolerantOfAPastedPrefix(t *testing.T) {
	for _, index := range selectedTargets(t) {
		target := deploymentTargets[index]
		t.Run(target.name, func(t *testing.T) {
			var paths []string
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				paths = append(paths, request.URL.Path)
				response.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(response, `{"customers":[]}`)
			}))
			defer server.Close()

			for _, path := range []string{"/customers", "/api/v1/customers"} {
				if _, _, err := runAt(t, server, target.base, "api", "GET", path); err != nil {
					t.Fatalf("api GET %s failed: %v", path, err)
				}
			}
			if len(paths) != 2 || paths[0] != paths[1] {
				t.Errorf("relative and prefixed paths diverged: %v", paths)
			}
		})
	}
}

// whoami and doctor must both report the host requests actually went to, on every
// deployment. An operator debugging "which environment am I on" reads this line.
func TestWhoamiAndDoctorReportTheResolvedHost(t *testing.T) {
	for _, index := range selectedTargets(t) {
		target := deploymentTargets[index]
		t.Run(target.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				response.Header().Set("Content-Type", "application/json")
				response.Header().Set("X-Request-Id", "req_target")
				_, _ = io.WriteString(response, `{"organization":{"lago_id":"org_1","name":"Example Organization"}}`)
			}))
			defer server.Close()

			// Configured with the base form and with the pasted form, whoami must report
			// the same resolved URL both times.
			for _, configured := range []string{target.base, target.pasted} {
				stdout, _, err := runAt(t, server, configured, "--output", "json", "whoami")
				if err != nil {
					t.Fatalf("whoami on %s failed: %v", target.name, err)
				}
				var payload map[string]any
				if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
					t.Fatalf("whoami output is not JSON: %q", stdout)
				}
				if payload["resolved_api_url"] != target.resolved {
					t.Errorf("whoami (configured %q) reported resolved_api_url %v, want %q",
						configured, payload["resolved_api_url"], target.resolved)
				}
			}

			stdout, _, err := runAt(t, server, target.pasted, "--output", "json", "doctor")
			if err != nil {
				t.Fatalf("doctor on %s failed: %v", target.name, err)
			}
			var report struct {
				Checks []struct {
					Name   string `json:"name"`
					OK     bool   `json:"ok"`
					Detail string `json:"detail"`
				} `json:"checks"`
			}
			if err := json.Unmarshal([]byte(stdout), &report); err != nil {
				t.Fatalf("doctor output is not JSON: %q", stdout)
			}
			found := false
			for _, check := range report.Checks {
				if check.Name != "api_url" {
					continue
				}
				found = true
				if check.Detail != target.resolved {
					t.Errorf("doctor reported api_url %q, want %q", check.Detail, target.resolved)
				}
			}
			if !found {
				t.Errorf("doctor did not report the resolved API URL: %+v", report.Checks)
			}
		})
	}
}

// The region shorthand must configure exactly the same host as the explicit base URL,
// or `--region eu` and `--api-url https://api.eu.getlago.com` are two deployments that
// look like one.
func TestRegionShorthandAndExplicitURLSaveTheSameProfile(t *testing.T) {
	for _, testCase := range []struct{ region, want string }{
		{"us", "https://api.getlago.com/api/v1"},
		{"eu", "https://api.eu.getlago.com/api/v1"},
	} {
		t.Run(testCase.region, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				response.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(response, `{"organization":{"lago_id":"org_1","name":"Example Organization"}}`)
			}))
			defer server.Close()

			for _, argv := range [][]string{
				{"init", "--region", testCase.region},
				{"init", "--region", testCase.region, "--api-url", testCase.want},
			} {
				setCleanEnvironment(t)
				path := filepath.Join(t.TempDir(), "config.toml")
				t.Setenv("LAGO_CONFIG_FILE", path)
				app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
				app.dialContext = localDialer(server.Listener.Addr().String())
				root := NewRoot(app)
				root.SetOut(io.Discard)
				root.SetErr(io.Discard)
				root.SetArgs(append([]string{
					"--api-key", "lago_test_FAKE000000000000000000000000", "--mode", "test", "--insecure",
				}, argv...))
				if err := root.Execute(); err != nil {
					t.Fatalf("%v failed: %v", argv, err)
				}
				saved := loadProfile(t, path)
				if saved != testCase.want {
					t.Errorf("%v saved api_url %q, want %q", argv, saved, testCase.want)
				}
			}
		})
	}
}

// init must normalize before saving. A pasted /api/v1 that went into the config file
// raw is how whoami ended up reporting a URL that was not the one being called.
func TestInitSavesTheNormalizedURL(t *testing.T) {
	for _, index := range selectedTargets(t) {
		target := deploymentTargets[index]
		t.Run(target.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if !strings.HasSuffix(request.URL.Path, "/api/v1/organizations") {
					t.Errorf("init validated against %q", request.URL.Path)
				}
				response.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(response, `{"organization":{"lago_id":"org_1","name":"Example Organization"}}`)
			}))
			defer server.Close()

			for _, configured := range []string{target.base, target.pasted} {
				setCleanEnvironment(t)
				path := filepath.Join(t.TempDir(), "config.toml")
				t.Setenv("LAGO_CONFIG_FILE", path)
				app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
				app.dialContext = localDialer(server.Listener.Addr().String())
				root := NewRoot(app)
				root.SetOut(io.Discard)
				root.SetErr(io.Discard)
				root.SetArgs([]string{
					"--api-key", "lago_test_FAKE000000000000000000000000", "--mode", "test", "--insecure",
					"init", "--region", "self-hosted", "--api-url", configured,
				})
				if err := root.Execute(); err != nil {
					t.Fatalf("init with %q failed: %v", configured, err)
				}
				if saved := loadProfile(t, path); saved != target.resolved {
					t.Errorf("init with %q saved %q, want %q", configured, saved, target.resolved)
				}
			}
		})
	}
}

// A region and an explicit URL that name different deployments is an ambiguity, not a
// precedence question. Silently preferring one is how somebody writes to the wrong
// continent. Agreeing after normalization is merely redundant and must be accepted.
func TestInitRefusesAConflictingRegionAndURL(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"organization":{"lago_id":"org_1"}}`)
	}))
	defer server.Close()

	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "config.toml"))
	app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
	app.dialContext = localDialer(server.Listener.Addr().String())
	root := NewRoot(app)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{
		"--api-key", "lago_test_FAKE000000000000000000000000", "--mode", "test", "--insecure",
		"init", "--region", "us", "--api-url", "https://api.eu.getlago.com",
	})
	err := root.Execute()
	if err == nil {
		t.Fatal("init accepted --region us together with the EU API URL")
	}
	if !strings.Contains(err.Error(), "different deployments") {
		t.Errorf("conflict error does not explain itself: %v", err)
	}
}

// Pasting the dashboard URL must fail by name on every deployment, and must never reach
// the network: an API key sent to the frontend is a credential in the wrong place.
func TestAppURLNeverReachesTheNetwork(t *testing.T) {
	for _, testCase := range []struct{ pasted, wantHost string }{
		{"https://app.getlago.com", "api.getlago.com"},
		{"https://app.eu.getlago.com", "api.eu.getlago.com"},
		{"https://app.lago.acme.test", "api.lago.acme.test"},
	} {
		t.Run(testCase.pasted, func(t *testing.T) {
			reached := false
			server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				reached = true
			}))
			defer server.Close()

			_, _, err := runAt(t, server, testCase.pasted, "whoami")
			if err == nil {
				t.Fatalf("%s was accepted as an API URL", testCase.pasted)
			}
			if reached {
				t.Fatalf("%s sent a request before being refused", testCase.pasted)
			}
			if !strings.Contains(err.Error(), "dashboard") {
				t.Errorf("error does not say it is the dashboard: %v", err)
			}
		})
	}
}

// loadProfile returns the api_url the current profile was saved with.
func loadProfile(t *testing.T, path string) string {
	t.Helper()
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	name := loaded.CurrentProfile
	if name == "" {
		name = "default"
	}
	return loaded.Profiles[name].APIURL
}
