package apperr

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// The exit codes are a public contract: scripts branch on them. Freeze the values
// so a reordering of the const block cannot silently change what `lago` returns.
func TestExitCodeContractIsFrozen(t *testing.T) {
	t.Parallel()
	for name, expected := range map[string]int{
		"success": ExitSuccess, "general": ExitGeneral, "usage": ExitUsage,
		"auth": ExitAuth, "notFound": ExitNotFound, "validation": ExitValidation,
		"rateLimit": ExitRateLimit, "server": ExitServer, "network": ExitNetwork,
	} {
		if expected < 0 || expected > 8 {
			t.Errorf("%s exit code %d is outside the published 0-8 range", name, expected)
		}
	}
	frozen := []int{ExitSuccess, ExitGeneral, ExitUsage, ExitAuth, ExitNotFound,
		ExitValidation, ExitRateLimit, ExitServer, ExitNetwork}
	for index, code := range frozen {
		if code != index {
			t.Fatalf("exit code %d moved to position %d; the contract is frozen", code, index)
		}
	}
}

// A nil *Error must not panic when used as an error value.
func TestNilErrorIsSafe(t *testing.T) {
	t.Parallel()
	var err *Error
	if err.Error() != "" {
		t.Fatalf("nil error rendered %q", err.Error())
	}
}

func TestNewCarriesCodeMessageAndSuggestion(t *testing.T) {
	t.Parallel()
	err := New(ExitValidation, "amount_cents must be positive", "Pass a positive integer.")
	if err.ExitCode != ExitValidation {
		t.Errorf("exit code = %d", err.ExitCode)
	}
	if err.Error() != "amount_cents must be positive" {
		t.Errorf("message = %q", err.Error())
	}
	if err.Suggestion != "Pass a positive integer." {
		t.Errorf("suggestion = %q", err.Suggestion)
	}
	if ExitCode(err) != ExitValidation {
		t.Errorf("ExitCode() = %d", ExitCode(err))
	}
}

// Every error surfaced to a script must carry a suggestion and, when the API
// provided one, a request ID. Encode is what --output json emits.
func TestEncodeIncludesDiagnosticFields(t *testing.T) {
	t.Parallel()
	err := &Error{
		ExitCode:   ExitValidation,
		Status:     422,
		Code:       "value_is_invalid",
		Message:    "amount_cents must be positive",
		RequestID:  "req_abc123",
		Suggestion: "Pass a positive integer.",
		Details:    map[string]any{"field": "amount_cents"},
	}
	var payload struct {
		Error struct {
			Status     int            `json:"status"`
			Code       string         `json:"code"`
			Message    string         `json:"message"`
			RequestID  string         `json:"request_id"`
			Suggestion string         `json:"suggestion"`
			Details    map[string]any `json:"details"`
		} `json:"error"`
	}
	if decodeErr := json.Unmarshal(Encode(err), &payload); decodeErr != nil {
		t.Fatalf("Encode produced invalid JSON: %v", decodeErr)
	}
	if payload.Error.Status != 422 || payload.Error.Code != "value_is_invalid" {
		t.Errorf("status/code lost: %+v", payload.Error)
	}
	if payload.Error.RequestID != "req_abc123" {
		t.Errorf("request ID lost: %+v", payload.Error)
	}
	if payload.Error.Suggestion == "" || payload.Error.Details["field"] != "amount_cents" {
		t.Errorf("suggestion or details lost: %+v", payload.Error)
	}
	// ExitCode is deliberately not serialised: it is the process exit status.
	if strings.Contains(string(Encode(err)), "ExitCode") {
		t.Error("Encode leaked the internal ExitCode field")
	}
}

// A plain error still encodes as the same JSON shape, so consumers can parse one
// schema regardless of where the failure came from.
func TestEncodeWrapsPlainErrors(t *testing.T) {
	t.Parallel()
	payload := Encode(errors.New("connection reset by peer"))
	var decoded map[string]map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Encode produced invalid JSON: %v", err)
	}
	if decoded["error"]["message"] != "connection reset by peer" {
		t.Fatalf("message lost: %s", payload)
	}
	if ExitCode(errors.New("boom")) != ExitGeneral {
		t.Error("an untyped error must map to the general exit code")
	}
	if ExitCode(nil) != ExitSuccess {
		t.Error("nil must map to success")
	}
}

// An *Error with no explicit exit code must not report success.
func TestZeroExitCodeFallsBackToGeneral(t *testing.T) {
	t.Parallel()
	if got := ExitCode(&Error{Message: "no code set"}); got != ExitGeneral {
		t.Fatalf("ExitCode = %d, want %d", got, ExitGeneral)
	}
}

// errors.Is/As must reach through Wrap so callers can classify the cause.
func TestWrapIsTransparentToErrorsPackage(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("dial tcp: i/o timeout")
	wrapped := Wrap(ExitNetwork, "reach the Lago API", sentinel)
	if !errors.Is(wrapped, sentinel) {
		t.Fatal("errors.Is cannot reach the cause")
	}
	if ExitCode(wrapped) != ExitNetwork {
		t.Fatalf("ExitCode = %d", ExitCode(wrapped))
	}
	var target *Error
	if !errors.As(wrapped, &target) || target.Message != "reach the Lago API" {
		t.Fatal("errors.As did not recover the typed error")
	}
}
