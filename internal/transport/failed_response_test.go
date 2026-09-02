package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
)

// QA F-11: a terminal network error used to come back with a nil response, so callers
// could not report how long the attempts took or how many there were.
func TestQA_F11_NetworkFailureReturnsTimedResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	client, err := New(Config{BaseURL: url, APIKey: "lago_test_FAKE000000000000000000000000", Insecure: true, Timeout: 5 * time.Second, NoRetry: true})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/customers", Idempotent: true})
	if err == nil {
		t.Fatal("request to a closed server succeeded")
	}
	if apperr.ExitCode(err) != apperr.ExitNetwork {
		t.Errorf("exit code = %d, want %d", apperr.ExitCode(err), apperr.ExitNetwork)
	}
	if response == nil {
		t.Fatal("network failure returned a nil response")
	}
	if response.Attempts != 1 || response.Timing.AttemptCount != 1 {
		t.Errorf("attempts = %d / %d, want 1", response.Attempts, response.Timing.AttemptCount)
	}
	if response.Timing.Total <= 0 {
		t.Errorf("total timing = %s, want > 0", response.Timing.Total)
	}
	if response.Status != 0 || len(response.Body) != 0 {
		t.Errorf("failed response carries status %d and %d body bytes; want none", response.Status, len(response.Body))
	}
}
