package redact

import (
	"strings"
	"testing"
)

func TestRedactorRemovesKnownSecretAndCredentialShapes(t *testing.T) {
	t.Parallel()
	secret := "lago_test_FAKEabcdefghijklmnopqrstuv"
	input := "Authorization: Bearer " + secret + ` {"api_key":"another-secret","password":"hunter2"} token=` + secret
	output := New(secret).String(input)
	for _, forbidden := range []string{secret, "another-secret", "hunter2"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("redacted output leaked %q: %s", forbidden, output)
		}
	}
	if strings.Count(output, Replacement) < 3 {
		t.Fatalf("expected redaction markers in %q", output)
	}
}

func FuzzRedactorNeverLeaksKnownSecret(f *testing.F) {
	f.Add("prefix", "suffix")
	f.Fuzz(func(t *testing.T, prefix, suffix string) {
		secret := "lago_test_FAKEabcdefghijklmnopqrstuv"
		output := New(secret).String(prefix + secret + suffix)
		if strings.Contains(output, secret) {
			t.Fatal("known secret leaked")
		}
	})
}
