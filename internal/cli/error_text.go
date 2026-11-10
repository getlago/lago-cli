package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/output"
)

// writeTextError prints an error for a human in default (table) output mode.
//
// The Lago API answers a 422 with a bare "Unprocessable Entity" and puts the useful
// part, which field failed and why, in error_details. QA read "check the command flags"
// three times without learning which flag. So the field lines come straight after the
// message, indented so they read as part of it, followed by the HTTP status, the Lago
// code, the request ID for support, and the suggestion. --output json is unchanged and
// carries the same details under error.details.
func writeTextError(errOut io.Writer, err error) {
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		fmt.Fprintf(errOut, "Error: %s\n", output.Sanitize(err.Error()))
		return
	}
	// API-controlled text reaches the terminal here, so it takes the same escaping as a
	// table cell (QA S-22). See output.Sanitize.
	fmt.Fprintf(errOut, "Error: %s\n", output.Sanitize(appErr.Message))
	lines := formatFieldErrors(appErr.FieldErrors)
	if len(lines) == 0 {
		lines = formatErrorDetails(appErr.Details)
	}
	for _, line := range lines {
		fmt.Fprintf(errOut, "  %s\n", output.Sanitize(line))
	}
	if appErr.Status > 0 {
		fmt.Fprintf(errOut, "HTTP status: %d\n", appErr.Status)
	}
	if appErr.Code != "" {
		fmt.Fprintf(errOut, "Lago code: %s\n", output.Sanitize(appErr.Code))
	}
	if appErr.RequestID != "" {
		fmt.Fprintf(errOut, "Request ID: %s\n", output.Sanitize(appErr.RequestID))
	}
	if appErr.Suggestion != "" {
		fmt.Fprintf(errOut, "Suggestion: %s\n", output.Sanitize(appErr.Suggestion))
	}
}

// formatErrorDetails renders the API's error_details map as `field: reason` lines in a
// stable order. Lago sends each field's reasons as an array of codes, which are joined
// with commas; a scalar is printed as is and a nested object as compact JSON so nothing
// the server said is dropped.
func formatErrorDetails(details map[string]any) []string {
	if len(details) == 0 {
		return nil
	}
	fields := make([]string, 0, len(details))
	for field := range details {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		value := details[field]
		if value == nil {
			continue
		}
		var reason string
		switch typed := value.(type) {
		case []any:
			parts := make([]string, 0, len(typed))
			for _, item := range typed {
				parts = append(parts, fmt.Sprint(item))
			}
			reason = strings.Join(parts, ", ")
		case map[string]any:
			encoded, err := json.Marshal(typed)
			if err != nil {
				reason = fmt.Sprint(typed)
			} else {
				reason = string(encoded)
			}
		default:
			reason = fmt.Sprint(typed)
		}
		lines = append(lines, field+": "+reason)
	}
	return lines
}

// formatFieldErrors renders API v2 `{field, code, meta}` entries as `field: code` lines,
// in the order the server sent them, with any meta appended as sorted `key=value` pairs:
// `base: provider_error (provider_error_code=card_declined, provider_name=stripe)`.
func formatFieldErrors(entries []map[string]any) []string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		line := fmt.Sprintf("%v: %v", entry["field"], entry["code"])
		meta, _ := entry["meta"].(map[string]any)
		if len(meta) > 0 {
			keys := make([]string, 0, len(meta))
			for key := range meta {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			pairs := make([]string, 0, len(keys))
			for _, key := range keys {
				value := meta[key]
				if nested, ok := value.([]any); ok {
					parts := make([]string, 0, len(nested))
					for _, item := range nested {
						parts = append(parts, fmt.Sprint(item))
					}
					value = strings.Join(parts, "|")
				}
				pairs = append(pairs, fmt.Sprintf("%s=%v", key, value))
			}
			line += " (" + strings.Join(pairs, ", ") + ")"
		}
		lines = append(lines, line)
	}
	return lines
}
