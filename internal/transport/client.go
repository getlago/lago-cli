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

const (
	USAPI = "https://api.getlago.com/api/v1"
	EUAPI = "https://api.eu.getlago.com/api/v1"
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
}

type Client struct {
	baseURL  *url.URL
	http     *http.Client
	config   Config
	redactor *redact.Redactor
}

type Request struct {
	Method     string
	Path       string
	Query      url.Values
	Headers    http.Header
	Body       []byte
	Idempotent bool
	DryRun     bool
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
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
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

func NormalizeBaseURL(raw string, insecure bool) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, apperr.New(apperr.ExitAuth, "Lago API URL is not configured", "Run `lago init` or set LAGO_API_URL.")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, apperr.New(apperr.ExitUsage, "invalid Lago API URL", "Use an absolute URL such as https://api.getlago.com/api/v1.")
	}
	if parsed.Scheme != "https" && !(insecure && parsed.Scheme == "http") {
		return nil, apperr.New(apperr.ExitUsage, "Lago API requires HTTPS", "Use HTTPS or pass --insecure explicitly for a trusted self-hosted HTTP instance.")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	path := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(path, "/api/v1") {
		path += "/api/v1"
	}
	parsed.Path = path
	return parsed, nil
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
			if attempt < maxAttempts && retryableNetwork(doErr) {
				wait := backoff(attempt)
				totalTiming.RetryWait += wait
				if err := waitContext(ctx, wait); err != nil {
					return nil, networkError(err)
				}
				continue
			}
			return nil, networkError(doErr)
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
					return nil, networkError(err)
				}
				continue
			}
			return nil, networkError(fmt.Errorf("read response: %w", readErr))
		}
		if limitedBody.N == 0 {
			return nil, apperr.New(apperr.ExitGeneral, "response exceeds the 64 MiB safety limit", "Use a paginated or streaming endpoint instead of buffering this response.")
		}
		if closeErr != nil {
			return nil, networkError(fmt.Errorf("close response: %w", closeErr))
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
				return nil, networkError(err)
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
			return response, responseError(response)
		}
		return response, nil
	}
	return nil, apperr.New(apperr.ExitGeneral, "request failed without a response", "Retry the command with --verbose.")
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
	result.Path = strings.TrimRight(c.baseURL.Path, "/") + "/" + strings.TrimLeft(cleanPath, "/")
	result.RawQuery = query.Encode()
	return &result, nil
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

func responseError(response *Response) error {
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
	return &apperr.Error{ExitCode: exitCode, Status: response.Status, Code: envelope.Code, Message: message, Details: envelope.Details, RequestID: response.RequestID, Suggestion: suggestion}
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
