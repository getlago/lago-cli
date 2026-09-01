package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/jmespath/go-jmespath"
	"gopkg.in/yaml.v3"
)

const (
	Table = "table"
	JSON  = "json"
	YAML  = "yaml"
)

type Renderer struct {
	Mode  string
	Query string
	Out   io.Writer

	// Identifiers restricts default table output to the terse identifier block.
	// It applies to table output only: --output json and --output yaml always
	// carry the complete resource. See DECISIONS.md.
	Identifiers bool
}

func (r Renderer) Render(value any) error {
	if r.Out == nil {
		r.Out = io.Discard
	}
	if r.Query != "" {
		searchable, err := queryValue(value)
		if err != nil {
			return queryError(err)
		}
		queried, err := jmespath.Search(r.Query, searchable)
		if err != nil {
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid JMESPath query: %v", err), "Check --query syntax or remove the flag.")
		}
		value = queried
	}
	switch r.Mode {
	case "", Table:
		if r.Identifiers {
			return renderIdentifiers(r.Out, value)
		}
		return renderTable(r.Out, value)
	case JSON:
		encoder := json.NewEncoder(r.Out)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(value)
	case YAML:
		encoder := yaml.NewEncoder(r.Out)
		defer encoder.Close()
		return encoder.Encode(yamlValue(value))
	default:
		return apperr.New(apperr.ExitUsage, "output must be table, json, or yaml", "Pass --output table, --output json, or --output yaml.")
	}
}

func renderTable(out io.Writer, value any) error {
	value = unwrapSingle(value)
	rows, ok := value.([]any)
	if !ok {
		if object, objectOK := value.(map[string]any); objectOK {
			w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
			for _, key := range sortedKeys(object) {
				fmt.Fprintf(w, "%s\t%s\n", strings.ToUpper(key), scalar(object[key]))
			}
			return w.Flush()
		}
		_, err := fmt.Fprintln(out, scalar(value))
		return err
	}
	if len(rows) == 0 {
		_, err := fmt.Fprintln(out, "No results.")
		return err
	}
	first, ok := rows[0].(map[string]any)
	if !ok {
		for _, row := range rows {
			if _, err := fmt.Fprintln(out, scalar(row)); err != nil {
				return err
			}
		}
		return nil
	}
	columns := preferredColumns(first)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for index, column := range columns {
		if index > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, strings.ToUpper(column))
	}
	fmt.Fprintln(w)
	for _, rowValue := range rows {
		row, _ := rowValue.(map[string]any)
		for index, column := range columns {
			if index > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, scalar(row[column]))
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

// identifierKeys are the fields an operator needs back from a create or update: the
// Lago ID to address the resource by, the external ID they chose, and the human name
// or code they will recognise it by. Order is the render order, not a preference list.
var identifierKeys = []string{"lago_id", "external_id", "code", "name"}

// renderIdentifiers prints only the identity of a created or updated resource.
//
// A create already echoes back every attribute the caller just sent; the one thing the
// caller does not have is the identifier Lago minted. Printing 40 rows to hide that one
// is the finding this renderer closes. Full detail stays one flag away: --output json.
//
// It never prints nothing: a response that carries no recognisable identifier falls
// back to the full table, because a blank terminal is worse than a verbose one.
func renderIdentifiers(out io.Writer, value any) error {
	unwrapped := unwrapSingle(value)
	switch typed := unwrapped.(type) {
	case map[string]any:
		fields := identifiersOf(typed)
		if len(fields) == 0 {
			return renderTable(out, value)
		}
		w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
		for _, key := range fields {
			fmt.Fprintf(w, "%s\t%s\n", strings.ToUpper(key), scalar(typed[key]))
		}
		return w.Flush()
	case []any:
		if len(typed) == 0 {
			return renderTable(out, value)
		}
		rows := make([]any, 0, len(typed))
		for _, rowValue := range typed {
			row, ok := rowValue.(map[string]any)
			if !ok {
				return renderTable(out, value)
			}
			fields := identifiersOf(row)
			if len(fields) == 0 {
				return renderTable(out, value)
			}
			reduced := make(map[string]any, len(fields))
			for _, key := range fields {
				reduced[key] = row[key]
			}
			rows = append(rows, reduced)
		}
		return renderTable(out, rows)
	default:
		return renderTable(out, value)
	}
}

// identifiersOf returns the identifier fields present on object, in render order.
// A field whose value is an empty string is treated as absent: printing
// `EXTERNAL_ID` with nothing after it tells the operator less than omitting it.
func identifiersOf(object map[string]any) []string {
	fields := make([]string, 0, len(identifierKeys))
	for _, key := range identifierKeys {
		value, exists := object[key]
		if !exists || value == nil {
			continue
		}
		if text, isText := value.(string); isText && strings.TrimSpace(text) == "" {
			continue
		}
		fields = append(fields, key)
	}
	return fields
}

func unwrapSingle(value any) any {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return value
	}
	for _, nested := range object {
		return nested
	}
	return value
}

func preferredColumns(object map[string]any) []string {
	preferred := []string{"lago_id", "id", "code", "external_id", "name", "status", "amount_cents", "currency", "created_at"}
	columns := make([]string, 0, len(object))
	seen := map[string]bool{}
	for _, key := range preferred {
		if _, ok := object[key]; ok {
			columns = append(columns, key)
			seen[key] = true
		}
	}
	for _, key := range sortedKeys(object) {
		if !seen[key] && isScalar(object[key]) && len(columns) < 8 {
			columns = append(columns, key)
		}
	}
	return columns
}

func sortedKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isScalar(value any) bool {
	switch value.(type) {
	case nil, string, bool, float64, json.Number, int, int64:
		return true
	default:
		return false
	}
}

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
