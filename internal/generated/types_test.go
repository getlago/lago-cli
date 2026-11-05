package generated

import "testing"

func TestPinnedOperationsLoaded(t *testing.T) {
	t.Parallel()
	if SpecVersion == "" || len(SpecSHA256) != 64 || len(Operations) == 0 {
		t.Fatalf("spec=%s sha=%s operations=%d", SpecVersion, SpecSHA256, len(Operations))
	}
}
