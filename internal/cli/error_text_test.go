package cli

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// runAgainst422 executes argv in default (table) output mode against a server that
// answers every request with the given 422 envelope, returning stderr and the exit code.
func runAgainst422(t *testing.T, envelope string, argv ...string) (string, int) {
	t.Helper()
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("X-Request-Id", "req_42")
		response.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = response.Write([]byte(envelope))
	})
	profileAt(t, server.URL)
	var stdout, stderr bytes.Buffer
	code := ExecuteArgs(argv, strings.NewReader(""), &stdout, &stderr)
	return stderr.String(), code
}

// assertDetailLine checks that a `field: reason` line sits between the Error line and
// the Lago code line, where an operator reading top-down finds it first.
func assertDetailLine(t *testing.T, stderr, line string) {
	t.Helper()
	errorAt := strings.Index(stderr, "Error: ")
	detailAt := strings.Index(stderr, "  "+line+"\n")
	codeAt := strings.Index(stderr, "Lago code: ")
	if errorAt < 0 || detailAt < 0 || codeAt < 0 {
		t.Fatalf("stderr lacks the error, detail, or code line:\n%s", stderr)
	}
	if !(errorAt < detailAt && detailAt < codeAt) {
		t.Errorf("detail line is out of order:\n%s", stderr)
	}
}

// QA E-5c: an event with a malformed timestamp came back as "Unprocessable Entity,
// check the command flags" with no hint which flag. The value sent here passes the
// CLI's own timestamp parsing (Unix seconds); what is under test is how a server-side
// rejection is printed, so the mocked API answers 422 regardless of the value.
func TestQA_E5c_TimestampInvalidFormatIsPrinted(t *testing.T) {
	stderr, code := runAgainst422(t, `{"status":422,"error":"Unprocessable Entity","code":"validation_errors","error_details":{"timestamp":["invalid_format"]}}`,
		"events", "send", "--code", "requests", "--external-subscription-id", "sub_1", "--timestamp", "1788338088")
	if code != apperr.ExitValidation {
		t.Fatalf("exit code = %d, want %d", code, apperr.ExitValidation)
	}
	assertDetailLine(t, stderr, "timestamp: invalid_format")
}

// QA L-2d: creating a customer with an existing code.
func TestQA_L2d_CodeValueAlreadyExistsIsPrinted(t *testing.T) {
	stderr, _ := runAgainst422(t, `{"status":422,"error":"Unprocessable Entity","code":"validation_errors","error_details":{"code":["value_already_exist"]}}`,
		"customers", "create", "--input", `{"customer":{"external_id":"dup","name":"Dup"}}`)
	assertDetailLine(t, stderr, "code: value_already_exist")
}

// QA L-3d: a credit note over the remaining fee amount.
func TestQA_L3d_AmountCentsHigherThanRemainingFeeAmountIsPrinted(t *testing.T) {
	stderr, _ := runAgainst422(t, `{"status":422,"error":"Unprocessable Entity","code":"validation_errors","error_details":{"amount_cents":["higher_than_remaining_fee_amount"]}}`,
		"api", "POST", "/credit_notes", "--data", `{"credit_note":{}}`)
	assertDetailLine(t, stderr, "amount_cents: higher_than_remaining_fee_amount")
}

// QA L-3d: a refund with no payment provider linked reports on `base`.
func TestQA_L3d_BaseNoLinkedPaymentProviderIsPrinted(t *testing.T) {
	stderr, _ := runAgainst422(t, `{"status":422,"error":"Unprocessable Entity","code":"validation_errors","error_details":{"base":["no_linked_payment_provider"]}}`,
		"api", "POST", "/credit_notes", "--data", `{"credit_note":{}}`)
	assertDetailLine(t, stderr, "base: no_linked_payment_provider")
}

// The text error keeps everything it printed before, in a fixed order, and adds the
// HTTP status. --output json is unchanged and covered by TestJSONErrorEnvelopeIsStable.
func TestTextErrorKeepsStatusCodeRequestIDAndSuggestion(t *testing.T) {
	stderr, _ := runAgainst422(t, `{"status":422,"error":"Unprocessable Entity","code":"validation_errors","error_details":{"code":["value_already_exist"],"amount_cents":["invalid"]}}`,
		"api", "POST", "/plans", "--data", `{"plan":{}}`)
	want := "Error: Unprocessable Entity\n" +
		"  amount_cents: invalid\n" +
		"  code: value_already_exist\n" +
		"HTTP status: 422\n" +
		"Lago code: validation_errors\n" +
		"Request ID: req_42\n" +
		"Suggestion: Check the command flags and Lago API validation details.\n"
	if !strings.HasSuffix(stderr, want) {
		t.Errorf("stderr =\n%s\nwant suffix:\n%s", stderr, want)
	}
}

func TestFormatErrorDetailsShapes(t *testing.T) {
	t.Parallel()
	lines := formatErrorDetails(map[string]any{
		"timestamp": []any{"invalid_format"},
		"code":      []any{"value_already_exist", "too_long"},
		"base":      "no_linked_payment_provider",
		"nested":    map[string]any{"field": "reason"},
		"empty":     nil,
	})
	want := []string{
		"base: no_linked_payment_provider",
		"code: value_already_exist, too_long",
		`nested: {"field":"reason"}`,
		"timestamp: invalid_format",
	}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("lines = %q, want %q", lines, want)
	}
	if formatErrorDetails(nil) != nil {
		t.Error("nil details produced lines")
	}
}

// A plain error that is not an *apperr.Error still prints one Error line.
func TestWriteTextErrorHandlesPlainErrors(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	writeTextError(&buffer, io.ErrUnexpectedEOF)
	if buffer.String() != "Error: unexpected EOF\n" {
		t.Errorf("plain error printed as %q", buffer.String())
	}
}
