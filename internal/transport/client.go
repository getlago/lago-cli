package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/redact"
)

// USAPI and EUAPI are the cloud base URLs behind `--region us` and `--region eu`. They
// carry no /api/v1: the CLI appends it, so the region shorthand and an explicitly passed
// base URL normalize to exactly the same host and path.
const (
	USAPI = "https://api.getlago.com"
	EUAPI = "https://api.eu.getlago.com"
)

type Config struct {
	BaseURL   string
	APIKey    string
	Timeout   time.Duration
	Insecure  bool
	NoRetry   bool
	Verbose   bool
	UserAgent string
	Err       io.Writer

	// DialContext replaces the default dialer. It exists so the CLI's own tests can
	// exercise the production hostnames -- api.getlago.com, api.eu.getlago.com, a
	// self-hosted host on a custom port -- against a local server, instead of proving
	// URL handling only for 127.0.0.1. Nil means the default dialer, which is the only
	// thing any released code path uses.
	DialContext func(ctx context.Context, network, address string) (net.Conn, error)
}

type Client struct {
	baseURL  *url.URL
	http     *http.Client
	config   Config
	redactor *redact.Redactor
}

type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers http.Header
	Body    []byte
	// Idempotent allows bounded retries on network errors, 429, and 5xx. Callers set it
	// only for reads and for usage events that carry a timestamp: lago-api does not read
	// an Idempotency-Key header, so no other mutation is safe to replay.
	Idempotent bool
	DryRun     bool

	// Subjects are the identifiers this request addressed, used to turn the API's bare
	// 404 into a message naming what was not found. See Subject.
	Subjects []Subject
}

// Subject is one identifier a request addressed: the kind of thing it names, and the
// value that was passed.
//
// Lago answers a wrong identifier with `404 Not Found` and nothing else. QA passed a
// plan code where a subscription external ID belonged (`ai_plan_...` as
// `--external-subscription-id`) and got a bare "Not Found" that read as "this
// subscription has no usage" rather than "this subscription does not exist". Carrying
// the identifiers into the error is what makes the difference visible.
type Subject struct {
	Kind  string
	Value string
}

type Response struct {
	Status     int
	Headers    http.Header
	Body       []byte
	RequestID  string
	Attempts   int
	Timing     Timing
	DryRunData map[string]any
}

type Timing struct {
	DNS           time.Duration `json:"dns"`
	Connect       time.Duration `json:"connect"`
	TLS           time.Duration `json:"tls"`
	TimeToFirst   time.Duration `json:"time_to_first_byte"`
	Download      time.Duration `json:"download"`
	RoundTrip     time.Duration `json:"api_round_trip"`
	RetryWait     time.Duration `json:"retry_wait"`
	Total         time.Duration `json:"total"`
	CLIOverhead   time.Duration `json:"cli_overhead"`
	AttemptCount  int           `json:"attempt_count"`
	ResponseBytes int           `json:"response_bytes"`
}

func New(cfg Config) (*Client, error) {
	baseURL, err := NormalizeBaseURL(cfg.BaseURL, cfg.Insecure)
	if err != nil {
		return nil, err
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Err == nil {
		cfg.Err = io.Discard
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	dialContext := cfg.DialContext
	if dialContext == nil {
		dialContext = dialer.DialContext
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			// #nosec G402 -- this can only be enabled through the explicit, loudly warned --insecure path.
			InsecureSkipVerify: cfg.Insecure,
		},
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("stopped after 5 redirects")
			}
			if len(via) > 0 && !sameOrigin(via[0].URL, req.URL) {
				req.Header.Del("Authorization")
				return errors.New("refusing cross-origin redirect for authenticated request")
			}
			return nil
		},
	}
	return &Client{baseURL: baseURL, http: httpClient, config: cfg, redactor: redact.New(cfg.APIKey)}, nil
}

// APIPrefix is the path segment every Lago API request lives under. The CLI appends it,
// so an operator passes a base URL and pasting the full path is equally correct.
const APIPrefix = "/api/v1"

// NormalizeBaseURL turns anything an operator might paste into the one base URL the
// client will use, or refuses it with a message naming the correct host.
//
// It is the single normalizer for every deployment target: cloud US, cloud EU, and any
// self-hosted shape including custom ports and sub-paths behind a proxy. QA hit two
// failures this closes. A pasted `https://api.getlago.com/api/v1` must not become a
// request to `/api/v1/api/v1`, and pasting the Lago **app** URL must not silently send an
// authenticated request with a live API key to the dashboard: `app.getlago.com` answers,
// it just answers with HTML, so the failure reads as a confusing parse error rather than
// "you used the wrong host".
func NormalizeBaseURL(raw string, insecure bool) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, apperr.New(apperr.ExitAuth, "Lago API URL is not configured", "Run `lago init` or set LAGO_API_URL.")
	}
	// A bare host such as `api.getlago.com` parses as a relative path with no scheme.
	// Naming that is more useful than reporting an unspecific "invalid URL".
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, apperr.New(apperr.ExitUsage,
			fmt.Sprintf("invalid Lago API URL %q", raw),
			"Use an absolute URL with a scheme, such as https://api.getlago.com for cloud US, https://api.eu.getlago.com for cloud EU, or your own base URL when self-hosting.")
	}
	// QA S-16, N-9: `https://api.getlago.com@evil.example` parses with a host of
	// evil.example and a userinfo of api.getlago.com. Whatever the CLI printed as the
	// host had to be the host it dialled, so userinfo is refused outright rather than
	// stripped. Redacted() keeps a pasted password out of the error text.
	if parsed.User != nil {
		return nil, apperr.New(apperr.ExitUsage,
			fmt.Sprintf("Lago API URL %q embeds credentials before the host", parsed.Redacted()),
			"Remove the user:password@ part. The API key is sent as a Bearer token from the profile or --api-key, never in the URL.")
	}
	if parsed.Scheme != "https" && !(insecure && parsed.Scheme == "http") {
		return nil, apperr.New(apperr.ExitUsage, "Lago API requires HTTPS", "Use HTTPS or pass --insecure explicitly for a trusted self-hosted HTTP instance.")
	}
	if suggestion, isApp := APIHostFor(parsed.Hostname()); isApp {
		return nil, apperr.New(apperr.ExitUsage,
			fmt.Sprintf("%s is the Lago dashboard, not the Lago API", parsed.Hostname()),
			fmt.Sprintf("Use https://%s instead. The API and the app are different hosts, and an API key sent to the app host will not authenticate.", suggestion))
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	// Strip any /api/v1 the operator already pasted, then append exactly one. Stripping
	// in a loop covers the doubled form a previous version of this function produced and
	// wrote into people's config files.
	path := strings.TrimRight(parsed.Path, "/")
	for {
		trimmed, found := strings.CutSuffix(path, APIPrefix)
		if !found {
			break
		}
		path = strings.TrimRight(trimmed, "/")
	}
	parsed.Path = path + APIPrefix
	return parsed, nil
}

// APIHostFor reports the api.* host to use when hostname is a Lago app host, so the
// error can name the right one instead of just refusing.
//
// Lago's own dashboards are `app.getlago.com` and `app.eu.getlago.com`. Self-hosted
// deployments overwhelmingly follow the same `app.` / `api.` split, so the rule is the
// prefix rather than a list of two. A single-domain self-hosted deployment is untouched:
// its host does not start with `app.`.
func APIHostFor(hostname string) (string, bool) {
	lower := strings.ToLower(hostname)
	if rest, isApp := strings.CutPrefix(lower, "app."); isApp && rest != "" {
		return "api." + rest, true
	}
	return "", false
}

func (c *Client) Do(ctx context.Context, request Request) (*Response, error) {
	started := time.Now()
	method := strings.ToUpper(request.Method)
	target, err := c.resolve(request.Path, request.Query)
	if err != nil {
		return nil, err
	}
	headers := request.Headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}
	if headers.Get("Accept") == "" {
		headers.Set("Accept", "application/json")
	}
	if len(request.Body) > 0 && headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", "application/json")
	}
	if c.config.UserAgent != "" {
		headers.Set("User-Agent", c.config.UserAgent)
	}
	if c.config.APIKey != "" {
		headers.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	if request.DryRun {
		visibleHeaders := c.redactedHeaders(headers)
		var body any
		if len(request.Body) > 0 {
			redactedBody := c.redactor.String(string(request.Body))
			decoder := json.NewDecoder(strings.NewReader(redactedBody))
			decoder.UseNumber()
			if err := decoder.Decode(&body); err != nil {
				body = redactedBody
			}
		}
		return &Response{DryRunData: map[string]any{"method": method, "url": c.redactedURL(target), "headers": visibleHeaders, "body": body}}, nil
	}

	maxAttempts := 3
	if c.config.NoRetry || !request.Idempotent {
		maxAttempts = 1
	}
	var totalTiming Timing
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attemptTiming := Timing{}
		httpRequest, err := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(request.Body))
		if err != nil {
			return nil, apperr.Wrap(apperr.ExitUsage, "build HTTP request", err)
		}
		httpRequest.Header = cloneHeaders(headers)
		traceSnapshot := attachTrace(httpRequest)
		if c.config.Verbose {
			fmt.Fprintf(c.config.Err, "> %s %s\n", method, c.redactedURL(target))
			fmt.Fprintln(c.config.Err, formatHeaders(c.redactedHeaders(httpRequest.Header)))
			if len(request.Body) > 0 {
				fmt.Fprintln(c.config.Err, c.redactor.String(string(request.Body)))
			}
		}
		attemptStart := time.Now()
		httpResponse, doErr := c.http.Do(httpRequest)
		attemptTiming.RoundTrip = time.Since(attemptStart)
		if doErr != nil {
			mergeTiming(&totalTiming, attemptTiming)
			if attempt < maxAttempts && retryableNetwork(doErr) {
				wait := backoff(attempt)
				totalTiming.RetryWait += wait
				if err := waitContext(ctx, wait); err != nil {
					return failedResponse(started, totalTiming, attempt), networkError(err)
				}
				continue
			}
			return failedResponse(started, totalTiming, attempt), networkError(doErr)
		}
		downloadStart := time.Now()
		limitedBody := &io.LimitedReader{R: httpResponse.Body, N: (64 << 20) + 1}
		body, readErr := io.ReadAll(limitedBody)
		closeErr := httpResponse.Body.Close()
		attemptTiming.Download = time.Since(downloadStart)
		attemptTiming.ResponseBytes = len(body)
		traceTiming := traceSnapshot()
		attemptTiming.DNS = traceTiming.DNS
		attemptTiming.Connect = traceTiming.Connect
		attemptTiming.TLS = traceTiming.TLS
		attemptTiming.TimeToFirst = traceTiming.TimeToFirst
		mergeTiming(&totalTiming, attemptTiming)
		if readErr != nil {
			if attempt < maxAttempts && retryableNetwork(readErr) {
				wait := backoff(attempt)
				totalTiming.RetryWait += wait
				if err := waitContext(ctx, wait); err != nil {
					return failedResponse(started, totalTiming, attempt), networkError(err)
				}
				continue
			}
			return failedResponse(started, totalTiming, attempt), networkError(fmt.Errorf("read response: %w", readErr))
		}
		if limitedBody.N == 0 {
			return failedResponse(started, totalTiming, attempt), apperr.New(apperr.ExitGeneral, "response exceeds the 64 MiB safety limit", "Use a paginated or streaming endpoint instead of buffering this response.")
		}
		if closeErr != nil {
			return failedResponse(started, totalTiming, attempt), networkError(fmt.Errorf("close response: %w", closeErr))
		}
		if c.config.Verbose {
			fmt.Fprintf(c.config.Err, "< %d %s\n", httpResponse.StatusCode, httpResponse.Status)
			fmt.Fprintln(c.config.Err, formatHeaders(c.redactedHeaders(httpResponse.Header)))
			if len(body) > 0 {
				fmt.Fprintln(c.config.Err, c.redactor.String(string(body)))
			}
		}
		if attempt < maxAttempts && retryableStatus(httpResponse.StatusCode) {
			wait := retryAfter(httpResponse.Header.Get("Retry-After"))
			if wait == 0 {
				wait = backoff(attempt)
			}
			totalTiming.RetryWait += wait
			if err := waitContext(ctx, wait); err != nil {
				return failedResponse(started, totalTiming, attempt), networkError(err)
			}
			continue
		}
		response := &Response{
			Status:    httpResponse.StatusCode,
			Headers:   httpResponse.Header.Clone(),
			Body:      body,
			RequestID: firstHeader(httpResponse.Header, "X-Request-Id", "X-Lago-Request-Id"),
			Attempts:  attempt,
			Timing:    totalTiming,
		}
		response.Timing.Total = time.Since(started)
		response.Timing.AttemptCount = attempt
		if httpResponse.StatusCode >= 400 {
			return response, responseError(response, request.Subjects)
		}
		return response, nil
	}
	return failedResponse(started, totalTiming, maxAttempts), apperr.New(apperr.ExitGeneral, "request failed without a response", "Retry the command with --verbose.")
}

// failedResponse is the partial response returned alongside a terminal network error.
// It carries no status or body, only the attempt count and the timing measured so far,
// so a caller can still print `--timing` for the request that failed: that is when the
// latency breakdown is worth having.
func failedResponse(started time.Time, timing Timing, attempt int) *Response {
	timing.Total = time.Since(started)
	timing.AttemptCount = attempt
	return &Response{Attempts: attempt, Timing: timing}
}

func (c *Client) redactedHeaders(headers http.Header) http.Header {
	result := cloneHeaders(headers)
	for name, values := range result {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "authorization") {
			result[name] = []string{"Bearer " + redact.Replacement}
			continue
		}
		if strings.Contains(lower, "api-key") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") {
			result[name] = []string{redact.Replacement}
			continue
		}
		for index, value := range values {
			values[index] = c.redactor.String(value)
		}
	}
	return result
}

func (c *Client) redactedURL(target *url.URL) string {
	visible := *target
	query := visible.Query()
	for name, values := range query {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "key") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") {
			query[name] = []string{redact.Replacement}
			continue
		}
		for index, value := range values {
			values[index] = c.redactor.String(value)
		}
	}
	visible.RawQuery = query.Encode()
	return visible.Redacted()
}

func (c *Client) resolve(path string, query url.Values) (*url.URL, error) {
	result := *c.baseURL
	cleanPath := strings.TrimSpace(path)
	if strings.HasPrefix(cleanPath, "http://") || strings.HasPrefix(cleanPath, "https://") {
		return nil, apperr.New(apperr.ExitUsage, "absolute API paths are not allowed", "Use a relative path and select the server with --api-url.")
	}
	result.Path = strings.TrimRight(c.baseURL.Path, "/") + "/" + strings.TrimLeft(NormalizeRequestPath(cleanPath), "/")
	result.RawQuery = query.Encode()
	return &result, nil
}

// NormalizeRequestPath strips a redundant /api/v1 prefix from a request path.
//
// The base URL already ends in /api/v1, so `lago api GET /api/v1/customers` -- the form
// anyone copying from the API reference will type -- would otherwise request
// /api/v1/api/v1/customers and 404. No Lago endpoint lives under a second /api/v1, so
// the prefix is unambiguously a paste artefact.
func NormalizeRequestPath(path string) string {
	trimmed := "/" + strings.TrimLeft(path, "/")
	if trimmed == APIPrefix {
		return "/"
	}
	if rest, found := strings.CutPrefix(trimmed, APIPrefix+"/"); found {
		return "/" + rest
	}
	return path
}

func DecodeJSON(data []byte) (any, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return map[string]any{}, nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func responseError(response *Response, subjects []Subject) error {
	var envelope struct {
		Status  int            `json:"status"`
		Error   string         `json:"error"`
		Code    string         `json:"code"`
		Details map[string]any `json:"error_details"`
	}
	_ = json.Unmarshal(response.Body, &envelope)
	message := envelope.Error
	if message == "" {
		message = strings.TrimSpace(string(response.Body))
	}
	if message == "" {
		message = http.StatusText(response.Status)
	}
	exitCode, suggestion := classify(response.Status)
	if response.Status == http.StatusNotFound {
		if described, ok := describeNotFound(envelope.Code, subjects); ok {
			message = described
			suggestion = "Check that each identifier is the right kind for the flag it was passed to: a plan code is not a subscription external ID, and a Lago ID is not an external ID."
		}
	}
	return &apperr.Error{ExitCode: exitCode, Status: response.Status, Code: envelope.Code, Message: message, Details: envelope.Details, RequestID: response.RequestID, Suggestion: suggestion}
}

// describeNotFound turns a 404 into a message naming the resource type and, when the
// request carried it, the value, so a 404 never reads as an empty result.
//
// The Lago code is the authority on what was not found: `add_on_not_found` on an
// `invoices create` means the add-on code in the body, not the customer in the path.
// QA run 3 hit exactly that, and the old subject-only message blamed the customer.
// When the code names a resource, the message uses it and picks the matching subject
// if one exists; a code that names no subject is reported without a value rather than
// with the wrong one. Without a usable code, the subjects alone describe the request.
func describeNotFound(code string, subjects []Subject) (string, bool) {
	resource := notFoundResource(code)
	if resource != "" {
		for _, subject := range subjects {
			if normalizeKind(subject.Kind) == resource {
				return fmt.Sprintf("no %s %q exists in this organization", subject.Kind, subject.Value), true
			}
		}
		return "no matching " + strings.ReplaceAll(resource, "_", " ") + " exists in this organization", true
	}
	if len(subjects) == 0 {
		return "", false
	}
	parts := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		parts = append(parts, fmt.Sprintf("%s %q", subject.Kind, subject.Value))
	}
	if len(parts) == 1 {
		return "no " + parts[0] + " exists in this organization", true
	}
	return "not found: " + strings.Join(parts, ", ") + " (one of these does not exist in this organization)", true
}

// notFoundResource extracts the resource from a Lago `<resource>_not_found` code, or
// returns "" when the code has another shape.
func notFoundResource(code string) string {
	resource, found := strings.CutSuffix(code, "_not_found")
	if !found || resource == "" {
		return ""
	}
	return resource
}

// normalizeKind maps a subject kind as the CLI names it ("add on", "billable metric")
// onto the snake_case resource a Lago code uses ("add_on", "billable_metric").
func normalizeKind(kind string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(kind)), " ", "_")
}

func classify(status int) (int, string) {
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return apperr.ExitAuth, "Run `lago init` or check LAGO_API_KEY and its permissions."
	case status == http.StatusNotFound:
		return apperr.ExitNotFound, "Check the resource identifier and active profile."
	case status == http.StatusTooManyRequests:
		return apperr.ExitRateLimit, "Wait for Retry-After or retry with a lower request rate."
	case status >= 500:
		return apperr.ExitServer, "Retry later; use --verbose and include the request ID when contacting support."
	default:
		return apperr.ExitValidation, "Check the command flags and Lago API validation details."
	}
}

func networkError(err error) error {
	return &apperr.Error{ExitCode: apperr.ExitNetwork, Message: err.Error(), Suggestion: "Check DNS, TLS, network access, and --timeout; run `lago doctor` for diagnostics.", Cause: err}
}

func retryableStatus(status int) bool { return status == http.StatusTooManyRequests || status >= 500 }

func retryableNetwork(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, io.ErrUnexpectedEOF)
}

func backoff(attempt int) time.Duration {
	capDuration := 200 * time.Millisecond * time.Duration(1<<(attempt-1))
	if capDuration > 2*time.Second {
		capDuration = 2 * time.Second
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(capDuration)+1))
	if err != nil {
		return capDuration / 2
	}
	return time.Duration(value.Int64())
}

func retryAfter(value string) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if date, err := http.ParseTime(value); err == nil {
		if wait := time.Until(date); wait > 0 {
			return wait
		}
	}
	return 0
}

func waitContext(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func sameOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) && strings.EqualFold(left.Host, right.Host)
}

func cloneHeaders(headers http.Header) http.Header {
	if headers == nil {
		return make(http.Header)
	}
	return headers.Clone()
}

func firstHeader(headers http.Header, names ...string) string {
	for _, name := range names {
		if value := headers.Get(name); value != "" {
			return value
		}
	}
	return ""
}

func formatHeaders(headers http.Header) string {
	var builder strings.Builder
	for key, values := range headers {
		for _, value := range values {
			fmt.Fprintf(&builder, "%s: %s\n", key, value)
		}
	}
	return strings.TrimSpace(builder.String())
}

func attachTrace(request *http.Request) func() Timing {
	var mutex sync.Mutex
	var timing Timing
	var dnsStart, connectStart, tlsStart, wroteRequest time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			mutex.Lock()
			dnsStart = time.Now()
			mutex.Unlock()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			mutex.Lock()
			defer mutex.Unlock()
			if !dnsStart.IsZero() {
				timing.DNS += time.Since(dnsStart)
			}
		},
		ConnectStart: func(_, _ string) {
			mutex.Lock()
			connectStart = time.Now()
			mutex.Unlock()
		},
		ConnectDone: func(_, _ string, _ error) {
			mutex.Lock()
			defer mutex.Unlock()
			if !connectStart.IsZero() {
				timing.Connect += time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() {
			mutex.Lock()
			tlsStart = time.Now()
			mutex.Unlock()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			mutex.Lock()
			defer mutex.Unlock()
			if !tlsStart.IsZero() {
				timing.TLS += time.Since(tlsStart)
			}
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			mutex.Lock()
			wroteRequest = time.Now()
			mutex.Unlock()
		},
		GotFirstResponseByte: func() {
			mutex.Lock()
			defer mutex.Unlock()
			if !wroteRequest.IsZero() {
				timing.TimeToFirst += time.Since(wroteRequest)
			}
		},
	}
	*request = *request.WithContext(httptrace.WithClientTrace(request.Context(), trace))
	return func() Timing {
		mutex.Lock()
		defer mutex.Unlock()
		return timing
	}
}

func mergeTiming(total *Timing, attempt Timing) {
	total.DNS += attempt.DNS
	total.Connect += attempt.Connect
	total.TLS += attempt.TLS
	total.TimeToFirst += attempt.TimeToFirst
	total.Download += attempt.Download
	total.RoundTrip += attempt.RoundTrip
	total.ResponseBytes += attempt.ResponseBytes
}
