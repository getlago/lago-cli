package generated

import "testing"

func TestPinnedOperationsLoaded(t *testing.T) {
	t.Parallel()
	if SpecVersion != "1.52.1" || len(SpecSHA256) != 64 || len(Operations) != 217 {
		t.Fatalf("spec=%s sha=%s operations=%d", SpecVersion, SpecSHA256, len(Operations))
	}
}
