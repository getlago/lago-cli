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
			value: map[string]any{"invoices": []any{map[string]any{"created_at": "2026-01-01", "total_amount_cents": json.Number("100"), "lago_id": "inv_1", "currency": "EUR"}}},
			want:  []string{"LAGO_ID", "TOTAL_AMOUNT_CENTS", "inv_1", "100", "EUR"},
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
			name:  "nested objects are summarised, not dropped",
			value: map[string]any{"fee": map[string]any{"item": map[string]any{"type": "charge"}}},
			want:  []string{"ITEM", "type=charge"},
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
// more than one key must not be unwrapped, or fields would silently disappear. The one
// exception is the list envelope, one array beside `meta`, handled by unwrapList.
func TestOnlySingleKeyEnvelopesAreUnwrapped(t *testing.T) {
	t.Parallel()
	out, err := render(t, Table, "", map[string]any{"invoice": map[string]any{"lago_id": "inv_1"}, "meta": map[string]any{"page": json.Number("1")}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "META") || !strings.Contains(out, "INVOICE") {
		t.Errorf("an object beside meta was unwrapped and lost a field:\n%s", out)
	}
	out, err = render(t, Table, "", map[string]any{"a": []any{"x"}, "b": []any{"y"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("two arrays were unwrapped and one was lost:\n%s", out)
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

// The terse identifier block is the default output of every create and update. It must
// print the identity fields in a fixed order and nothing else, so an operator can pipe
// the ID straight into the next command.
func TestIdentifierRendererPrintsOnlyIdentity(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		value  any
		want   []string
		absent []string
	}{
		{
			name: "created resource keeps identity in a fixed order",
			value: map[string]any{"customer": map[string]any{
				"name": "Example", "external_id": "cus_1", "lago_id": "1a90",
				"currency": "USD", "created_at": "2026-09-01T10:00:00Z",
			}},
			want:   []string{"LAGO_ID", "1a90", "EXTERNAL_ID", "cus_1", "NAME", "Example"},
			absent: []string{"CURRENCY", "USD", "CREATED_AT"},
		},
		{
			name:   "status is an identifier of the new state",
			value:  map[string]any{"invoice": map[string]any{"lago_id": "inv_1", "status": "finalized", "total_amount_cents": json.Number("100")}},
			want:   []string{"LAGO_ID", "inv_1", "STATUS", "finalized"},
			absent: []string{"TOTAL_AMOUNT_CENTS"},
		},
		{
			name:   "code is printed for resources with no external ID",
			value:  map[string]any{"plan": map[string]any{"lago_id": "2b90", "code": "quickstart", "amount_cents": json.Number("0")}},
			want:   []string{"LAGO_ID", "CODE", "quickstart"},
			absent: []string{"AMOUNT_CENTS"},
		},
		{
			name:   "a null or blank identifier is omitted rather than printed empty",
			value:  map[string]any{"subscription": map[string]any{"lago_id": "3c90", "external_id": "sub_1", "name": nil, "code": "   "}},
			want:   []string{"LAGO_ID", "EXTERNAL_ID"},
			absent: []string{"NAME", "CODE"},
		},
		{
			name:   "no identifiers falls back to the full table",
			value:  map[string]any{"customer": map[string]any{"sequential_id": json.Number("7"), "slug": "EXA-001"}},
			want:   []string{"SEQUENTIAL_ID", "7", "SLUG", "EXA-001"},
			absent: []string{},
		},
		{
			name:   "a collection reduces to identity columns",
			value:  map[string]any{"customers": []any{map[string]any{"lago_id": "1a90", "external_id": "cus_1", "currency": "USD"}}},
			want:   []string{"LAGO_ID", "EXTERNAL_ID", "cus_1"},
			absent: []string{"CURRENCY", "USD"},
		},
		{
			name:   "a collection of scalars is passed through, not reduced",
			value:  map[string]any{"codes": []any{"first", "second"}},
			want:   []string{"first", "second"},
			absent: []string{},
		},
		{
			name:   "a collection whose rows carry no identifier keeps every column",
			value:  map[string]any{"fees": []any{map[string]any{"amount_cents": json.Number("500"), "units": "3"}}},
			want:   []string{"AMOUNT_CENTS", "500", "UNITS"},
			absent: []string{},
		},
		{
			name:   "an empty collection still says so",
			value:  map[string]any{"customers": []any{}},
			want:   []string{"No results."},
			absent: []string{},
		},
		{
			name:   "a bare scalar response is passed through",
			value:  "just-a-string",
			want:   []string{"just-a-string"},
			absent: []string{},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var buffer bytes.Buffer
			err := Renderer{Mode: Table, Out: &buffer, Identifiers: true}.Render(testCase.value)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}
			out := buffer.String()
			for _, want := range testCase.want {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q:\n%s", want, out)
				}
			}
			for _, absent := range testCase.absent {
				if strings.Contains(out, absent) {
					t.Errorf("output should not contain %q:\n%s", absent, out)
				}
			}
		})
	}
}

// The terse default is a table concern only. JSON and YAML always carry the complete
// resource, because that is the machine-readable contract scripts depend on.
func TestIdentifiersNeverReduceStructuredOutput(t *testing.T) {
	t.Parallel()
	value := map[string]any{"customer": map[string]any{"lago_id": "1a90", "currency": "USD"}}
	for _, mode := range []string{JSON, YAML} {
		var buffer bytes.Buffer
		if err := (Renderer{Mode: mode, Out: &buffer, Identifiers: true}).Render(value); err != nil {
			t.Fatalf("%s render failed: %v", mode, err)
		}
		if !strings.Contains(buffer.String(), "currency") || !strings.Contains(buffer.String(), "USD") {
			t.Errorf("%s output was reduced to identifiers:\n%s", mode, buffer.String())
		}
	}
}

// The identifier order is part of the documented contract: lago_id first, so the field
// an operator most often pipes onward is the first line of output.
func TestIdentifierOrderIsStable(t *testing.T) {
	t.Parallel()
	if got := identifierKeys; len(got) != 5 || got[0] != "lago_id" || got[1] != "external_id" || got[2] != "code" || got[3] != "name" || got[4] != "status" {
		t.Fatalf("identifier order changed to %v; update the README and snapshots deliberately", got)
	}
}

// A query that matches nothing must print a hint naming the keys that were available,
// keep null on stdout, and stay quiet whenever there is nothing to explain.
func TestEmptyMatchHint(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name      string
		value     any
		query     string
		wantOut   string
		wantHint  []string
		quietHint bool
	}{
		{
			name:     "wrapper omitted",
			value:    map[string]any{"customers": []any{map[string]any{"lago_id": "cus_1"}}, "meta": map[string]any{}},
			query:    "lago_id",
			wantOut:  "null",
			wantHint: []string{"query matched nothing", "customers", "meta"},
		},
		{
			name:      "successful query stays quiet",
			value:     map[string]any{"customers": []any{map[string]any{"lago_id": "cus_1"}}},
			query:     "customers[].lago_id",
			wantOut:   "cus_1",
			quietHint: true,
		},
		{
			name:      "a null response is an answer, not a missed match",
			value:     nil,
			query:     "anything",
			wantOut:   "null",
			quietHint: true,
		},
		{
			name:      "an empty list is a match, not a miss",
			value:     map[string]any{"customers": []any{}},
			query:     "customers",
			wantOut:   "[]",
			quietHint: true,
		},
		{
			name:     "a non-object response still gets the bare hint",
			value:    []any{"first"},
			query:    "missing",
			wantOut:  "null",
			wantHint: []string{"query matched nothing"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var out, errOut bytes.Buffer
			err := Renderer{Mode: JSON, Query: testCase.query, Out: &out, Err: &errOut}.Render(testCase.value)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}
			if !strings.Contains(out.String(), testCase.wantOut) {
				t.Errorf("stdout = %q, want it to contain %q", out.String(), testCase.wantOut)
			}
			if testCase.quietHint {
				if strings.Contains(errOut.String(), "matched nothing") {
					t.Errorf("hint fired when it should not have: %q", errOut.String())
				}
				return
			}
			for _, want := range testCase.wantHint {
				if !strings.Contains(errOut.String(), want) {
					t.Errorf("hint missing %q: %q", want, errOut.String())
				}
			}
		})
	}
}

// A renderer with no Err writer must not panic when a query matches nothing.
func TestEmptyMatchHintWithNoErrorWriterIsDiscarded(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := (Renderer{Mode: JSON, Query: "missing", Out: &out}).Render(map[string]any{"customers": []any{}}); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if strings.TrimSpace(out.String()) != "null" {
		t.Errorf("stdout = %q, want null", out.String())
	}
}
