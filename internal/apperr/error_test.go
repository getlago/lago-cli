package apperr

import (
	"errors"
	"strings"
	"testing"
)

func TestExitCodeAndEncoding(t *testing.T) {
	t.Parallel()
	err := &Error{ExitCode: ExitAuth, Status: 401, Code: "unauthorized", Message: "bad key", RequestID: "req_fake", Suggestion: "run init"}
	if got := ExitCode(err); got != ExitAuth {
		t.Fatalf("ExitCode = %d", got)
	}
	encoded := string(Encode(err))
	for _, expected := range []string{`"status": 401`, `"code": "unauthorized"`, `"request_id": "req_fake"`} {
		if !strings.Contains(encoded, expected) {
			t.Fatalf("encoding %q does not contain %q", encoded, expected)
		}
	}
	if ExitCode(errors.New("plain")) != ExitGeneral || ExitCode(nil) != ExitSuccess {
		t.Fatal("generic or nil exit-code mapping changed")
	}
}

func TestWrapPreservesCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	err := Wrap(ExitNetwork, "network", cause)
	if !errors.Is(err, cause) {
		t.Fatal("wrapped cause was not preserved")
	}
}
