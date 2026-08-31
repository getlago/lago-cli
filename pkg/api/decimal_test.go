package api

import (
	"encoding/json"
	"testing"
)

func TestDecimalRoundTrip(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"0", "-42", "12.3400", "1e-9"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			decimal, err := NewDecimal(value)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(decimal)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != value {
				t.Fatalf("encoded %q, want %q", encoded, value)
			}
			var decoded Decimal
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != decimal {
				t.Fatalf("decoded %q, want %q", decoded, decimal)
			}
		})
	}
}

func TestDecimalRejectsInvalidValues(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "01", "NaN", "Inf", "1.2.3", "$12"} {
		if _, err := NewDecimal(value); err == nil {
			t.Errorf("NewDecimal(%q) unexpectedly succeeded", value)
		}
	}
	invalid := Decimal("NaN")
	if _, err := json.Marshal(invalid); err == nil {
		t.Fatal("marshalling invalid Decimal unexpectedly succeeded")
	}
}

func TestDecimalAcceptsQuotedAPIValueAndNull(t *testing.T) {
	t.Parallel()
	var decimal Decimal
	if err := json.Unmarshal([]byte(`"10.50"`), &decimal); err != nil {
		t.Fatal(err)
	}
	if decimal.String() != "10.50" {
		t.Fatalf("got %q", decimal)
	}
	if err := json.Unmarshal([]byte("null"), &decimal); err != nil {
		t.Fatal(err)
	}
}
