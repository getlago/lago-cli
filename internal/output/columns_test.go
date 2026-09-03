package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func headerLine(t *testing.T, value any) string {
	t.Helper()
	out, err := render(t, Table, "", value)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(strings.SplitN(out, "\n", 2)[0]), " ")
}

func TestColumnAllowlistOrdersKnownResources(t *testing.T) {
	t.Parallel()
	customer := map[string]any{"created_at": "2026-01-01", "currency": "USD", "email": "ops@acme.test", "name": "Acme", "external_id": "acme", "lago_id": "cus_1", "slug": "ACM-1", "timezone": "UTC"}
	invoice := map[string]any{"lago_id": "inv_1", "number": "LAG-1", "status": "finalized", "payment_status": "pending", "currency": "EUR", "total_amount_cents": json.Number("1200"), "issuing_date": "2026-01-01", "fees_amount_cents": json.Number("1000")}
	subscription := map[string]any{"lago_id": "sub_1", "external_id": "s1", "plan_code": "pro", "status": "active", "started_at": "2026-01-01", "ending_at": nil, "billing_time": "calendar"}
	plan := map[string]any{"lago_id": "pl_1", "code": "pro", "name": "Pro", "interval": "monthly", "amount_cents": json.Number("4900"), "amount_currency": "USD", "trial_period": json.Number("0")}
	for _, tc := range []struct {
		key  string
		row  map[string]any
		want string
	}{
		{"customers", customer, "LAGO_ID EXTERNAL_ID NAME EMAIL CURRENCY CREATED_AT"},
		{"invoices", invoice, "LAGO_ID NUMBER STATUS PAYMENT_STATUS CURRENCY TOTAL_AMOUNT_CENTS ISSUING_DATE"},
		{"subscriptions", subscription, "LAGO_ID EXTERNAL_ID PLAN_CODE STATUS STARTED_AT ENDING_AT"},
		{"plans", plan, "LAGO_ID CODE NAME INTERVAL AMOUNT_CENTS AMOUNT_CURRENCY"},
	} {
		if got := headerLine(t, map[string]any{tc.key: []any{tc.row}}); got != tc.want {
			t.Errorf("%s header = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestColumnAllowlistDropsAbsentColumns(t *testing.T) {
	t.Parallel()
	got := headerLine(t, map[string]any{"customers": []any{map[string]any{"lago_id": "cus_1", "name": "Acme"}}})
	if got != "LAGO_ID NAME" {
		t.Errorf("header = %q, want LAGO_ID NAME", got)
	}
}

func TestColumnAllowlistFallsBackWhenNothingMatches(t *testing.T) {
	t.Parallel()
	got := headerLine(t, map[string]any{"customers": []any{map[string]any{"foo": json.Number("1"), "bar": "x"}}})
	if got != "BAR FOO" {
		t.Errorf("header = %q, want the heuristic BAR FOO", got)
	}
}

// QA L-2h: the old renderer took columns from the first row only, so a field absent
// there vanished for the whole page.
func TestHeuristicColumnsUnionAcrossRows(t *testing.T) {
	t.Parallel()
	out, err := render(t, Table, "", map[string]any{"things": []any{
		map[string]any{"lago_id": "t1"},
		map[string]any{"lago_id": "t2", "status": "active"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "STATUS") || !strings.Contains(out, "active") {
		t.Errorf("column present only in a later row was dropped:\n%s", out)
	}
}

func TestHeuristicColumnsPairMoneyWithCurrency(t *testing.T) {
	t.Parallel()
	got := headerLine(t, map[string]any{"fees": []any{map[string]any{"zzz": "last", "currency": "EUR", "total_amount_cents": json.Number("1"), "amount_cents": json.Number("2"), "amount_currency": "USD", "lago_id": "f1", "created_at": "2026"}}})
	if got != "LAGO_ID AMOUNT_CENTS AMOUNT_CURRENCY TOTAL_AMOUNT_CENTS CURRENCY CREATED_AT ZZZ" {
		t.Errorf("header = %q", got)
	}
}

func TestHeuristicColumnCapIsEight(t *testing.T) {
	t.Parallel()
	row := map[string]any{}
	for _, key := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		row[key] = key
	}
	row["nested"] = map[string]any{"x": 1}
	got := headerLine(t, map[string]any{"things": []any{row}})
	if got != "A B C D E F G H" {
		t.Errorf("header = %q, want the first eight scalar columns", got)
	}
}

func TestResourceColumnsAccessorReturnsACopy(t *testing.T) {
	t.Parallel()
	copied := ResourceColumns()
	copied["customers"][0] = "mutated"
	if resourceColumns["customers"][0] != "lago_id" {
		t.Fatal("ResourceColumns exposed the internal slice")
	}
}
