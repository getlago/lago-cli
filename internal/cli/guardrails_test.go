package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
)

// writeProfile installs a single-profile config in the given mode pointing at server.
func writeProfile(t *testing.T, mode, serverURL string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	file := config.File{
		Version:        1,
		CurrentProfile: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Region:   "self-hosted",
				APIKey:   "lago_" + mode + "_FAKEabcdefghijklmnopqrstuv",
				APIURL:   serverURL,
				Mode:     mode,
				Insecure: true,
			},
		},
	}
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAGO_CONFIG_FILE", path)
	return path
}

// runCommand executes argv against a fresh root with the supplied stdin.
func runCommand(t *testing.T, stdin string, argv ...string) (string, string, error) {
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

// A destructive command with no TTY and no --confirm must refuse, not proceed on EOF.
// This is the guardrail that keeps a CI script from deleting a live customer.
func TestDestructiveCommandRefusesWithoutTTYOrConfirm(t *testing.T) {
	setCleanEnvironment(t)
	var reached bool
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached = true
	}))
	defer server.Close()
	writeProfile(t, config.ModeLive, server.URL)

	_, _, err := runCommand(t, "", "customers", "delete", "cus_live_1")
	if err == nil {
		t.Fatal("destructive live command proceeded with no confirmation")
	}
	if reached {
		t.Fatal("destructive live command reached the API before confirming")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	if !strings.Contains(err.Error(), "cus_live_1") {
		t.Errorf("refusal does not name the resource: %v", err)
	}
}

// In live mode the confirmation must be the exact identifier. "y" is not enough.
func TestLiveConfirmationRequiresTypingTheIdentifier(t *testing.T) {
	setCleanEnvironment(t)
	var reached bool
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached = true
	}))
	defer server.Close()
	writeProfile(t, config.ModeLive, server.URL)

	_, _, err := runCommand(t, "", "customers", "delete", "cus_live_1", "--confirm", "y")
	if err == nil || reached {
		t.Fatal("live delete accepted a confirmation that was not the identifier")
	}
}

// A wrong --confirm value must never fall through to the API.
func TestMismatchedConfirmNeverReachesTheAPI(t *testing.T) {
	setCleanEnvironment(t)
	var reached bool
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached = true
	}))
	defer server.Close()
	writeProfile(t, config.ModeTest, server.URL)

	_, _, err := runCommand(t, "", "customers", "delete", "cus_1", "--confirm", "cus_2")
	if err == nil || reached {
		t.Fatal("delete proceeded with a mismatched --confirm")
	}
}

// Money-moving operations that are not DELETE must be gated too. finalize, void and
// retry-payment are the operations that shipped ungated before this guardrail existed.
func TestMoneyMovingOperationsAreConfirmationGated(t *testing.T) {
	setCleanEnvironment(t)
	for _, argv := range [][]string{
		{"invoices", "finalize", "inv_1"},
		{"invoices", "void", "inv_1"},
		{"invoices", "retry-payment", "inv_1"},
		{"credit-notes", "void", "cn_1"},
	} {
		t.Run(strings.Join(argv[:2], "-"), func(t *testing.T) {
			var reached bool
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				reached = true
				response.Header().Set("Content-Type", "application/json")
				_, _ = response.Write([]byte(`{}`))
			}))
			defer server.Close()
			writeProfile(t, config.ModeLive, server.URL)

			if _, _, err := runCommand(t, "", argv...); err == nil {
				t.Errorf("%v proceeded in live mode with no confirmation", argv)
			}
			if reached {
				t.Errorf("%v reached the API before confirming", argv)
			}
		})
	}
}

// The [LIVE] banner must appear on every live request so operators cannot mistake
// which environment a script is pointed at.
func TestLiveModeBannerIsPrintedToStderr(t *testing.T) {
	setCleanEnvironment(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"customers":[]}`))
	}))
	defer server.Close()
	writeProfile(t, config.ModeLive, server.URL)

	_, stderr, err := runCommand(t, "", "customers", "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(stderr, "[LIVE]") {
		t.Errorf("live request printed no [LIVE] banner: %q", stderr)
	}
}

// A test-mode profile must never print the live banner, and must reach only the
// configured test host.
func TestTestModeProfileNeverSignalsLive(t *testing.T) {
	setCleanEnvironment(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"customers":[]}`))
	}))
	defer server.Close()
	writeProfile(t, config.ModeTest, server.URL)

	_, stderr, err := runCommand(t, "", "customers", "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if strings.Contains(stderr, "[LIVE]") {
		t.Errorf("test-mode profile printed the live banner: %q", stderr)
	}
}

// promptConfirmation is the confirmation policy itself. Live mode must accept only the
// exact identifier; test mode accepts y/N; both refuse on empty input (EOF).
func TestPromptConfirmationPolicy(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name       string
		input      string
		live       bool
		wantAccept bool
	}{
		{"live exact identifier", "cus_1\n", true, true},
		{"live wrong identifier", "cus_2\n", true, false},
		{"live yes is not enough", "y\n", true, false},
		{"live empty input", "", true, false},
		{"live whitespace tolerated", "  cus_1  \n", true, true},
		{"test yes", "y\n", false, true},
		{"test uppercase yes", "Y\n", false, true},
		{"test no", "n\n", false, false},
		{"test empty defaults to no", "", false, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var prompt bytes.Buffer
			err := promptConfirmation(strings.NewReader(testCase.input), &prompt, "cus_1", testCase.live)
			if testCase.wantAccept && err != nil {
				t.Fatalf("confirmation rejected valid input %q: %v", testCase.input, err)
			}
			if !testCase.wantAccept && err == nil {
				t.Fatalf("confirmation accepted %q", testCase.input)
			}
			if testCase.live && !strings.Contains(prompt.String(), "[LIVE]") {
				t.Errorf("live prompt lacks the [LIVE] marker: %q", prompt.String())
			}
			if !testCase.live && strings.Contains(prompt.String(), "[LIVE]") {
				t.Errorf("test prompt claims live: %q", prompt.String())
			}
		})
	}
}
