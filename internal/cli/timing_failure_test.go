package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// QA F-10: --timing printed nothing when the request failed, which is exactly when the
// latency and attempt breakdown is needed. A failed HTTP request now prints one timing
// line before the error.
func TestQA_F10_TimingIsPrintedOnHTTPFailure(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte(`{"status":500,"error":"Internal Server Error"}`))
	})
	profileAt(t, server.URL)

	var stdout, stderr bytes.Buffer
	code := ExecuteArgs([]string{"--timing", "--no-retry", "customers", "get", "cus_1"}, strings.NewReader(""), &stdout, &stderr)
	if code != apperr.ExitServer {
		t.Fatalf("exit code = %d, want %d\n%s", code, apperr.ExitServer, stderr.String())
	}
	if strings.Count(stderr.String(), "timing: ") != 1 {
		t.Fatalf("want exactly one timing line on failure:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), `"attempt_count":1`) {
		t.Errorf("timing line lacks attempt_count:\n%s", stderr.String())
	}
	if strings.Index(stderr.String(), "timing: ") > strings.Index(stderr.String(), "Error: ") {
		t.Errorf("timing should print before the error summary:\n%s", stderr.String())
	}
}

// QA F-11: a network failure used to return no response at all, so retry behaviour was
// invisible. The transport now returns the timing it measured, and --timing prints it.
func TestQA_F11_TimingIsPrintedOnNetworkFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()
	profileAt(t, url)

	var stdout, stderr bytes.Buffer
	code := ExecuteArgs([]string{"--timing", "--no-retry", "api", "GET", "/customers"}, strings.NewReader(""), &stdout, &stderr)
	if code != apperr.ExitNetwork {
		t.Fatalf("exit code = %d, want %d\n%s", code, apperr.ExitNetwork, stderr.String())
	}
	if strings.Count(stderr.String(), "timing: ") != 1 {
		t.Fatalf("want exactly one timing line on network failure:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), `"attempt_count":1`) {
		t.Errorf("timing line lacks attempt_count:\n%s", stderr.String())
	}
}

// A successful request still prints exactly one timing line: the failure path must not
// double up with the render path.
func TestTimingPrintsOnceOnSuccess(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1"}}`))
	})
	profileAt(t, server.URL)
	var stdout, stderr bytes.Buffer
	if code := ExecuteArgs([]string{"--timing", "customers", "get", "cus_1"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d\n%s", code, stderr.String())
	}
	if strings.Count(stderr.String(), "timing: ") != 1 {
		t.Fatalf("want exactly one timing line on success:\n%s", stderr.String())
	}
}
