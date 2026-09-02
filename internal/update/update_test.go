package update

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

func TestLatestStableAndBeta(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/releases/latest" {
			_, _ = io.WriteString(response, `{"tag_name":"v1.2.0","html_url":"https://example.test/release"}`)
			return
		}
		_, _ = io.WriteString(response, `[{"tag_name":"v1.3.0-beta.2","prerelease":true},{"tag_name":"v1.3.0-beta.1","prerelease":true}]`)
	}))
	defer server.Close()
	check, _, err := Latest(context.Background(), "1.1.0", "stable", "test", server.URL)
	if err != nil || !check.UpdateAvailable || check.Latest != "1.2.0" {
		t.Fatalf("check=%#v error=%v", check, err)
	}
	beta, _, err := Latest(context.Background(), "1.2.0", "beta", "test", server.URL)
	if err != nil || !beta.UpdateAvailable || beta.Latest != "1.3.0-beta.2" {
		t.Fatalf("beta=%#v error=%v", beta, err)
	}
}

// `lago upgrade` prints a command; it never replaces the running binary. Detect has to
// name the channel that owns a given path, because printing `brew upgrade` for a binary
// Homebrew does not manage produces a brew error rather than an upgrade.
func TestDetectClassifiesTheInstallingChannel(t *testing.T) {
	for _, testCase := range []struct {
		name string
		path string
		env  map[string]string
		want Method
	}{
		{name: "homebrew cellar", path: "/opt/homebrew/Cellar/lago/1.0.0/bin/lago", want: Homebrew},
		{name: "intel homebrew prefix", path: "/usr/local/Homebrew/bin/lago", want: Homebrew},
		{name: "linuxbrew", path: "/home/linuxbrew/.linuxbrew/Cellar/lago/1.0.0/bin/lago", want: Homebrew},
		{name: "default gopath bin", path: filepath.Join(mustHome(t), "go", "bin", "lago"), want: GoInstall},
		{name: "explicit GOBIN", path: "/srv/tools/lago", env: map[string]string{"GOBIN": "/srv/tools"}, want: GoInstall},
		{name: "explicit GOPATH", path: "/w/gopath/bin/lago", env: map[string]string{"GOPATH": "/w/gopath"}, want: GoInstall},
		{name: "manual install", path: "/usr/local/bin/lago", want: Unknown},
		{name: "windows manual install", path: `C:\\Program Files\\lago\\lago.exe`, want: Unknown},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("GOBIN", "")
			t.Setenv("GOPATH", "")
			for name, value := range testCase.env {
				t.Setenv(name, value)
			}
			if got := Detect(testCase.path); got != testCase.want {
				t.Errorf("Detect(%q) = %q, want %q", testCase.path, got, testCase.want)
			}
		})
	}
}

// Each recognised channel maps to exactly one command, and an unrecognised install
// yields no command so the caller knows to print both.
func TestUpgradeCommandNamesOneChannel(t *testing.T) {
	method, command, err := UpgradeCommand()
	if err != nil {
		t.Fatalf("UpgradeCommand failed: %v", err)
	}
	switch method {
	case Homebrew:
		if command != "brew upgrade getlago/tap/lago" {
			t.Errorf("homebrew command = %q", command)
		}
	case GoInstall:
		if command != "go install github.com/getlago/lago-cli/cmd/lago@latest" {
			t.Errorf("go install command = %q", command)
		}
	case Unknown:
		if command != "" {
			t.Errorf("unknown install returned a command: %q", command)
		}
	default:
		t.Fatalf("unexpected method %q", method)
	}
}

func mustHome(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory in this environment")
	}
	return home
}

// The release endpoint is GitHub, not Lago. Its failures must never surface as
// ExitServer, which the exit-code table documents as a Lago server 5xx: a script would
// conclude Lago is down when only the update check failed. Every non-200 is a network
// class error with a suggestion that names the likely cause.
func TestLatestClassifiesReleaseAPIFailuresAsNetworkErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		status  int
		mention string
	}{
		{http.StatusNotFound, "private"},
		{http.StatusForbidden, "rate-limited"},
		{http.StatusTooManyRequests, "rate-limited"},
		{http.StatusInternalServerError, "Retry later"},
	}
	for _, testCase := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(testCase.status)
		}))
		_, _, err := Latest(context.Background(), "1.0.0", "stable", "test", server.URL)
		server.Close()
		if err == nil {
			t.Fatalf("status %d: Latest succeeded", testCase.status)
		}
		var appErr *apperr.Error
		if !errors.As(err, &appErr) {
			t.Fatalf("status %d: error is not an apperr.Error: %v", testCase.status, err)
		}
		if appErr.ExitCode != apperr.ExitNetwork {
			t.Errorf("status %d: exit code = %d, want %d", testCase.status, appErr.ExitCode, apperr.ExitNetwork)
		}
		if appErr.Status != testCase.status {
			t.Errorf("status %d: recorded status = %d", testCase.status, appErr.Status)
		}
		if !strings.Contains(appErr.Message, "GitHub") {
			t.Errorf("status %d: message does not name GitHub: %q", testCase.status, appErr.Message)
		}
		if !strings.Contains(appErr.Suggestion, testCase.mention) || !strings.Contains(appErr.Suggestion, "go install") {
			t.Errorf("status %d: suggestion misses %q or the upgrade command: %q", testCase.status, testCase.mention, appErr.Suggestion)
		}
	}
}

// A version that is not a semver tag came from `make build`, `go install` of a commit,
// or a local VERSION override, never from a release. Such a binary has nothing on
// GitHub to compare itself with.
func TestIsDevelopmentRecognizesSourceBuilds(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"dev", "qa-local", "9e0eefe", "", "unknown"} {
		if !IsDevelopment(version) {
			t.Errorf("IsDevelopment(%q) = false, want true", version)
		}
	}
	for _, version := range []string{"1.0.0", "v1.0.0", "1.2.3-beta.1", "v0.1.0"} {
		if IsDevelopment(version) {
			t.Errorf("IsDevelopment(%q) = true, want false", version)
		}
	}
}
