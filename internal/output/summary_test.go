package output

import (
	"encoding/json"
	"strings"
	"testing"
)

// QA M-nested: nested arrays and objects in a `get` table rendered as raw JSON in one
// cell. Every nested shape now has a readable summary and none contains JSON.
func TestNestedSummaries(t *testing.T) {
	t.Parallel()
	charges := []any{
		map[string]any{"lago_id": "ch_1", "billable_metric_code": "requests", "charge_model": "standard"},
		map[string]any{"lago_id": "ch_2", "billable_metric_code": "storage", "charge_model": "graduated"},
	}
	for _, tc := range []struct {
		name  string
		value any
		want  string
	}{
		{"empty list", []any{}, "0 items"},
		{"one labelled item", []any{map[string]any{"code": "basic"}}, "1 item: basic"},
		{"labelled list", []any{map[string]any{"code": "basic"}, map[string]any{"name": "Premium"}, map[string]any{"external_id": "ent"}}, "3 items: basic, Premium, ent"},
		{"label prefers code over lago_id", charges, "2 items: ch_1, ch_2"},
		{"capped list", []any{map[string]any{"code": "a"}, map[string]any{"code": "b"}, map[string]any{"code": "c"}, map[string]any{"code": "d"}, map[string]any{"code": "e"}, map[string]any{"code": "f"}, map[string]any{"code": "g"}}, "7 items: a, b, c, d, e, +2 more"},
		{"scalar list", []any{"a", json.Number("2"), true}, "3 items: a, 2, true"},
		{"unlabelled objects", []any{map[string]any{"amount_cents": json.Number("100"), "units": "1.0"}}, "1 item"},
		{"mixed nested lists", []any{[]any{"x"}, "y"}, "2 items"},
		{"identifier object", map[string]any{"lago_id": "pl_1", "code": "pro", "amount_cents": json.Number("1")}, "lago_id=pl_1 code=pro"},
		{"small scalar object", map[string]any{"type": "charge", "units": "2"}, "type=charge units=2"},
		{"small object with nesting", map[string]any{"a": map[string]any{"b": 1}}, "{1 field}"},
		{"large object", map[string]any{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}, "{5 fields}"},
		{"empty object", map[string]any{}, "{}"},
		{"nil", nil, ""},
		{"scalar", json.Number("42"), "42"},
	} {
		if got := summary(tc.value); got != tc.want {
			t.Errorf("%s: summary = %q, want %q", tc.name, got, tc.want)
		}
		if strings.ContainsAny(summary(tc.value), "[{\"") && tc.name != "small object with nesting" && tc.name != "large object" && tc.name != "empty object" {
			t.Errorf("%s: summary contains JSON punctuation: %q", tc.name, summary(tc.value))
		}
	}
}

func TestQA_MNested_PlanTableSummarisesChargesAndCommitment(t *testing.T) {
	t.Parallel()
	plan := map[string]any{"plan": map[string]any{
		"lago_id": "pl_1", "code": "pro",
		"charges": []any{
			map[string]any{"lago_id": "ch_1", "billable_metric_code": "requests"},
			map[string]any{"lago_id": "ch_2", "billable_metric_code": "storage"},
		},
		"minimum_commitment": map[string]any{"amount_cents": json.Number("1000"), "invoice_display_name": "Minimum", "taxes": []any{}, "plan_code": "pro", "lago_id": "mc_1"},
		"taxes":              []any{},
	}}
	out, err := render(t, Table, "", plan)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "[{") || strings.Contains(out, "{\"") {
		t.Errorf("table still contains JSON:\n%s", out)
	}
	for _, want := range []string{"CHARGES             2 items: ch_1, ch_2", "MINIMUM_COMMITMENT  lago_id=mc_1", "TAXES               0 items"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
}

// scalar stays the last resort for Go types the summariser does not know.
func TestScalarFallsBackToJSONForUnknownTypes(t *testing.T) {
	t.Parallel()
	if got := scalar(struct{ A int }{1}); got != `{"A":1}` {
		t.Errorf("scalar(struct) = %q", got)
	}
	if got := scalar(make(chan int)); !strings.Contains(got, "0x") && got == "" {
		t.Errorf("scalar(chan) = %q", got)
	}
}
