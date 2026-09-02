package output

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func renderBoth(t *testing.T, renderer Renderer, value any) (string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	renderer.Out = &out
	renderer.Err = &errOut
	if err := renderer.Render(value); err != nil {
		t.Fatal(err)
	}
	return out.String(), errOut.String()
}

// QA E-4, M-list: every `list` printed the page as one JSON string in a single cell,
// because only one-key envelopes were unwrapped. The list envelope, one array beside
// `meta`, now renders one row per item with meta on stderr.
func TestQA_E4_ListWrapperWithMetaRendersRows(t *testing.T) {
	t.Parallel()
	page := func(totalPages int) map[string]any {
		return map[string]any{
			"customers": []any{
				map[string]any{"lago_id": "cus_1", "external_id": "acme", "name": "Acme"},
				map[string]any{"lago_id": "cus_2", "external_id": "globex", "name": "Globex"},
			},
			"meta": map[string]any{"current_page": json.Number("1"), "total_pages": json.Number(itoa(totalPages)), "total_count": json.Number("250")},
		}
	}
	out, errOut := renderBoth(t, Renderer{Mode: Table}, page(3))
	for _, want := range []string{"LAGO_ID", "EXTERNAL_ID", "NAME", "cus_1", "globex"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	for _, absent := range []string{"CUSTOMERS", "META", "[{", "total_pages"} {
		if strings.Contains(out, absent) {
			t.Errorf("stdout still carries %q:\n%s", absent, out)
		}
	}
	if rows := strings.Count(out, "\n"); rows != 3 {
		t.Errorf("rows = %d, want header + 2:\n%s", rows, out)
	}
	if errOut != "page 1 of 3 (250 total); use --page N or --all\n" {
		t.Errorf("pagination hint = %q", errOut)
	}

	if _, errOut := renderBoth(t, Renderer{Mode: Table}, page(1)); errOut != "" {
		t.Errorf("single page produced a hint: %q", errOut)
	}
	if _, errOut := renderBoth(t, Renderer{Mode: Table, AllPages: true}, page(3)); errOut != "" {
		t.Errorf("--all rendering produced a hint: %q", errOut)
	}
	noMeta := map[string]any{"customers": page(3)["customers"]}
	if out, errOut := renderBoth(t, Renderer{Mode: Table}, noMeta); !strings.Contains(out, "cus_2") || errOut != "" {
		t.Errorf("list without meta: out=%q err=%q", out, errOut)
	}
	empty := map[string]any{"customers": []any{}, "meta": map[string]any{"total_pages": json.Number("0")}}
	if out, _ := renderBoth(t, Renderer{Mode: Table}, empty); out != "No results.\n" {
		t.Errorf("empty list = %q", out)
	}
	if out, _ := renderBoth(t, Renderer{Mode: Table}, map[string]any{"customers": "not a list", "meta": map[string]any{}}); !strings.Contains(out, "CUSTOMERS") {
		t.Errorf("non-array beside meta must stay key/value:\n%s", out)
	}
	if out, _ := renderBoth(t, Renderer{Mode: Table}, map[string]any{"customers": []any{}, "meta": "bogus"}); !strings.Contains(out, "META") {
		t.Errorf("non-object meta must stay key/value:\n%s", out)
	}
	if out, _ := renderBoth(t, Renderer{Mode: Table}, map[string]any{"meta": map[string]any{"total_pages": json.Number("2")}}); !strings.Contains(out, "TOTAL_PAGES") {
		t.Errorf("a lone meta object is a plain object:\n%s", out)
	}
}

// The terse renderer reduces a list envelope too, instead of falling back to the
// old `WALLETS <json>` / `META` rows.
func TestQA_MList_TerseArrayUnwrapsThroughMeta(t *testing.T) {
	t.Parallel()
	value := map[string]any{
		"wallet_transactions": []any{
			map[string]any{"lago_id": "wt_1", "amount": "10.0", "status": "settled"},
			map[string]any{"lago_id": "wt_2", "amount": "5.0", "status": "pending"},
		},
		"meta": map[string]any{"total_pages": json.Number("1")},
	}
	out, _ := renderBoth(t, Renderer{Mode: Table, Identifiers: true}, value)
	if !strings.Contains(out, "LAGO_ID") || !strings.Contains(out, "wt_2") || strings.Contains(out, "META") || strings.Contains(out, "AMOUNT") {
		t.Errorf("terse list rendering:\n%s", out)
	}
	if rows := strings.Count(out, "\n"); rows != 3 {
		t.Errorf("rows = %d, want header + 2:\n%s", rows, out)
	}
}

func TestPaginationHintHandlesEveryMetaShape(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		meta map[string]any
		want string
	}{
		{map[string]any{"total_pages": json.Number("2")}, "page 1 of 2; use --page N or --all\n"},
		{map[string]any{"total_pages": float64(4), "current_page": int64(2), "total_count": 40}, "page 2 of 4 (40 total); use --page N or --all\n"},
		{map[string]any{"total_pages": "two"}, ""},
		{map[string]any{}, ""},
		{nil, ""},
	} {
		var errOut bytes.Buffer
		Renderer{Err: &errOut}.hintPagination(tc.meta)
		if errOut.String() != tc.want {
			t.Errorf("hint for %v = %q, want %q", tc.meta, errOut.String(), tc.want)
		}
	}
	Renderer{}.hintPagination(map[string]any{"total_pages": json.Number("9")})
}

func itoa(value int) string { return strconv.Itoa(value) }
