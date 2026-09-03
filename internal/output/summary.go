package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// maxSummaryLabels caps how many item labels a nested list prints before `+N more`.
const maxSummaryLabels = 5

// summary renders one value for a table cell. Scalars print as themselves. A nested
// list or object prints as a short description rather than a JSON blob: `3 items: basic,
// premium, enterprise`, `code=quickstart name=Quickstart`, `{12 fields}`. The structured
// form is one flag away in --output json; a table cell is for reading, not parsing.
func summary(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case []any:
		return summarizeList(typed)
	case map[string]any:
		return summarizeObject(typed)
	default:
		return scalar(value)
	}
}

func summarizeList(items []any) string {
	count := len(items)
	noun := "items"
	if count == 1 {
		noun = "item"
	}
	if count == 0 {
		return "0 items"
	}
	labels := make([]string, 0, count)
	for _, item := range items {
		var label string
		switch typed := item.(type) {
		case map[string]any:
			label = identifierLabel(typed)
		case []any, nil:
			label = ""
		default:
			label = scalar(typed)
		}
		if label == "" {
			return fmt.Sprintf("%d %s", count, noun)
		}
		labels = append(labels, label)
	}
	if len(labels) > maxSummaryLabels {
		labels = append(labels[:maxSummaryLabels], fmt.Sprintf("+%d more", len(labels)-maxSummaryLabels))
	}
	return fmt.Sprintf("%d %s: %s", count, noun, strings.Join(labels, ", "))
}

func summarizeObject(object map[string]any) string {
	if len(object) == 0 {
		return "{}"
	}
	if label := identifierPairs(object); label != "" {
		return label
	}
	if len(object) <= 4 {
		pairs := make([]string, 0, len(object))
		for _, key := range sortedKeys(object) {
			if !isScalar(object[key]) {
				pairs = nil
				break
			}
			pairs = append(pairs, key+"="+scalar(object[key]))
		}
		if pairs != nil {
			return strings.Join(pairs, " ")
		}
	}
	noun := "fields"
	if len(object) == 1 {
		noun = "field"
	}
	return fmt.Sprintf("{%d %s}", len(object), noun)
}

// identifierLabel names one nested item by its most readable identifier: the code or
// external ID an operator typed, then the name, then the Lago ID as a last resort.
func identifierLabel(object map[string]any) string {
	for _, key := range []string{"code", "external_id", "name", "lago_id"} {
		if text, ok := object[key].(string); ok && strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}

// identifierPairs renders every identifier key an object carries as `key=value`, in
// identifier order, or "" when it carries none.
func identifierPairs(object map[string]any) string {
	fields := identifiersOf(object)
	if len(fields) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(fields))
	for _, key := range fields {
		pairs = append(pairs, key+"="+scalar(object[key]))
	}
	return strings.Join(pairs, " ")
}

// scalar prints a leaf value. It is the last resort for Go types the renderer does not
// know, where compact JSON is more honest than fmt's struct syntax.
func scalar(value any) string {
	if value == nil {
		return ""
	}
	if stringValue, ok := value.(string); ok {
		return stringValue
	}
	if isScalar(value) {
		return fmt.Sprint(value)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(encoded)
}

func isScalar(value any) bool {
	switch value.(type) {
	case nil, string, bool, float64, json.Number, int, int64:
		return true
	default:
		return false
	}
}

func sortedKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
