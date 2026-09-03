package output

import (
	"sort"
	"strings"
)

// resourceColumns declares the list columns for the resources QA signed off as
// release-blocking. Keys are the response wrapper (`customers` in
// `{"customers": [...], "meta": {...}}`), values are the columns in render order. A column
// absent from every row is dropped; a resource with no entry, or whose rows match none of
// its columns, falls back to heuristicColumns. Field names are verified against the
// embedded spec by internal/contract, so a spec rename fails the build here rather than
// printing an empty column.
var resourceColumns = map[string][]string{
	"customers":     {"lago_id", "external_id", "name", "email", "currency", "created_at"},
	"invoices":      {"lago_id", "number", "status", "payment_status", "currency", "total_amount_cents", "issuing_date"},
	"subscriptions": {"lago_id", "external_id", "plan_code", "status", "started_at", "ending_at"},
	"plans":         {"lago_id", "code", "name", "interval", "amount_cents", "amount_currency"},
}

// ResourceColumns returns a copy of the declared per-resource column lists, for tests
// that check every declared column exists in the spec.
func ResourceColumns() map[string][]string {
	copied := make(map[string][]string, len(resourceColumns))
	for key, columns := range resourceColumns {
		copied[key] = append([]string(nil), columns...)
	}
	return copied
}

// maxHeuristicColumns bounds the fallback so an unknown resource with 40 attributes
// still fits a terminal. Declared resources are not capped: their lists are the cap.
const maxHeuristicColumns = 8

// columnsFor picks the columns for a list of rows under the given wrapper key.
func columnsFor(key string, rows []map[string]any) []string {
	if declared, ok := resourceColumns[key]; ok {
		present := make([]string, 0, len(declared))
		for _, column := range declared {
			if anyRowHas(rows, column) {
				present = append(present, column)
			}
		}
		if len(present) > 0 {
			return present
		}
	}
	return heuristicColumns(rows)
}

// heuristicColumns ranks the union of scalar keys across every row: identifiers, then
// state, then each money amount immediately followed by its currency, then dates, then
// everything else alphabetically. The union matters: the old renderer read the first
// row only, so a column that happened to be absent there vanished for the whole page.
func heuristicColumns(rows []map[string]any) []string {
	candidates := map[string]bool{}
	for _, row := range rows {
		for key, value := range row {
			if isScalar(value) {
				candidates[key] = true
			}
		}
	}
	ordered := make([]string, 0, len(candidates))
	taken := map[string]bool{}
	take := func(key string) {
		if candidates[key] && !taken[key] {
			ordered = append(ordered, key)
			taken[key] = true
		}
	}
	for _, key := range identifierKeys {
		take(key)
	}
	take("status")
	take("payment_status")
	for _, key := range sortedCandidates(candidates) {
		if !strings.HasSuffix(key, "amount_cents") {
			continue
		}
		take(key)
		take(currencyPartner(key, candidates))
	}
	for _, key := range sortedCandidates(candidates) {
		if strings.HasSuffix(key, "_at") || strings.HasSuffix(key, "_date") {
			take(key)
		}
	}
	for _, key := range sortedCandidates(candidates) {
		take(key)
	}
	if len(ordered) > maxHeuristicColumns {
		ordered = ordered[:maxHeuristicColumns]
	}
	return ordered
}

// currencyPartner returns the currency column that belongs to a money column:
// `amount_cents` pairs with `amount_currency`, `total_amount_cents` with
// `total_amount_currency`, and either falls back to a plain `currency`.
func currencyPartner(moneyKey string, candidates map[string]bool) string {
	partner := strings.TrimSuffix(moneyKey, "cents") + "currency"
	if candidates[partner] {
		return partner
	}
	return "currency"
}

func sortedCandidates(candidates map[string]bool) []string {
	keys := make([]string, 0, len(candidates))
	for key := range candidates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func anyRowHas(rows []map[string]any, column string) bool {
	for _, row := range rows {
		if _, ok := row[column]; ok {
			return true
		}
	}
	return false
}
