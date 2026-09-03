package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// QA M-empty: `events send` printed five blank rows because the endpoint is async and
// answers with most fields null. Null and blank values are omitted from a key/value
// table; when every value is blank the keys still print.
func TestQA_MEmpty_NilAndBlankFieldsAreDropped(t *testing.T) {
	t.Parallel()
	out, err := render(t, Table, "", map[string]any{"event": map[string]any{
		"lago_id": "e1", "code": "requests", "external_customer_id": nil, "transaction_id": "t1",
		"precise_total_amount_cents": "   ", "timestamp": nil,
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, absent := range []string{"EXTERNAL_CUSTOMER_ID", "PRECISE_TOTAL_AMOUNT_CENTS", "TIMESTAMP"} {
		if strings.Contains(out, absent) {
			t.Errorf("blank field %s was printed:\n%s", absent, out)
		}
	}
	if strings.Count(out, "\n") != 3 {
		t.Errorf("want 3 rows (code, lago_id, transaction_id):\n%s", out)
	}
	out, err = render(t, Table, "", map[string]any{"event": map[string]any{"a": nil, "b": ""}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("an all-blank object printed nothing:\n%q", out)
	}
}

// QA M-empty: a terse array response renders one row per item, with status.
func TestQA_MEmpty_TerseArrayRendersOneRowPerItemWithStatus(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	value := map[string]any{"wallet_transactions": []any{
		map[string]any{"lago_id": "wt_1", "status": "settled", "amount": "10.0", "transaction_status": "purchased"},
		map[string]any{"lago_id": "wt_2", "status": "pending", "amount": "5.0"},
	}}
	if err := (Renderer{Mode: Table, Out: &buffer, Identifiers: true}).Render(value); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want header + 2 rows:\n%s", buffer.String())
	}
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "LAGO_ID STATUS" {
		t.Errorf("header = %q", got)
	}
	if !strings.Contains(lines[2], "wt_2") || !strings.Contains(lines[2], "pending") || strings.Contains(buffer.String(), "AMOUNT") {
		t.Errorf("rows:\n%s", buffer.String())
	}
}

func TestWritePairsDropsBlanksAndSanitizes(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	err := WritePairs(&buffer, []Pair{{Key: "name", Value: "Acme\x1b[0m"}, {Key: "timezone", Value: ""}, {Key: "mode", Value: "test"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "NAME  Acme\\x1b[0m\nMODE  test\n"
	if buffer.String() != want {
		t.Errorf("pairs = %q, want %q", buffer.String(), want)
	}
	_ = json.Number("0")
}
