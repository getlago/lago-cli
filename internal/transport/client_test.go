package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
)

func TestClientRetriesIdempotentServerFailureAndRedactsLogs(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	secret := "lago_test_FAKEabcdefghijklmnopqrstuv"
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/organizations" {
			t.Errorf("path = %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer "+secret {
			t.Errorf("authorization was not sent")
		}
		if attempts.Add(1) == 1 {
			response.Header().Set("X-Request-Id", "req_fake_retry")
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(response, `{"error":"temporary"}`)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"organization":{"name":"Test"}}`)
	}))
	defer server.Close()
	var logs bytes.Buffer
	client, err := New(Config{BaseURL: server.URL, APIKey: secret, Insecure: true, Timeout: 2 * time.Second, Verbose: true, Err: &logs})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Attempts != 2 || attempts.Load() != 2 {
		t.Fatalf("attempts = %d/%d", result.Attempts, attempts.Load())
	}
	if strings.Contains(logs.String(), secret) || !strings.Contains(logs.String(), "[REDACTED]") {
		t.Fatalf("verbose logs were not redacted: %s", logs.String())
	}
}

func TestClientDoesNotRetryNonIdempotentMutation(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		response.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL, APIKey: "fake", Insecure: true, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), Request{Method: http.MethodPost, Path: "/events"})
	if apperr.ExitCode(err) != apperr.ExitServer || attempts.Load() != 1 {
		t.Fatalf("exit=%d attempts=%d error=%v", apperr.ExitCode(err), attempts.Load(), err)
	}
}

func TestClientClassifiesAuthAndSurfacesRequestID(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("X-Request-Id", "req_fake_auth")
		response.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(response, `{"error":"invalid API key","code":"unauthorized"}`)
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL, APIKey: "fake", Insecure: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
	appError, ok := err.(*apperr.Error)
	if !ok || appError.ExitCode != apperr.ExitAuth || appError.RequestID != "req_fake_auth" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestDryRunNeverCallsServerAndRedactsAuthorization(t *testing.T) {
	t.Parallel()
	client, err := New(Config{BaseURL: "https://api.getlago.com", APIKey: "lago_test_FAKEabcdefghijklmnopqrstuv"})
	if err != nil {
		t.Fatal(err)
	}
	headers := make(http.Header)
	headers.Set("X-Webhook-Secret", "private-value")
	response, err := client.Do(context.Background(), Request{Method: http.MethodPost, Path: "/events", Query: map[string][]string{"api_key": {"query-secret"}}, Headers: headers, Body: []byte(`{"api_key":"body-secret","event":{}}`), DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	encoded := response.DryRunData["headers"].(http.Header).Get("Authorization")
	if encoded != "Bearer [REDACTED]" {
		t.Fatalf("authorization = %q", encoded)
	}
	serialized, _ := json.Marshal(response.DryRunData)
	for _, secret := range []string{"private-value", "query-secret", "body-secret", "abcdefghijklmnopqrstuv"} {
		if strings.Contains(string(serialized), secret) {
			t.Fatalf("dry run leaked %q: %s", secret, serialized)
		}
	}
}

func TestNormalizeBaseURLRequiresExplicitInsecure(t *testing.T) {
	t.Parallel()
	if _, err := NormalizeBaseURL("http://localhost:3000", false); apperr.ExitCode(err) != apperr.ExitUsage {
		t.Fatalf("HTTP without insecure returned %v", err)
	}
	parsed, err := NormalizeBaseURL("http://localhost:3000/", true)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.String() != "http://localhost:3000/api/v1" {
		t.Fatalf("normalized URL = %s", parsed)
	}
}

func TestClientRetriesTruncatedIdempotentResponse(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			response.Header().Set("Content-Length", "100")
			_, _ = io.WriteString(response, "{}")
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"ok":true}`)
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL, APIKey: "fake", Insecure: true, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
	if err != nil {
		t.Fatal(err)
	}
	if response.Attempts != 2 {
		t.Fatalf("attempts = %d", response.Attempts)
	}
}

func TestClientHonorsZeroRetryAfter(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			response.Header().Set("Retry-After", "0")
			response.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(response, `{}`)
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL, APIKey: "fake", Insecure: true, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
	if err != nil || response.Attempts != 2 {
		t.Fatalf("response=%#v error=%v", response, err)
	}
}

func TestClientTimeoutUsesNetworkExitCode(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = io.WriteString(response, `{}`)
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL, APIKey: "fake", Insecure: true, Timeout: 10 * time.Millisecond, NoRetry: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
	if apperr.ExitCode(err) != apperr.ExitNetwork {
		t.Fatalf("exit=%d error=%v", apperr.ExitCode(err), err)
	}
}
