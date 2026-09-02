package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// releaseAPI serves a GitHub-shaped release-metadata endpoint.
func releaseAPI(t *testing.T, latest string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"tag_name":"v`+latest+`","html_url":"https://example.test/release"}`)
	}))
	t.Cleanup(server.Close)
	return server
}

// `lago upgrade` no longer replaces the running binary: with only Homebrew and
// `go install` supported, no install is one the CLI itself owns. It must print the
// command for the channel that installed it, and never claim to have upgraded anything.
func TestUpgradePrintsACommandAndNeverSelfInstalls(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	t.Setenv("LAGO_UPDATE_API_BASE", releaseAPI(t, "9.9.9").URL)

	stdout, err := executeUpgradeAs(t, "1.0.0")
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if !strings.Contains(stdout, "9.9.9") {
		t.Errorf("upgrade did not report the available release: %q", stdout)
	}
	// The test binary is not brew- or go-install-managed, so both commands are printed.
	for _, want := range []string{"brew upgrade getlago/tap/lago", "go install github.com/getlago/lago-cli/cmd/lago@latest"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("upgrade did not print %q:\n%s", want, stdout)
		}
	}
	for _, forbidden := range []string{"Upgraded Lago CLI", "install.sh", "scoop", "winget", "Scoop", "Winget"} {
		if strings.Contains(stdout, forbidden) {
			t.Errorf("upgrade output references the removed self-update path or a parked channel (%q):\n%s", forbidden, stdout)
		}
	}
}

// A binary already on the latest release must say so and print no command at all.
func TestUpgradeSaysNothingToDoWhenCurrent(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	t.Setenv("LAGO_UPDATE_API_BASE", releaseAPI(t, "1.0.0").URL)

	var stdout strings.Builder
	app := NewApp(strings.NewReader(""), &stdout, io.Discard, "1.0.0")
	root := NewRoot(app)
	root.SetArgs([]string{"upgrade"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	if err := root.Execute(); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "already current") {
		t.Errorf("upgrade did not report a current install: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "brew upgrade") {
		t.Errorf("upgrade printed a command for an install that is already current: %q", stdout.String())
	}
}

// The version check itself must still work and must honour the channel flag.
func TestVersionCheckUsesTheSelectedChannel(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	t.Setenv("LAGO_UPDATE_API_BASE", releaseAPI(t, "2.0.0").URL)

	stdout, _, err := execute(t, "", "--output", "json", "version", "--check")
	if err != nil {
		t.Fatalf("version --check failed: %v", err)
	}
	if !strings.Contains(stdout, `"latest": "2.0.0"`) {
		t.Errorf("version --check did not report the latest release: %s", stdout)
	}
}

// executeUpgradeAs runs `lago upgrade` for a binary that reports the given version.
func executeUpgradeAs(t *testing.T, version string) (string, error) {
	t.Helper()
	var stdout strings.Builder
	app := NewApp(strings.NewReader(""), &stdout, io.Discard, version)
	root := NewRoot(app)
	root.SetArgs([]string{"upgrade"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	err := root.Execute()
	return stdout.String(), err
}

// A source build (`dev`, a local VERSION override, a commit hash) was not installed
// from any release, so there is nothing on GitHub to compare it with. `upgrade` must
// say so and print the rebuild command without touching the network: on a private
// repository or behind a filtering proxy the network call could only produce a
// misleading error.
func TestUpgradeOnADevelopmentBuildNeverCallsGitHub(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	t.Cleanup(server.Close)
	t.Setenv("LAGO_UPDATE_API_BASE", server.URL)

	for _, version := range []string{"dev", "qa-local", "9e0eefe"} {
		stdout, err := executeUpgradeAs(t, version)
		if err != nil {
			t.Fatalf("upgrade on %q failed: %v", version, err)
		}
		if !strings.Contains(stdout, "development build") || !strings.Contains(stdout, "go install github.com/getlago/lago-cli/cmd/lago@latest") {
			t.Errorf("upgrade on %q did not explain the source build:\n%s", version, stdout)
		}
		if strings.Contains(stdout, "brew upgrade") {
			t.Errorf("upgrade on %q suggested Homebrew for a source build:\n%s", version, stdout)
		}
	}
	if called {
		t.Error("upgrade on a development build contacted the release API")
	}
}

// GitHub is not Lago. When the release endpoint answers 404 (private repository, no
// release yet) or 403 (proxy, rate limit), the failure must be reported as a network
// class error, exit 8, never as exit 7 which the exit-code table defines as a Lago
// server 5xx. A script reading 7 would conclude Lago is down.
func TestUpgradeReportsGitHubFailuresAsNetworkErrors(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
		setCleanEnvironment(t)
		t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml"))
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(status)
		}))
		t.Cleanup(server.Close)
		t.Setenv("LAGO_UPDATE_API_BASE", server.URL)

		_, err := executeUpgradeAs(t, "1.0.0")
		if err == nil {
			t.Fatalf("upgrade succeeded against a %d release API", status)
		}
		if code := apperr.ExitCode(err); code != apperr.ExitNetwork {
			t.Errorf("status %d: exit code = %d, want %d (network); got error %v", status, code, apperr.ExitNetwork, err)
		}
		if !strings.Contains(err.Error(), "GitHub") {
			t.Errorf("status %d: error does not name GitHub: %v", status, err)
		}
	}
}
