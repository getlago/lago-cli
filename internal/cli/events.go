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
	"sync"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type eventJob struct {
	line int
	body []byte
	key  string
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
				headers := make(http.Header)
				headers.Set("Idempotency-Key", job.key)
				response, requestErr := client.Do(cmd.Context(), transport.Request{Method: http.MethodPost, Path: "/events", Headers: headers, Body: job.body, Idempotent: true, DryRun: app.dryRun})
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
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), 10<<20)
	for line := 1; scanner.Scan(); line++ {
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}
		body, key, encodeErr := prepareEvent(raw)
		if encodeErr != nil {
			close(jobs)
			for range results {
			}
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid event on line %d: %v", line, encodeErr), "Use one JSON event object per line; an optional top-level event wrapper is accepted.")
		}
		select {
		case jobs <- eventJob{line: line, body: body, key: key}:
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

func prepareEvent(raw []byte) ([]byte, string, error) {
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, "", err
	}
	event := payload
	if wrapped, ok := payload["event"].(map[string]any); ok {
		event = wrapped
	} else {
		payload = map[string]any{"event": event}
	}
	key, _ := event["transaction_id"].(string)
	if key == "" {
		key = uuid.NewString()
		event["transaction_id"] = key
	}
	encoded, err := json.Marshal(payload)
	return encoded, key, err
}
