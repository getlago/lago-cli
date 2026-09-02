package cli

import (
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/generated"
)

// A raw request is gated by what the spec says about the operation it addresses, not
// by the verb alone. Paths the spec does not know fall back to the generator's own
// default-deny rule so an unknown DELETE is never silently ungated.
func TestQA_S28_ClassifyRequestMatchesTemplatesAndFallsBackToVerb(t *testing.T) {
	t.Parallel()
	cases := []struct {
		method, path string
		dangerous    bool
		matched      bool
	}{
		{"DELETE", "/customers/abc", true, true},
		{"delete", "customers/abc/", true, true},
		{"DELETE", "/api/v1/customers/abc", true, true},
		{"POST", "/invoices/inv_1/void", true, true},
		{"POST", "/api/v1/invoices/inv_1/retry_payment", true, true},
		{"PUT", "/invoices/inv_1/finalize", true, true},
		{"GET", "/customers", false, true},
		{"GET", "/customers?page=2", false, true},
		{"POST", "/customers", false, true},
		{"PUT", "/subscriptions/sub_1", false, true},
		{"DELETE", "/unknown/thing", true, false},
		{"POST", "/unknown/thing", false, false},
		{"POST", "/unknown/thing/refund", true, false},
	}
	for _, tc := range cases {
		dangerous, operationID := classifyRequest(tc.method, tc.path)
		if dangerous != tc.dangerous {
			t.Errorf("%s %s: dangerous = %v, want %v", tc.method, tc.path, dangerous, tc.dangerous)
		}
		if (operationID != "") != tc.matched {
			t.Errorf("%s %s: matched operation %q, want matched=%v", tc.method, tc.path, operationID, tc.matched)
		}
	}
}

// Every generated operation, addressed by its own path with parameters filled in,
// must classify exactly as the generator classified it. This is what keeps `lago api`
// and fixture steps in lockstep with the generated commands.
func TestQA_S28_ClassifyRequestAgreesWithEveryGeneratedOperation(t *testing.T) {
	t.Parallel()
	for _, operation := range generated.Operations {
		path := operation.Path
		for strings.Contains(path, "{") {
			start := strings.Index(path, "{")
			end := strings.Index(path, "}")
			path = path[:start] + "fake_" + path[start+1:end] + path[end+1:]
		}
		dangerous, operationID := classifyRequest(operation.Method, path)
		if operationID == "" {
			t.Errorf("%s %s did not match any operation", operation.Method, path)
			continue
		}
		if dangerous != operation.Dangerous {
			t.Errorf("%s %s (%s): dangerous = %v, generator says %v", operation.Method, path, operationID, dangerous, operation.Dangerous)
		}
	}
}
