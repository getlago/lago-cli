package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
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

	stdout, _, err := execute(t, "", "upgrade")
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
