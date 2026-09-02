package generated

import "testing"

// The vocabulary is the fallback classification for requests no operation claims, so
// its matching must be case-insensitive and substring-based, and a plain read must
// never match.
func TestMatchesDestructiveVocabulary(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"/invoices/x/void", "FINALIZE", "retry_payment", "terminate", "destroy-plan-charge", "refund"} {
		if !MatchesDestructiveVocabulary(value) {
			t.Errorf("%q should be destructive", value)
		}
	}
	for _, value := range []string{"", "/customers", "/customers/{id}", "create", "list-applied", "/plans/x/charges"} {
		if MatchesDestructiveVocabulary(value) {
			t.Errorf("%q should not be destructive", value)
		}
	}
	if len(DestructiveVocabulary) != 7 {
		t.Errorf("vocabulary has %d entries, want 7 (keep internal/contract/parity_test.go in sync)", len(DestructiveVocabulary))
	}
}
