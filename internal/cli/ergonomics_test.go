package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

func TestNormalizeEventTimestamp(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ in, want string }{
		{"1788338088", "1788338088"},
		{"1788338088.123", "1788338088.123"},
		{" 1788338088 ", "1788338088"},
		{"2026-09-01T14:34:48Z", "1788273288"},
		{"2026-09-01T16:34:48+02:00", "1788273288"},
		{"2026-09-01T14:34:48.250Z", "1788273288.25"},
		{"2026-09-01T14:34:48.000000001Z", "1788273288.000000001"},
		{"", ""},
	} {
		got, err := normalizeEventTimestamp(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("normalizeEventTimestamp(%q) = %q, %v; want %q", tc.in, got, err, tc.want)
		}
	}
	for _, bad := range []string{"yesterday", "2026-09-01", "1e9", "2026-09-01 14:34:48"} {
		if _, err := normalizeEventTimestamp(bad); err == nil {
			t.Errorf("%q was accepted as a timestamp", bad)
		}
	}
}

// QA run 3 and run 4: a millisecond epoch passed straight through and would have been
// stored 3000 years out; `2026`, `+17`, `-5`, `Inf`, `NaN` and hex all parsed as
// numbers. Digits only, inside a plausible window, everything else is refused with a
// message that names the likely mistake.
func TestQA_Timestamp_RejectsImplausibleNumbers(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ in, mention string }{
		{"1788273288000", "millisecond"},
		{"1788273288000.5", "millisecond"},
		{"100000000000", "millisecond"},
		{"2026", "too small"},
		{"999999999", "too small"},
		{"0", "too small"},
		{"+1788273288", "neither"},
		{"-5", "neither"},
		{"Inf", "neither"},
		{"NaN", "neither"},
		{"0x1p3", "neither"},
		{"1_000_000_000", "neither"},
	} {
		got, err := normalizeEventTimestamp(tc.in)
		if err == nil {
			t.Errorf("%q was accepted as %q", tc.in, got)
			continue
		}
		if !strings.Contains(err.Error(), tc.mention) {
			t.Errorf("%q: error %q does not mention %q", tc.in, err, tc.mention)
		}
	}
	for _, ok := range []string{"1000000000", "1788273288", "1788273288.5", "99999999999"} {
		if _, err := normalizeEventTimestamp(ok); err != nil {
			t.Errorf("%q was refused: %v", ok, err)
		}
	}
}

// The same window applies when the timestamp arrives as a JSON number, through
// --input or an NDJSON line, where no string normalization runs.
func TestQA_Timestamp_JSONNumbersAreCheckedToo(t *testing.T) {
	t.Parallel()
	if _, err := normalizeEventBodyTimestamp([]byte(`{"event":{"code":"c","external_subscription_id":"s","timestamp":1788273288000}}`)); err == nil {
		t.Error("--input body with a millisecond epoch was accepted")
	}
	if _, err := normalizeEventBodyTimestamp([]byte(`{"event":{"code":"c","external_subscription_id":"s","timestamp":1788273288}}`)); err != nil {
		t.Errorf("--input body with plausible seconds was refused: %v", err)
	}
	if _, _, _, err := prepareEvent([]byte(`{"code":"c","external_subscription_id":"s","transaction_id":"t","timestamp":1788273288000}`)); err == nil {
		t.Error("NDJSON line with a millisecond epoch was accepted")
	}
	if _, _, _, err := prepareEvent([]byte(`{"code":"c","external_subscription_id":"s","transaction_id":"t","timestamp":1788273288}`)); err != nil {
		t.Errorf("NDJSON line with plausible seconds was refused: %v", err)
	}
}

// An RFC 3339 --timestamp reaches the API as Unix seconds, on the flag path, the --input
// path, and the --file stream alike. Unix seconds pass through untouched.
func TestEventsTimestampAcceptsRFC3339(t *testing.T) {
	var mutex sync.Mutex
	var timestamps []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		var payload map[string]map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		mutex.Lock()
		timestamps = append(timestamps, strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(jsonString(payload["event"]["timestamp"]), `"`), `"`), " ", "")))
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt"}}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "--output", "json", "events", "send", "--code", "requests", "--external-subscription-id", "sub_1", "--timestamp", "2026-09-01T14:34:48Z"); err != nil {
		t.Fatalf("flag path failed: %v", err)
	}
	if _, _, err := execute(t, "", "--output", "json", "events", "send", "--code", "requests", "--external-subscription-id", "sub_1", "--timestamp", "1788359688"); err != nil {
		t.Fatalf("unix flag path failed: %v", err)
	}
	if _, _, err := execute(t, "", "--output", "json", "events", "send", "--input", `{"event":{"code":"requests","external_subscription_id":"sub_1","transaction_id":"t1","timestamp":"2026-09-01T14:34:48.5Z"}}`); err != nil {
		t.Fatalf("input path failed: %v", err)
	}
	path := filepath.Join(t.TempDir(), "events.ndjson")
	if err := os.WriteFile(path, []byte(`{"code":"requests","external_subscription_id":"sub_1","timestamp":"2026-09-01T16:34:48+02:00"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute(t, "", "--output", "json", "events", "send", "--file", path); err != nil {
		t.Fatalf("file path failed: %v", err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if strings.Join(timestamps, "|") != "1788273288|1788359688|1788273288.5|1788273288" {
		t.Errorf("timestamps on the wire = %v", timestamps)
	}
}

func TestEventsTimestampRejectsGarbage(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) {
		t.Error("an invalid timestamp reached the API")
	})
	profileAt(t, server.URL)
	_, _, err := execute(t, "", "events", "send", "--code", "requests", "--external-subscription-id", "sub_1", "--timestamp", "yesterday")
	if err == nil || apperr.ExitCode(err) != apperr.ExitUsage || !strings.Contains(err.Error(), "RFC 3339") {
		t.Errorf("unexpected result for a garbage timestamp: %v", err)
	}
	_, _, err = execute(t, "", "events", "send", "--input", `{"event":{"code":"requests","external_subscription_id":"sub_1","timestamp":"yesterday"}}`)
	if err == nil || apperr.ExitCode(err) != apperr.ExitUsage {
		t.Errorf("unexpected result for a garbage --input timestamp: %v", err)
	}
}

// The bash completion script is cobra's V2 form, which asks the binary at runtime
// instead of embedding every flag of every command: kilobytes, not hundreds of them.
func TestBashCompletionIsTheCompactV2Script(t *testing.T) {
	setCleanEnvironment(t)
	stdout, _, err := execute(t, "", "completion", "bash")
	if err != nil {
		t.Fatal(err)
	}
	if len(stdout) > 64<<10 {
		t.Errorf("bash completion is %d bytes; the V2 script is a few KB", len(stdout))
	}
	if !strings.Contains(stdout, "__complete") || !strings.Contains(stdout, "__lago_") {
		t.Error("bash completion is not the dynamic V2 script")
	}
	checked, err := os.ReadFile(filepath.Join("..", "..", "completions", "lago.bash"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked) > 64<<10 {
		t.Errorf("checked-in completions/lago.bash is %d bytes; run make generate", len(checked))
	}
}

func jsonString(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

var _ = io.Discard
