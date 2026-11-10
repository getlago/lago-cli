package apperr

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	ExitSuccess    = 0
	ExitGeneral    = 1
	ExitUsage      = 2
	ExitAuth       = 3
	ExitNotFound   = 4
	ExitValidation = 5
	ExitRateLimit  = 6
	ExitServer     = 7
	ExitNetwork    = 8
)

type Error struct {
	ExitCode int            `json:"-"`
	Status   int            `json:"status,omitempty"`
	Code     string         `json:"code,omitempty"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details,omitempty"`
	// FieldErrors is the API v2 `details` array as the server sent it, one
	// `{field, code, meta}` entry each. Details above keeps the v1 field-to-codes shape
	// for both versions; this carries the meta (provider error codes, bounds, allowed
	// values) that the v1 shape has no room for.
	FieldErrors []map[string]any `json:"field_errors,omitempty"`
	RequestID   string           `json:"request_id,omitempty"`
	Suggestion  string           `json:"suggestion,omitempty"`
	Cause       error            `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func New(exitCode int, message, suggestion string) *Error {
	return &Error{ExitCode: exitCode, Message: message, Suggestion: suggestion}
}

func Wrap(exitCode int, message string, cause error) *Error {
	return &Error{ExitCode: exitCode, Message: message, Cause: cause}
}

func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var appErr *Error
	if errors.As(err, &appErr) && appErr.ExitCode != 0 {
		return appErr.ExitCode
	}
	return ExitGeneral
}

func Encode(err error) []byte {
	var appErr *Error
	if !errors.As(err, &appErr) {
		appErr = &Error{ExitCode: ExitGeneral, Message: err.Error()}
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{"error": appErr}, "", "  ")
	if marshalErr != nil {
		return []byte(fmt.Sprintf(`{"error":{"message":%q}}`, err.Error()))
	}
	return payload
}
