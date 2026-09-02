package cli

import (
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// QA N-9: a request path carrying userinfo must not become a request to another host.
func TestQA_N9_APIPathWithUserinfoIsRejected(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) {
		t.Error("a path with userinfo reached the API")
	})
	profileAt(t, server.URL)
	for _, path := range []string{"//user:pw@evil.example/customers", "https://api.getlago.com@evil.example/customers"} {
		_, _, err := execute(t, "", "api", "GET", path)
		if err == nil {
			t.Errorf("api GET %s was accepted", path)
			continue
		}
		if apperr.ExitCode(err) != apperr.ExitUsage {
			t.Errorf("%s: exit code = %d, want %d", path, apperr.ExitCode(err), apperr.ExitUsage)
		}
	}
}

// QA S-16: init refuses a URL with credentials before saving anything.
func TestQA_S16_InitRefusesAURLWithCredentials(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)
	_, _, err := execute(t, "", "--api-key", "lago_test_FAKE000000000000000000000000", "--mode", "test", "init", "--region", "self-hosted", "--api-url", "https://api.getlago.com@evil.example")
	if err == nil {
		t.Fatal("init saved a profile with credentials in the URL")
	}
	if apperr.ExitCode(err) != apperr.ExitUsage || !strings.Contains(err.Error(), "embeds credentials") {
		t.Errorf("unexpected refusal: %v", err)
	}
}

// QA F-19: --limit 0 reached the API as per_page=0 and lago-api answered 500. The bound
// is checked client-side, at exit 2, before any request.
func TestQA_F19_LimitMustBeAnIntegerInRange(t *testing.T) {
	var mutex sync.Mutex
	var queries []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		queries = append(queries, request.URL.RawQuery)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"customers":[],"meta":{"total_pages":1}}`))
	})
	profileAt(t, server.URL)

	for _, bad := range []string{"0", "1001", "-1", "abc", "1.5", ""} {
		_, _, err := execute(t, "", "customers", "list", "--limit", bad)
		if err == nil {
			t.Errorf("--limit %q was accepted", bad)
			continue
		}
		if apperr.ExitCode(err) != apperr.ExitUsage || !strings.Contains(err.Error(), "between 1 and 1000") {
			t.Errorf("--limit %q: unexpected refusal %v", bad, err)
		}
	}
	mutex.Lock()
	if len(queries) != 0 {
		t.Fatalf("an invalid --limit reached the API: %v", queries)
	}
	mutex.Unlock()

	for _, good := range []string{"1", "1000", " 25 "} {
		if _, _, err := execute(t, "", "--output", "json", "customers", "list", "--limit", good); err != nil {
			t.Errorf("--limit %q was refused: %v", good, err)
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	if strings.Join(queries, "|") != "per_page=1|per_page=1000|per_page=25" {
		t.Errorf("forwarded queries = %v", queries)
	}
}

func TestQA_F19_PageMustBePositive(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"customers":[],"meta":{"total_pages":1}}`))
	})
	profileAt(t, server.URL)
	for _, bad := range []string{"0", "-2", "two"} {
		_, _, err := execute(t, "", "customers", "list", "--page", bad)
		if err == nil || apperr.ExitCode(err) != apperr.ExitUsage || !strings.Contains(err.Error(), "positive integer") {
			t.Errorf("--page %q: unexpected result %v", bad, err)
		}
	}
	if _, _, err := execute(t, "", "--output", "json", "customers", "list", "--page", "3"); err != nil {
		t.Errorf("--page 3 was refused: %v", err)
	}
}
