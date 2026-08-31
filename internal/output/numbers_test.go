package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// compact re-encodes indented renderer output so tests compare values, not layout.
func compact(t *testing.T, raw []byte) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(buffer.String())
}

func decodeExact(t *testing.T, raw string) any {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

const invoicesFixture = `{"invoices":[
  {"lago_id":"a","amount_cents":150000,"currency":"EUR"},
  {"lago_id":"b","amount_cents":50,"currency":"EUR"}
]}`

func TestNumericQueriesMatchDecodedNumbers(t *testing.T) {
	t.Parallel()
	value := decodeExact(t, invoicesFixture)
	for _, testCase := range []struct{ query, want string }{
		{"invoices[?amount_cents > `1000`].lago_id", `["a"]`},
		{"invoices[?amount_cents < `1000`].lago_id", `["b"]`},
		{"sum(invoices[].amount_cents)", `150050`},
		{"max_by(invoices, &amount_cents).lago_id", `"a"`},
		{"length(invoices)", `2`},
	} {
		var buffer bytes.Buffer
		renderer := Renderer{Mode: JSON, Query: testCase.query, Out: &buffer}
		if err := renderer.Render(value); err != nil {
			t.Errorf("query %s failed: %v", testCase.query, err)
			continue
		}
		if got := compact(t, buffer.Bytes()); got != testCase.want {
			t.Errorf("query %s = %s, want %s", testCase.query, got, testCase.want)
		}
	}
}

// A number too large for exact float64 must be refused, never silently rounded.
func TestQueryRefusesValuesItCannotRepresentExactly(t *testing.T) {
	t.Parallel()
	value := decodeExact(t, `{"invoices":[{"amount_cents":9007199254740993}]}`)
	var buffer bytes.Buffer
	renderer := Renderer{Mode: JSON, Query: "invoices[0].amount_cents", Out: &buffer}
	err := renderer.Render(value)
	if err == nil {
		t.Fatal("query silently rounded a value float64 cannot represent")
	}
	if !strings.Contains(err.Error(), "losing precision") {
		t.Errorf("refusal does not name the cause: %v", err)
	}
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("refusal is not a typed CLI error: %v", err)
	}
	if appErr.ExitCode != apperr.ExitUsage {
		t.Errorf("refusal exit code = %d, want %d", appErr.ExitCode, apperr.ExitUsage)
	}
	if !strings.Contains(appErr.Suggestion, "jq") {
		t.Errorf("refusal does not teach the alternative: %s", appErr.Suggestion)
	}
}

// JSON and YAML must agree that a monetary field is a number, with identical digits.
func TestJSONAndYAMLAgreeOnNumericTypes(t *testing.T) {
	t.Parallel()
	value := decodeExact(t, `{"amount_cents":150000,"big_cents":9007199254740993,"unit":1.5,"code":"EUR"}`)

	var jsonBuffer, yamlBuffer bytes.Buffer
	if err := (Renderer{Mode: JSON, Out: &jsonBuffer}).Render(value); err != nil {
		t.Fatal(err)
	}
	if err := (Renderer{Mode: YAML, Out: &yamlBuffer}).Render(value); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"amount_cents: 150000", "big_cents: 9007199254740993", "unit: 1.5", "code: EUR"} {
		if !strings.Contains(yamlBuffer.String(), want) {
			t.Errorf("YAML missing %q:\n%s", want, yamlBuffer.String())
		}
	}
	for _, unwanted := range []string{`"150000"`, `"9007199254740993"`, `"1.5"`} {
		if strings.Contains(yamlBuffer.String(), unwanted) {
			t.Errorf("YAML quoted a number (%s):\n%s", unwanted, yamlBuffer.String())
		}
	}
	if !strings.Contains(jsonBuffer.String(), `"big_cents": 9007199254740993`) {
		t.Errorf("JSON lost exact digits:\n%s", jsonBuffer.String())
	}
}
