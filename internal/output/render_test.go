package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func render(t *testing.T, mode, query string, value any) (string, error) {
	t.Helper()
	var buffer bytes.Buffer
	err := Renderer{Mode: mode, Query: query, Out: &buffer}.Render(value)
	return buffer.String(), err
}

// Table output is what an operator reads. It must put identity and money columns
// first, render every scalar kind, and never panic on a shape it did not expect.
func TestTableRendersEveryValueShape(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name  string
		value any
		want  []string
	}{
		{
			name:  "collection puts identity and amount first",
			value: map[string]any{"invoices": []any{map[string]any{"created_at": "2026-01-01", "amount_cents": json.Number("100"), "lago_id": "inv_1", "currency": "EUR"}}},
			want:  []string{"LAGO_ID", "AMOUNT_CENTS", "inv_1", "100", "EUR"},
		},
		{
			name:  "single object renders as key/value rows",
			value: map[string]any{"customer": map[string]any{"name": "Example", "external_id": "cus_1"}},
			want:  []string{"NAME", "Example", "EXTERNAL_ID", "cus_1"},
		},
		{
			name:  "empty collection says so rather than printing nothing",
			value: map[string]any{"invoices": []any{}},
			want:  []string{"No results."},
		},
		{
			name:  "scalar list renders one per line",
			value: []any{"first", "second"},
			want:  []string{"first", "second"},
		},
		{
			name:  "bare scalar renders alone",
			value: "just-a-string",
			want:  []string{"just-a-string"},
		},
		{
			name:  "nested objects are serialised, not dropped",
			value: map[string]any{"fee": map[string]any{"item": map[string]any{"type": "charge"}}},
			want:  []string{"charge"},
		},
		{
			name:  "booleans and nulls render without panicking",
			value: map[string]any{"invoice": map[string]any{"paid": true, "voided_at": nil, "attempts": json.Number("3")}},
			want:  []string{"PAID", "true", "ATTEMPTS", "3"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			out, err := render(t, Table, "", testCase.value)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}
			for _, want := range testCase.want {
				if !strings.Contains(out, want) {
					t.Errorf("table output missing %q:\n%s", want, out)
				}
			}
		})
	}
}

// An unwrapped single-key envelope is the Lago API's response shape. Objects with
// more than one key must not be unwrapped, or fields would silently disappear.
func TestOnlySingleKeyEnvelopesAreUnwrapped(t *testing.T) {
	t.Parallel()
	out, err := render(t, Table, "", map[string]any{"invoice": map[string]any{"lago_id": "inv_1"}, "meta": map[string]any{"page": json.Number("1")}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "META") || !strings.Contains(out, "INVOICE") {
		t.Errorf("a two-key response was unwrapped and lost a field:\n%s", out)
	}
}

// The default mode is table, and an unknown mode must be a usage error naming the
// valid choices rather than silently falling back.
func TestOutputModeSelection(t *testing.T) {
	t.Parallel()
	value := map[string]any{"customer": map[string]any{"lago_id": "cus_1"}}

	empty, err := render(t, "", "", value)
	if err != nil || !strings.Contains(empty, "cus_1") {
		t.Fatalf("empty mode did not default to table: %q %v", empty, err)
	}
	if _, err := render(t, "xml", "", value); err == nil {
		t.Fatal("unknown output mode was accepted")
	} else if !strings.Contains(err.Error(), "table") || !strings.Contains(err.Error(), "json") {
		t.Errorf("mode error does not name the valid choices: %v", err)
	}
}

// An invalid JMESPath expression must be a usage error that points at the flag.
func TestInvalidQueryIsAUsageError(t *testing.T) {
	t.Parallel()
	_, err := render(t, JSON, "invoices[?", map[string]any{"invoices": []any{}})
	if err == nil {
		t.Fatal("a malformed query was accepted")
	}
	if !strings.Contains(err.Error(), "JMESPath") {
		t.Errorf("query error does not name the cause: %v", err)
	}
}

// A renderer with no writer must not panic; it discards.
func TestNilWriterDiscards(t *testing.T) {
	t.Parallel()
	if err := (Renderer{Mode: JSON}).Render(map[string]any{"a": "b"}); err != nil {
		t.Fatalf("nil writer render failed: %v", err)
	}
}

// Deeply nested and repeated numbers must all convert, including inside arrays.
func TestNumberConversionReachesNestedValues(t *testing.T) {
	t.Parallel()
	value := decodeExact(t, `{"a":{"b":[{"cents":1},{"cents":2}]},"list":[3,4]}`)

	converted, err := queryValue(value)
	if err != nil {
		t.Fatalf("nested conversion failed: %v", err)
	}
	nested := converted.(map[string]any)["a"].(map[string]any)["b"].([]any)
	if _, ok := nested[0].(map[string]any)["cents"].(float64); !ok {
		t.Error("a number nested in an array of objects was not converted")
	}
	if _, ok := converted.(map[string]any)["list"].([]any)[0].(float64); !ok {
		t.Error("a number in a bare array was not converted")
	}

	yamlTree := yamlValue(value).(map[string]any)
	if _, ok := yamlTree["list"].([]any); !ok {
		t.Error("yamlValue did not preserve array structure")
	}
}

// An unrepresentable number anywhere in the tree must fail the whole query, and
// the message must name the field that caused it.
func TestUnrepresentableNumberNamesItsField(t *testing.T) {
	t.Parallel()
	value := decodeExact(t, `{"invoice":{"total_amount_cents":9007199254740993}}`)
	_, err := queryValue(value)
	if err == nil {
		t.Fatal("an unrepresentable number was silently converted")
	}
	for _, want := range []string{"invoice", "total_amount_cents"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not name %q: %v", want, err)
		}
	}
}

// Values that are exactly representable must pass, including negatives and decimals.
func TestExactlyRepresentableNumbersAreAccepted(t *testing.T) {
	t.Parallel()
	// Integers within float64's exact range, and any non-integer rate or percentage.
	for _, literal := range []string{"0", "-100", "1.5", "-0.25", "9007199254740992", "1e3", "0.1", "0.0001"} {
		if _, err := exactFloat(json.Number(literal)); err != nil {
			t.Errorf("%s was rejected: %v", literal, err)
		}
	}
	// Integer minor units beyond 2^53 are provably not the value the API returned.
	for _, literal := range []string{"9007199254740993", "-9007199254740993"} {
		if _, err := exactFloat(json.Number(literal)); err == nil {
			t.Errorf("%s was accepted despite not round-tripping through float64", literal)
		}
	}
	if _, err := exactFloat(json.Number("not-a-number")); err == nil {
		t.Error("a non-numeric literal was accepted")
	}
}
