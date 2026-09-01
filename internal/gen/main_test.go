package main

import "testing"

func TestCommandNamingHelpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, got, want string
	}{
		{"snake case", normalizeName("billable_metrics"), "billable-metrics"},
		{"camel case", normalizeName("invoicePreview"), "invoice-preview"},
		{"plural", singular("customers"), "customer"},
		{"summary", derivedAction("createEvent", "events", "/events", "Send usage events"), "send"},
		{"custom path", derivedAction("invoicePreview", "invoices", "/invoices/preview", "Create an invoice preview"), "preview"},
		{"qualified", qualifiedAction("findAllAppliedCoupons", "coupons"), "list-applied"},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %q, want %q", test.name, test.got, test.want)
		}
	}
}

func TestGeneratorRejectsMissingPaths(t *testing.T) {
	t.Parallel()
	if _, err := (generator{document: map[string]any{}}).operations(); err == nil {
		t.Fatal("missing paths unexpectedly generated")
	}
}

// The terse-output classification is a generator rule, so it is tested here rather than
// re-asserted at 48 call sites. The excluded cases are the ones that would be wrong to
// reduce: read-shaped mutations, state transitions, and bulk ingestion summaries.
func TestMutationClassification(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		method, action string
		want           bool
	}{
		{"POST", "create", true},
		{"PUT", "update", true},
		{"PATCH", "update", true},
		{"POST", "create-plan-charge", true},
		{"PUT", "update-subscription-alert", true},
		{"PATCH", "update-subscription", true},

		{"GET", "create", false}, // a read is never terse, whatever it is called
		{"DELETE", "delete", false},
		{"POST", "preview", false}, // invoices preview: the body is the answer
		{"POST", "estimate", false},
		{"POST", "estimate-fees", false},
		{"POST", "send", false},    // events send: the summary is the answer
		{"PUT", "finalize", false}, // the new state is the answer
		{"POST", "void", false},
		{"POST", "execute", false},
		{"POST", "download", false},
		{"POST", "checkout-url", false},
		{"PUT", "refresh", false},
		{"POST", "created", false}, // prefix match must be on a segment boundary
		{"POST", "updates", false},
	} {
		if got := mutationOperation(test.method, test.action); got != test.want {
			t.Errorf("mutationOperation(%q, %q) = %v, want %v", test.method, test.action, got, test.want)
		}
	}
}
