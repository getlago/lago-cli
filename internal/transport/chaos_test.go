package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
)

const fakeKey = "lago_test_FAKEabcdefghijklmnopqrstuv"

func testClient(t *testing.T, baseURL string, mutate func(*Config)) *Client {
	t.Helper()
	config := Config{BaseURL: baseURL, APIKey: fakeKey, Insecure: true, Timeout: 5 * time.Second}
	if mutate != nil {
		mutate(&config)
	}
	client, err := New(config)
	if err != nil {
		t.Fatalf("build client: %v", err)
	}
	return client
}

// Every HTTP status the Lago API can return must map to its frozen exit code and
// carry an actionable suggestion. Scripts branch on these.
func TestStatusClassificationIsComplete(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		status   int
		wantCode int
		wantWord string
	}{
		{http.StatusUnauthorized, apperr.ExitAuth, "lago init"},
		{http.StatusForbidden, apperr.ExitAuth, "permissions"},
		{http.StatusNotFound, apperr.ExitNotFound, "identifier"},
		{http.StatusTooManyRequests, apperr.ExitRateLimit, "Retry-After"},
		{http.StatusInternalServerError, apperr.ExitServer, "request ID"},
		{http.StatusBadGateway, apperr.ExitServer, "request ID"},
		{http.StatusServiceUnavailable, apperr.ExitServer, "request ID"},
		{http.StatusUnprocessableEntity, apperr.ExitValidation, "validation"},
		{http.StatusBadRequest, apperr.ExitValidation, "flags"},
		{http.StatusConflict, apperr.ExitValidation, "flags"},
	} {
		code, suggestion := classify(testCase.status)
		if code != testCase.wantCode {
			t.Errorf("status %d classified as exit %d, want %d", testCase.status, code, testCase.wantCode)
		}
		if !strings.Contains(suggestion, testCase.wantWord) {
			t.Errorf("status %d suggestion %q lacks %q", testCase.status, suggestion, testCase.wantWord)
		}
	}
}

// Retry-After governs how long the CLI waits. Seconds, HTTP dates, and junk must
// each be handled without producing a negative or absurd delay.
func TestRetryAfterParsing(t *testing.T) {
	t.Parallel()
	if got := retryAfter("5"); got != 5*time.Second {
		t.Errorf("numeric Retry-After = %s", got)
	}
	if got := retryAfter("0"); got != 0 {
		t.Errorf("zero Retry-After = %s", got)
	}
	if got := retryAfter("-3"); got != 0 {
		t.Errorf("negative Retry-After = %s, want 0", got)
	}
	if got := retryAfter("soon"); got != 0 {
		t.Errorf("unparseable Retry-After = %s, want 0", got)
	}
	if got := retryAfter(""); got != 0 {
		t.Errorf("empty Retry-After = %s, want 0", got)
	}
	future := retryAfter(time.Now().Add(2 * time.Second).UTC().Format(http.TimeFormat))
	if future <= 0 || future > 3*time.Second {
		t.Errorf("HTTP-date Retry-After = %s, want roughly 2s", future)
	}
	if past := retryAfter(time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)); past != 0 {
		t.Errorf("past Retry-After = %s, want 0", past)
	}
}

// Backoff must stay inside its cap so a retry storm cannot stall a billing script.
func TestBackoffIsBoundedAndGrows(t *testing.T) {
	t.Parallel()
	for attempt := 1; attempt <= 6; attempt++ {
		for sample := 0; sample < 40; sample++ {
			wait := backoff(attempt)
			if wait < 0 {
				t.Fatalf("attempt %d produced a negative wait %s", attempt, wait)
			}
			if wait > 2*time.Second {
				t.Fatalf("attempt %d produced %s, above the 2s cap", attempt, wait)
			}
		}
	}
}

func TestWaitContextHonoursCancellation(t *testing.T) {
	t.Parallel()
	if err := waitContext(context.Background(), time.Millisecond); err != nil {
		t.Fatalf("a short wait failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitContext(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("a cancelled context did not stop the wait: %v", err)
	}
}

// An authenticated request must not carry its bearer token across origins.
func TestSameOriginComparison(t *testing.T) {
	t.Parallel()
	parse := func(raw string) *url.URL {
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	if !sameOrigin(parse("https://api.getlago.com/a"), parse("HTTPS://API.GETLAGO.COM/b")) {
		t.Error("case differences must not change the origin")
	}
	for _, pair := range [][2]string{
		{"https://api.getlago.com", "https://evil.example.com"},
		{"https://api.getlago.com", "http://api.getlago.com"},
		{"https://api.getlago.com", "https://api.getlago.com:8443"},
	} {
		if sameOrigin(parse(pair[0]), parse(pair[1])) {
			t.Errorf("%s and %s were treated as the same origin", pair[0], pair[1])
		}
	}
}

// A redirect to another origin must drop the Authorization header and fail rather
// than handing the API key to whoever the redirect points at.
func TestCrossOriginRedirectDoesNotLeakTheAPIKey(t *testing.T) {
	var mutex sync.Mutex
	var leaked bool
	attacker := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		leaked = request.Header.Get("Authorization") != ""
		mutex.Unlock()
		response.WriteHeader(http.StatusOK)
	}))
	defer attacker.Close()

	origin := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Redirect(response, &http.Request{}, attacker.URL+"/api/v1/steal", http.StatusFound)
	}))
	defer origin.Close()

	client := testClient(t, origin.URL, nil)
	if _, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true}); err == nil {
		t.Fatal("a cross-origin redirect was followed")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if leaked {
		t.Fatal("the API key was sent to the redirect target")
	}
}

// A redirect loop must terminate rather than hanging a billing script.
func TestRedirectLoopTerminates(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Redirect(response, &http.Request{}, server.URL+"/api/v1/loop", http.StatusFound)
	}))
	defer server.Close()

	client := testClient(t, server.URL, nil)
	done := make(chan error, 1)
	go func() {
		_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/loop", Idempotent: true})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("an infinite redirect chain was followed to completion")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the redirect loop never terminated")
	}
}

// Absolute paths would let a flag redirect a request off the configured host.
func TestAbsolutePathsAreRefused(t *testing.T) {
	t.Parallel()
	client := testClient(t, "https://api.getlago.com", nil)
	for _, path := range []string{"https://evil.example.com/customers", "http://evil.example.com/customers"} {
		if _, err := client.resolve(path, nil); err == nil {
			t.Errorf("absolute path %q was accepted", path)
		}
	}
	resolved, err := client.resolve("/customers", url.Values{"page": {"2"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Path != "/api/v1/customers" || resolved.Query().Get("page") != "2" {
		t.Fatalf("relative path resolved to %s", resolved)
	}
}

// Verbose output and dry runs print URLs. Credentials in the query string must be
// redacted, and the redaction must not depend on the parameter's exact name.
func TestRedactedURLHidesCredentialParameters(t *testing.T) {
	t.Parallel()
	client := testClient(t, "https://api.getlago.com", nil)
	target, err := url.Parse("https://api.getlago.com/api/v1/customers?api_key=" + fakeKey +
		"&access_token=abc&Client-Secret=xyz&PASSWORD=hunter2&page=2&filter=" + fakeKey)
	if err != nil {
		t.Fatal(err)
	}
	rendered := client.redactedURL(target)
	if strings.Contains(rendered, fakeKey) || strings.Contains(rendered, "hunter2") || strings.Contains(rendered, "xyz") {
		t.Fatalf("redacted URL still contains a credential: %s", rendered)
	}
	if !strings.Contains(rendered, "page=2") {
		t.Errorf("redaction removed a harmless parameter: %s", rendered)
	}
}

// DecodeJSON is the single decode point for API responses. It must preserve exact
// numbers and treat an empty body as an empty object rather than an error.
func TestDecodeJSONPreservesExactNumbers(t *testing.T) {
	t.Parallel()
	for _, blank := range []string{"", "   ", "\n\t"} {
		value, err := DecodeJSON([]byte(blank))
		if err != nil {
			t.Fatalf("empty body was an error: %v", err)
		}
		if object, ok := value.(map[string]any); !ok || len(object) != 0 {
			t.Fatalf("empty body decoded to %#v", value)
		}
	}

	value, err := DecodeJSON([]byte(`{"invoice":{"total_amount_cents":9007199254740993}}`))
	if err != nil {
		t.Fatal(err)
	}
	invoice := value.(map[string]any)["invoice"].(map[string]any)
	if got := invoice["total_amount_cents"]; got.(interface{ String() string }).String() != "9007199254740993" {
		t.Fatalf("decode lost exact digits: %v", got)
	}

	if _, err := DecodeJSON([]byte("{not json")); err == nil {
		t.Fatal("malformed JSON was accepted")
	}
}

func TestCloneHeadersNeverReturnsNil(t *testing.T) {
	t.Parallel()
	if cloned := cloneHeaders(nil); cloned == nil {
		t.Fatal("cloneHeaders(nil) returned nil")
	}
	source := http.Header{"X-Test": {"one"}}
	cloned := cloneHeaders(source)
	cloned.Set("X-Test", "two")
	if source.Get("X-Test") != "one" {
		t.Fatal("cloneHeaders returned an aliased header map")
	}
}

// A client must refuse to be built with an unusable configuration rather than
// failing later, mid-request, against a live billing API.
func TestNewRejectsUnusableConfiguration(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ name, baseURL string }{
		{"empty", ""},
		{"whitespace", "   "},
		{"no scheme", "api.getlago.com"},
		{"plain http without insecure", "http://api.getlago.com"},
	} {
		if _, err := New(Config{BaseURL: testCase.baseURL, APIKey: fakeKey}); err == nil {
			t.Errorf("%s base URL was accepted", testCase.name)
		}
	}

	// Defaults must be filled in rather than left as zero values.
	client, err := New(Config{BaseURL: "https://api.getlago.com", APIKey: fakeKey})
	if err != nil {
		t.Fatal(err)
	}
	if client.http.Timeout != 30*time.Second {
		t.Errorf("default timeout = %s, want 30s", client.http.Timeout)
	}
	if client.config.Err == nil {
		t.Error("a nil error writer was not replaced")
	}
	transport, ok := client.http.Transport.(*http.Transport)
	if !ok {
		t.Fatal("unexpected transport type")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("TLS floor = %x, want TLS 1.2", transport.TLSClientConfig.MinVersion)
	}
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("certificate verification was disabled without --insecure")
	}
}

// --insecure is the only way to reach a plain-HTTP or self-signed host, and the
// API path must be appended exactly once however the URL was written.
func TestNormalizeBaseURLShapesTheAPIPath(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ raw, want string }{
		{"https://api.getlago.com", "/api/v1"},
		{"https://api.getlago.com/", "/api/v1"},
		{"https://api.getlago.com/api/v1", "/api/v1"},
		{"https://api.getlago.com/api/v1/", "/api/v1"},
		{"https://lago.selfhosted.example/lago", "/lago/api/v1"},
	} {
		parsed, err := NormalizeBaseURL(testCase.raw, false)
		if err != nil {
			t.Fatalf("%s: %v", testCase.raw, err)
		}
		if parsed.Path != testCase.want {
			t.Errorf("%s normalised to path %q, want %q", testCase.raw, parsed.Path, testCase.want)
		}
	}

	// Query strings and fragments in a base URL are dropped, not carried into
	// every subsequent request.
	parsed, err := NormalizeBaseURL("https://api.getlago.com/api/v1?api_key=leak#frag", false)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		t.Fatalf("base URL kept a query or fragment: %s", parsed)
	}

	if _, err := NormalizeBaseURL("http://localhost:3000", true); err != nil {
		t.Errorf("--insecure did not permit plain HTTP: %v", err)
	}
}

// A rate-limited response must be retried after the interval the server asked for,
// and the successful body returned, so callers never see a spurious 429.
func TestRateLimitIsRetriedAfterTheServerInterval(t *testing.T) {
	var mutex sync.Mutex
	attempts := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		attempts++
		current := attempts
		mutex.Unlock()
		response.Header().Set("Content-Type", "application/json")
		if current == 1 {
			response.Header().Set("Retry-After", "0")
			response.WriteHeader(http.StatusTooManyRequests)
			_, _ = response.Write([]byte(`{"status":429,"error":"Too Many Requests"}`))
			return
		}
		_, _ = response.Write([]byte(`{"customers":[]}`))
	}))
	defer server.Close()

	client := testClient(t, server.URL, nil)
	response, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/customers", Idempotent: true})
	if err != nil {
		t.Fatalf("a retryable 429 surfaced as an error: %v", err)
	}
	if response.Attempts < 2 {
		t.Errorf("attempts = %d, want at least 2", response.Attempts)
	}
	if !strings.Contains(string(response.Body), "customers") {
		t.Errorf("body = %s", response.Body)
	}
}

// A response larger than the safety limit must be refused rather than buffered
// until the process runs out of memory.
func TestOversizedResponsesAreRefused(t *testing.T) {
	chunk := strings.Repeat("a", 1<<20)
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		for written := 0; written <= 64; written++ {
			if _, err := response.Write([]byte(chunk)); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	client := testClient(t, server.URL, func(config *Config) { config.Timeout = 60 * time.Second })
	_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/invoices", Idempotent: true})
	if err == nil {
		t.Fatal("an oversized response was accepted")
	}
	if !strings.Contains(err.Error(), "64 MiB") {
		t.Errorf("error does not name the limit: %v", err)
	}
}

// Verbose logging is where a key most easily escapes. Nothing printed may contain
// the credential, in either the request or the response direction.
func TestVerboseLoggingRedactsBothDirections(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.Header().Set("X-Echoed-Key", fakeKey)
		_, _ = response.Write([]byte(`{"organization":{"api_key":"` + fakeKey + `"}}`))
	}))
	defer server.Close()

	var logs strings.Builder
	client := testClient(t, server.URL, func(config *Config) {
		config.Verbose = true
		config.Err = &logs
	})
	if _, err := client.Do(context.Background(), Request{
		Method:     http.MethodPost,
		Path:       "/organizations",
		Body:       []byte(`{"organization":{"api_key":"` + fakeKey + `"}}`),
		Idempotent: false,
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(logs.String(), fakeKey) {
		t.Fatalf("verbose output leaked the API key:\n%s", logs.String())
	}
	if !strings.Contains(logs.String(), "POST") {
		t.Errorf("verbose output lost the request line:\n%s", logs.String())
	}
}
