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
