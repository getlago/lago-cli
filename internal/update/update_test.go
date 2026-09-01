package update

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
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
