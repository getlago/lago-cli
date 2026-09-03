package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type eventJob struct {
	line int
	body []byte
	// retryable is true when the event carries a timestamp, so a replay with the same
	// transaction_id and timestamp is collapsed by the server rather than billed twice.
	retryable bool
}

type eventResult struct {
	line    int
	retried int
	err     error
}

func runEventStream(cmd *cobra.Command, app *App, source string, concurrency int) error {
	if concurrency < 1 || concurrency > 64 {
		return apperr.New(apperr.ExitUsage, "concurrency must be between 1 and 64", "Pass --concurrency 4 or another value in the allowed range.")
	}
	reader, closeReader, err := eventReader(app.In, source)
	if err != nil {
		return err
	}
	defer closeReader()
	client, err := app.Client(true)
	if err != nil {
		return err
	}
	if app.resolved.Profile.Mode == config.ModeLive {
		fmt.Fprintf(app.Err, "[LIVE] profile=%s; streaming usage events\n", app.resolved.Name)
	}

	jobs := make(chan eventJob, concurrency*2)
	results := make(chan eventResult, concurrency*2)
	var workers sync.WaitGroup
	for range concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				response, requestErr := client.Do(cmd.Context(), transport.Request{Method: http.MethodPost, Path: "/events", Body: job.body, Idempotent: job.retryable, DryRun: app.dryRun})
				retried := 0
				if response != nil && response.Attempts > 1 {
					retried = response.Attempts - 1
				}
				results <- eventResult{line: job.line, retried: retried, err: requestErr}
			}
		}()
	}
	go func() {
		workers.Wait()
		close(results)
	}()

	queued := 0
	retryUnsafe := 0
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), 10<<20)
	for line := 1; scanner.Scan(); line++ {
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}
		body, retryable, unsafe, encodeErr := prepareEvent(raw)
		if unsafe {
			retryUnsafe++
		}
		if encodeErr != nil {
			close(jobs)
			for range results {
			}
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid event on line %d: %v", line, encodeErr), "Use one JSON event object per line; an optional top-level event wrapper is accepted.")
		}
		select {
		case jobs <- eventJob{line: line, body: body, retryable: retryable}:
			queued++
		case <-cmd.Context().Done():
			close(jobs)
			for range results {
			}
			return apperr.Wrap(apperr.ExitNetwork, "event stream cancelled", cmd.Context().Err())
		}
	}
	scanErr := scanner.Err()
	close(jobs)
	if retryUnsafe > 0 {
		fmt.Fprintf(app.Err, "WARNING: %d event(s) carry a transaction_id but no timestamp and are not safe to resend: on the ClickHouse event store a missing timestamp defaults to the time of reception and is part of the deduplication key, so a retry bills them again. Add a timestamp (unix seconds) to every event.\n", retryUnsafe)
	}

	sent, failed, retried := 0, 0, 0
	failures := make([]map[string]any, 0, 5)
	firstExitCode := apperr.ExitValidation
	for result := range results {
		retried += result.retried
		if result.err == nil {
			sent++
			continue
		}
		failed++
		if len(failures) < 5 {
			failures = append(failures, map[string]any{"line": result.line, "error": result.err.Error()})
		}
		if failed == 1 {
			firstExitCode = apperr.ExitCode(result.err)
		}
	}
	if scanErr != nil {
		return apperr.Wrap(apperr.ExitUsage, "read event stream", scanErr)
	}
	summary := map[string]any{"queued": queued, "sent": sent, "failed": failed, "retried": retried, "dry_run": app.dryRun}
	if len(failures) > 0 {
		summary["failures"] = failures
	}
	if err := app.Renderer().Render(summary); err != nil {
		return err
	}
	if failed > 0 {
		return &apperr.Error{ExitCode: firstExitCode, Message: fmt.Sprintf("%d of %d events failed", failed, queued), Suggestion: "Fix the reported lines and resend them with the same transaction IDs."}
	}
	return nil
}

func eventReader(stdin io.Reader, source string) (io.Reader, func(), error) {
	if source == "-" {
		return stdin, func() {}, nil
	}
	file, err := os.Open(filepath.Clean(source))
	if err != nil {
		return nil, func() {}, apperr.Wrap(apperr.ExitGeneral, "open event stream", err)
	}
	return file, func() { _ = file.Close() }, nil
}

// prepareEvent wraps a raw NDJSON line into the API envelope and assigns a transaction
// ID when the line has none. retryable reports whether the transport may replay the
// event on 429/5xx: only when it carries a timestamp, because transaction_id plus
// timestamp is the server-side deduplication key. retryUnsafe reports whether the line
// carried its own transaction ID but no timestamp; see retryUnsafeWarning.
func prepareEvent(raw []byte) (body []byte, retryable bool, retryUnsafe bool, err error) {
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, false, false, err
	}
	event := payload
	if wrapped, ok := payload["event"].(map[string]any); ok {
		event = wrapped
	} else {
		payload = map[string]any{"event": event}
	}
	if timestamp, isText := event["timestamp"].(string); isText {
		normalized, err := normalizeEventTimestamp(timestamp)
		if err != nil {
			return nil, false, false, fmt.Errorf("timestamp: %w", err)
		}
		event["timestamp"] = normalized
	} else if err := checkNumericEventTimestamp(event["timestamp"]); err != nil {
		return nil, false, false, fmt.Errorf("timestamp: %w", err)
	}
	if transactionID, _ := event["transaction_id"].(string); transactionID == "" {
		event["transaction_id"] = uuid.NewString()
	} else {
		retryUnsafe = eventIsRetryUnsafe(event)
	}
	encoded, err := json.Marshal(payload)
	return encoded, eventHasTimestamp(event), retryUnsafe, err
}

// eventHasTimestamp reports whether an event pins its own timestamp instead of leaving
// it to the time of reception.
func eventHasTimestamp(event map[string]any) bool {
	timestamp, present := event["timestamp"]
	return present && timestamp != nil && timestamp != ""
}

// isEventTimestamp reports whether a generated body field is the event timestamp, the
// one field whose spec type (`oneOf [integer, string]`, Unix seconds) the CLI knows well
// enough to accept a calendar instant for.
func isEventTimestamp(wrapper string, path []string) bool {
	return wrapper == "event" && len(path) == 1 && path[0] == "timestamp"
}

// Bounds for a plausible Unix-seconds event timestamp. Below the floor (2001-09-09) a
// number is almost certainly not an epoch: `2026` typed as a year is 33 minutes after
// 1970 and would be billed nowhere. At or above the ceiling (year 5138) the value is a
// millisecond epoch pasted from JavaScript or a database, and sending it unchanged
// silently stores the event 3000 years out, outside every billing period.
const (
	minEventTimestampSeconds = 1_000_000_000
	maxEventTimestampSeconds = 100_000_000_000
)

var unixSecondsPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// normalizeEventTimestamp accepts plausible Unix seconds (with optional decimals)
// unchanged, and converts an RFC 3339 instant to Unix seconds, keeping sub-second
// precision as decimals. Lago stores the timestamp in UTC; a zone offset in the input
// is honoured, not dropped. Numbers are digits only: signs, exponents, hex and Inf/NaN
// are not timestamps even though strconv would parse them.
func normalizeEventTimestamp(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if unixSecondsPattern.MatchString(trimmed) {
		return checkUnixSeconds(trimmed)
	}
	instant, err := time.Parse(time.RFC3339Nano, trimmed)
	if err != nil {
		return "", fmt.Errorf("%q is neither Unix seconds nor an RFC 3339 instant", raw)
	}
	seconds := strconv.FormatInt(instant.Unix(), 10)
	if nanos := instant.Nanosecond(); nanos > 0 {
		fraction := strings.TrimRight(fmt.Sprintf("%09d", nanos), "0")
		return seconds + "." + fraction, nil
	}
	return seconds, nil
}

// checkUnixSeconds accepts a digits-only value when it sits inside the plausible
// window, and otherwise says what the number most likely is.
func checkUnixSeconds(digits string) (string, error) {
	seconds, err := strconv.ParseFloat(digits, 64)
	if err != nil {
		return "", fmt.Errorf("%q is not a Unix timestamp", digits)
	}
	switch {
	case seconds >= maxEventTimestampSeconds:
		return "", fmt.Errorf("%s looks like a millisecond epoch (year %d as seconds); divide by 1000 or pass an RFC 3339 instant", digits, time.Unix(int64(seconds), 0).UTC().Year())
	case seconds < minEventTimestampSeconds:
		return "", fmt.Errorf("%s is too small for Unix seconds (%s); pass Unix seconds since 1970 or an RFC 3339 instant", digits, time.Unix(int64(seconds), 0).UTC().Format(time.RFC3339))
	}
	return digits, nil
}

// checkNumericEventTimestamp applies the same window to a timestamp that arrived as a
// JSON number in an --input body or an NDJSON line, where no string normalization runs.
func checkNumericEventTimestamp(value any) error {
	number, ok := value.(json.Number)
	if !ok {
		return nil
	}
	_, err := checkUnixSeconds(number.String())
	return err
}

// normalizeEventBodyTimestamp applies normalizeEventTimestamp to a complete --input
// body for a single `events send`, so the two ways of passing an event agree.
func normalizeEventBodyTimestamp(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return body, nil // the API reports a malformed body with its own message
	}
	event, ok := payload["event"].(map[string]any)
	if !ok {
		return body, nil
	}
	timestamp, isText := event["timestamp"].(string)
	if !isText {
		if err := checkNumericEventTimestamp(event["timestamp"]); err != nil {
			return nil, apperr.New(apperr.ExitUsage, "invalid event timestamp: "+err.Error(), "Pass Unix seconds (1788338088, optionally with decimals) or an RFC 3339 instant (2026-09-02T09:30:00Z).")
		}
		return body, nil
	}
	normalized, err := normalizeEventTimestamp(timestamp)
	if err != nil {
		return nil, apperr.New(apperr.ExitUsage, "invalid event timestamp: "+err.Error(), "Pass Unix seconds (1788338088, optionally with decimals) or an RFC 3339 instant (2026-09-02T09:30:00Z).")
	}
	if normalized == timestamp {
		return body, nil
	}
	event["timestamp"] = normalized
	return json.Marshal(payload)
}

// eventIsRetryUnsafe reports whether an event names its own transaction_id but leaves
// timestamp to the server.
//
// Lago deduplicates events on transaction_id, but on the ClickHouse event store the
// timestamp is part of the deduplication key, and a missing timestamp defaults to the
// time of reception. So two sends of the same command are two distinct events and both
// are billed. Only a caller-chosen transaction_id is a promise of idempotency the CLI
// can break; a generated one is unique per run and a resend is a new event by design.
func eventIsRetryUnsafe(event map[string]any) bool {
	transactionID, _ := event["transaction_id"].(string)
	return transactionID != "" && !eventHasTimestamp(event)
}

const retryUnsafeWarning = "WARNING: --transaction-id without --timestamp is not safe to retry. Lago sets a missing timestamp to the time of reception, and on the ClickHouse event store the timestamp is part of the deduplication key, so resending this command bills a second event. Pass --timestamp <unix seconds> and resend the same value on retry."

// warnRetryUnsafeBody inspects a single-event request body and prints
// retryUnsafeWarning on stderr when it applies. It never fails the command: the
// request is valid, only the retry story is not.
func warnRetryUnsafeBody(app *App, body []byte) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return
	}
	event, ok := payload["event"].(map[string]any)
	if !ok {
		event = payload
	}
	if eventIsRetryUnsafe(event) {
		fmt.Fprintln(app.Err, retryUnsafeWarning)
	}
}
