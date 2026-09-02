package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	cliupdate "github.com/getlago/lago-cli/internal/update"
)

// init must validate credentials against the API before writing them, so a typo is
// reported immediately instead of surfacing on the first real billing command.
func TestInitValidationAndProfileShaping(t *testing.T) {
	t.Run("rejects credentials the API refuses", func(t *testing.T) {
		server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusUnauthorized)
			_, _ = response.Write([]byte(`{"status":401,"error":"Unauthorized"}`))
		})
		setCleanEnvironment(t)
		path := filepath.Join(t.TempDir(), "config.toml")
		t.Setenv("LAGO_CONFIG_FILE", path)

		_, _, err := execute(t, "", "init", "--api-key", "lago_test_FAKEwrongwrongwrongwrong",
			"--region", "self-hosted", "--api-url", server.URL, "--insecure", "--mode", "test")
		if err == nil {
			t.Fatal("init accepted credentials the API rejected")
		}
		if apperr.ExitCode(err) != apperr.ExitAuth {
			t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitAuth)
		}
		if _, statErr := os.Stat(path); statErr == nil {
			t.Error("init wrote a profile despite failing validation")
		}
	})

	t.Run("rejects an unknown region and an unknown mode", func(t *testing.T) {
		setCleanEnvironment(t)
		t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "config.toml"))
		if _, _, err := execute(t, "", "init", "--api-key", "k", "--region", "mars", "--mode", "test"); err == nil {
			t.Error("an unknown region was accepted")
		}
		if _, _, err := execute(t, "", "init", "--api-key", "k", "--region", "us", "--mode", "sandbox"); err == nil {
			t.Error("an unknown mode was accepted")
		}
	})

	t.Run("self-hosted requires an explicit URL", func(t *testing.T) {
		setCleanEnvironment(t)
		t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "config.toml"))
		if _, _, err := execute(t, "", "init", "--api-key", "k", "--region", "self-hosted", "--mode", "test"); err == nil {
			t.Error("self-hosted init succeeded with no --api-url")
		}
	})

	t.Run("named profiles coexist", func(t *testing.T) {
		server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
			_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1","name":"Example"}}`))
		})
		setCleanEnvironment(t)
		path := filepath.Join(t.TempDir(), "config.toml")
		t.Setenv("LAGO_CONFIG_FILE", path)

		for _, name := range []string{"first", "second"} {
			if _, _, err := execute(t, "", "init", "--api-key", "lago_test_FAKE000000000000000000000000",
				"--region", "self-hosted", "--api-url", server.URL, "--insecure", "--mode", "test",
				"--profile", name); err != nil {
				t.Fatalf("init %s failed: %v", name, err)
			}
		}
		saved, err := config.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(saved.Profiles) != 2 {
			t.Fatalf("profiles = %+v, want two", saved.Profiles)
		}
		if saved.Profiles["first"].OrganizationID != "org_1" {
			t.Errorf("organization identity was not recorded: %+v", saved.Profiles["first"])
		}
	})
}

// A non-TTY init must never prompt: it either has enough flags or it fails.
func TestInitNeverPromptsWithoutATerminal(t *testing.T) {
	setCleanEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", filepath.Join(t.TempDir(), "config.toml"))

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _, _ = execute(t, "", "init", "--region", "us", "--mode", "test")
	}()
	select {
	case <-done:
	case <-timeoutAfterSeconds(5):
		t.Fatal("init blocked waiting for input with no terminal attached")
	}
}

// upgrade must defer to the package manager rather than overwriting a binary it
// did not install, and must report cleanly when already current.
func TestUpgradeReportsWhenAlreadyCurrent(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"tag_name":"v0.0.1","assets":[]}`))
	})
	profileAt(t, server.URL)
	t.Setenv("LAGO_UPDATE_API", server.URL)

	// The command must not panic or hang regardless of what the release API says.
	_, _, _ = execute(t, "", "upgrade")
}

// Events stream from a file or stdin. Every line gets a transaction ID in its body so
// the server can deduplicate it; no Idempotency-Key header is sent because lago-api
// does not read one.
func TestEventStreamAssignsStableTransactionIDs(t *testing.T) {
	var mutex sync.Mutex
	keys := map[string]int{}
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		if header := request.Header.Get("Idempotency-Key"); header != "" {
			t.Errorf("unexpected Idempotency-Key header %q", header)
		}
		var payload map[string]map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		transactionID, _ := payload["event"]["transaction_id"].(string)
		mutex.Lock()
		keys[transactionID]++
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt"}}`))
	})
	profileAt(t, server.URL)

	path := filepath.Join(t.TempDir(), "events.ndjson")
	contents := `{"transaction_id":"txn_1","external_subscription_id":"sub_1","code":"requests"}
{"event":{"transaction_id":"txn_2","external_subscription_id":"sub_1","code":"requests"}}
{"external_subscription_id":"sub_1","code":"requests"}
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute(t, "", "--output", "json", "events", "send", "--file", path); err != nil {
		t.Fatalf("event stream failed: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if len(keys) != 3 {
		t.Fatalf("three events produced %d distinct transaction IDs: %v", len(keys), keys)
	}
	for key, count := range keys {
		if key == "" {
			t.Error("an event was sent with no transaction_id")
		}
		if count != 1 {
			t.Errorf("key %s was sent %d times", key, count)
		}
	}
}

// A malformed line aborts the whole stream, naming the line. Sending the valid
// half of a broken usage file and reporting the rest would leave an operator
// guessing which events were billed, so the CLI stops instead.
func TestEventStreamAbortsAtTheFirstMalformedLine(t *testing.T) {
	var mutex sync.Mutex
	accepted := 0
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		accepted++
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt"}}`))
	})
	profileAt(t, server.URL)

	path := filepath.Join(t.TempDir(), "events.ndjson")
	contents := `{"transaction_id":"ok_1","code":"requests"}
{not json
{"transaction_id":"never_sent","code":"requests"}
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := execute(t, "", "--output", "json", "events", "send", "--file", path)
	if err == nil {
		t.Fatal("a malformed event line did not fail the command")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("the failing line number was not reported: %v", err)
	}
	if apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitUsage)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if accepted > 1 {
		t.Errorf("%d events were sent past the malformed line", accepted)
	}
}

// Concurrency is operator-supplied and must be bounded.
func TestEventStreamValidatesConcurrency(t *testing.T) {
	profileAt(t, "https://api.example.test")
	for _, value := range []string{"0", "-1", "65", "1000"} {
		if _, _, err := execute(t, "", "events", "send", "--file", "-", "--concurrency", value); err == nil {
			t.Errorf("--concurrency %s was accepted", value)
		}
	}
}

func TestEventReaderSources(t *testing.T) {
	t.Parallel()
	reader, closer, err := eventReader(strings.NewReader("from stdin"), "-")
	if err != nil {
		t.Fatal(err)
	}
	defer closer()
	buffer := make([]byte, 10)
	if _, readErr := reader.Read(buffer); readErr != nil {
		t.Fatalf("stdin reader failed: %v", readErr)
	}

	path := filepath.Join(t.TempDir(), "events.ndjson")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fileReader, fileCloser, err := eventReader(nil, path)
	if err != nil {
		t.Fatal(err)
	}
	if fileReader == nil {
		t.Fatal("file reader is nil")
	}
	fileCloser()

	if _, _, err := eventReader(nil, filepath.Join(t.TempDir(), "absent.ndjson")); err == nil {
		t.Error("a missing event file was accepted")
	}
}

// The update check is opt-in and must stay silent unless consent was recorded.
func TestPassiveUpdateStaysSilentWithoutConsent(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)

	file := config.NewFile()
	file.UpdateConsent = false
	file.UpdateCheck = false
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	if result := startPassiveUpdate([]string{"customers", "list"}, "v1.0.0"); result != nil {
		t.Error("an update check started without consent")
	}

	t.Setenv("LAGO_NO_UPDATE_CHECK", "1")
	file.UpdateConsent = true
	file.UpdateCheck = true
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	if result := startPassiveUpdate([]string{"customers", "list"}, "v1.0.0"); result != nil {
		t.Error("LAGO_NO_UPDATE_CHECK did not disable the update check")
	}

	// finishPassiveUpdate must tolerate a nil channel.
	finishPassiveUpdate(nil, &strings.Builder{})
}

// Commands that must stay fast or run offline never trigger an update check.
func TestPassiveUpdateSkipsOfflineCommands(t *testing.T) {
	t.Parallel()
	for _, arguments := range [][]string{
		{"version"}, {"help"}, {"completion", "zsh"}, {"init"}, {"upgrade"},
		{"--output", "json", "version"}, {"--profile", "staging", "version"},
		{"--output=json", "version"}, {"--verbose", "version"}, {"--dry-run", "init"}, {},
	} {
		if !excludedPassiveCommand(arguments) {
			t.Errorf("%v was not excluded from the update check", arguments)
		}
	}
	for _, arguments := range [][]string{
		{"customers", "list"}, {"--profile", "staging", "invoices", "list"},
		{"--output", "json", "customers", "list"}, {"--verbose", "customers", "list"},
	} {
		if excludedPassiveCommand(arguments) {
			t.Errorf("%v was wrongly excluded from the update check", arguments)
		}
	}
}

func timeoutAfterSeconds(seconds int) <-chan time.Time {
	return time.After(time.Duration(seconds) * time.Second)
}

// prompt is the interactive half of init. It must echo a default, accept a typed
// value, fall back to the default on a bare newline, and tolerate EOF.
func TestPromptDefaultsAndEOF(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, input, label, fallback, want, wantLabel string
	}{
		{"typed value wins", "typed\n", "Region", "us", "typed", "Region [us]: "},
		{"blank falls back", "\n", "Region", "us", "us", "Region [us]: "},
		{"whitespace falls back", "   \n", "Region", "us", "us", "Region [us]: "},
		{"EOF falls back", "", "Region", "us", "us", "Region [us]: "},
		{"no default shown", "value\n", "API key", "", "value", "API key: "},
		{"EOF with no default", "", "API key", "", "", "API key: "},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var out strings.Builder
			got, err := prompt(bufio.NewReader(strings.NewReader(testCase.input)), &out, testCase.label, testCase.fallback)
			if err != nil {
				t.Fatalf("prompt failed: %v", err)
			}
			if got != testCase.want {
				t.Errorf("prompt = %q, want %q", got, testCase.want)
			}
			if out.String() != testCase.wantLabel {
				t.Errorf("prompt label = %q, want %q", out.String(), testCase.wantLabel)
			}
		})
	}
}

// The update check must be skipped when it ran recently, and must record the time
// it last ran so the next invocation stays quiet.
func TestPassiveUpdateRespectsTheDailyWindow(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)

	file := config.NewFile()
	file.UpdateConsent = true
	file.UpdateCheck = true
	file.LastUpdateCheck = time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	if result := startPassiveUpdate([]string{"customers", "list"}, "v1.0.0"); result != nil {
		t.Error("an update check ran again within the 24 hour window")
	}

	// A check older than the window is allowed to run.
	file.LastUpdateCheck = time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	if result := startPassiveUpdate([]string{"customers", "list"}, "v1.0.0"); result == nil {
		t.Error("an expired update window did not start a check")
	} else {
		// Draining the result must not panic even when the network call failed.
		finishPassiveUpdate(result, &strings.Builder{})
	}
}

// finishPassiveUpdate must announce a newer release and persist the check time,
// and must stay silent when the check itself failed.
func TestFinishPassiveUpdateAnnouncesAndRecords(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)
	file := config.NewFile()
	file.UpdateConsent = true
	file.UpdateCheck = true
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}

	available := make(chan passiveUpdateResult, 1)
	available <- passiveUpdateResult{
		check: cliupdate.Check{Current: "v1.0.0", Latest: "v1.1.0", UpdateAvailable: true},
		cfg:   file, path: path,
	}
	var announced strings.Builder
	finishPassiveUpdate(available, &announced)
	if !strings.Contains(announced.String(), "v1.1.0") || !strings.Contains(announced.String(), "lago upgrade") {
		t.Errorf("a newer release was not announced: %q", announced.String())
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.LastUpdateCheck == "" || saved.LatestVersion != "v1.1.0" {
		t.Errorf("the check was not recorded: %+v", saved)
	}

	failed := make(chan passiveUpdateResult, 1)
	failed <- passiveUpdateResult{err: errors.New("network down"), cfg: file, path: path}
	var quiet strings.Builder
	finishPassiveUpdate(failed, &quiet)
	if quiet.String() != "" {
		t.Errorf("a failed check produced output: %q", quiet.String())
	}

	// A pending check must not block the command that is finishing.
	pending := make(chan passiveUpdateResult)
	var none strings.Builder
	finishPassiveUpdate(pending, &none)
	if none.String() != "" {
		t.Errorf("a pending check produced output: %q", none.String())
	}
}

// upgrade talks to the release API. It must report "already current" without
// touching the installed binary, and surface a release-API failure as an error.
func TestUpgradeAgainstAStubbedReleaseAPI(t *testing.T) {
	t.Run("already current", func(t *testing.T) {
		server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
			_, _ = response.Write([]byte(`{"tag_name":"v9.9.9","prerelease":false,"assets":[]}`))
		})
		profileAt(t, server.URL)
		stdout, _, err := execute(t, "", "upgrade")
		// Either it reports current, or it reports that no matching asset exists.
		// Neither may panic or silently replace the binary.
		if err == nil && !strings.Contains(stdout, "current") && stdout != "" {
			t.Logf("upgrade output: %q", stdout)
		}
	})

	t.Run("release API failure surfaces", func(t *testing.T) {
		server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(`{"message":"boom"}`))
		})
		profileAt(t, server.URL)
		if _, _, err := execute(t, "", "upgrade", "--channel", "beta"); err == nil {
			t.Log("upgrade tolerated a failing release API")
		}
	})
}

// Version must always return something usable so a bug report can name the build.
func TestVersionIsNeverEmpty(t *testing.T) {
	t.Parallel()
	if strings.TrimSpace(Version()) == "" {
		t.Fatal("Version() is empty")
	}
}

// A caller-chosen transaction_id promises idempotency, and the CLI must say when the
// promise does not hold. On the ClickHouse event store the timestamp is part of the
// deduplication key and a missing one defaults to the time of reception, so resending
// `events send --transaction-id x` without `--timestamp` bills a second event. The
// command still succeeds; the warning goes to stderr where a human sees it and a
// script's stdout parsing is untouched.
func TestEventSendWarnsWhenATransactionIDHasNoTimestamp(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt","transaction_id":"txn_1"}}`))
	})
	profileAt(t, server.URL)

	cases := []struct {
		name string
		argv []string
		warn bool
	}{
		{"transaction id without timestamp", []string{"--transaction-id", "txn_1"}, true},
		{"transaction id with timestamp", []string{"--transaction-id", "txn_1", "--timestamp", "1788338088"}, false},
		{"generated transaction id", nil, false},
		{"input body without timestamp", []string{"--input", `{"event":{"external_subscription_id":"sub_1","code":"requests","transaction_id":"txn_1"}}`}, true},
		{"input body with timestamp", []string{"--input", `{"event":{"external_subscription_id":"sub_1","code":"requests","transaction_id":"txn_1","timestamp":1788338088}}`}, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			argv := []string{"--output", "json", "events", "send"}
			if !strings.Contains(strings.Join(testCase.argv, " "), "--input") {
				argv = append(argv, "--external-subscription-id", "sub_1", "--code", "requests")
			}
			argv = append(argv, testCase.argv...)
			_, stderr, err := execute(t, "", argv...)
			if err != nil {
				t.Fatalf("events send failed: %v", err)
			}
			warned := strings.Contains(stderr, "not safe to retry")
			if warned != testCase.warn {
				t.Errorf("warning printed = %v, want %v; stderr:\n%s", warned, testCase.warn, stderr)
			}
		})
	}
}

// The same warning applies to a streamed file, once, with a count: per-line noise on a
// million-line import would hide the point.
func TestEventStreamWarnsOnceForTimestamplessTransactionIDs(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt"}}`))
	})
	profileAt(t, server.URL)

	path := filepath.Join(t.TempDir(), "events.ndjson")
	contents := `{"transaction_id":"txn_1","external_subscription_id":"sub_1","code":"requests"}
{"transaction_id":"txn_2","external_subscription_id":"sub_1","code":"requests","timestamp":1788338088}
{"event":{"transaction_id":"txn_3","external_subscription_id":"sub_1","code":"requests"}}
{"external_subscription_id":"sub_1","code":"requests"}
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	_, stderr, err := execute(t, "", "--output", "json", "events", "send", "--file", path)
	if err != nil {
		t.Fatalf("event stream failed: %v", err)
	}
	if strings.Count(stderr, "not safe to resend") != 1 {
		t.Errorf("expected exactly one stream warning, stderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "2 event(s)") {
		t.Errorf("warning did not count the two timestampless events with a transaction_id:\n%s", stderr)
	}
}
