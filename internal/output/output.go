package output

import (
	"encoding/json"
	"fmt"
	"io"
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

	// Err receives diagnostics that must not pollute stdout: the hint printed when a
	// query matches nothing, and the paging hint under a partial list. Leaving it nil
	// discards them.
	Err io.Writer

	// Identifiers restricts default table output to the terse identifier block.
	// It applies to table output only: --output json and --output yaml always
	// carry the complete resource. See DECISIONS.md.
	Identifiers bool

	// AllPages is set by --all, which renders every page itself, so the per-page
	// "page N of M" hint would only be noise.
	AllPages bool
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
		r.hintOnEmptyMatch(value, queried)
		value = queried
	}
	switch r.Mode {
	case "", Table:
		if r.Identifiers {
			return r.renderIdentifiers(value)
		}
		return r.renderTable(value)
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

// hintOnEmptyMatch tells the operator when a query silently matched nothing.
//
// Lago wraps every response, so `--query lago_id` against `{"customers": [...]}` is a
// valid expression that matches nothing, and JMESPath answers `null`. QA read that null
// as "no data" twice. The hint goes to stderr and names the keys that were actually
// available; `null` still goes to stdout unchanged, because a script parsing it must not
// have to care that a human was told something.
//
// It stays quiet when the response itself was null, which is a real answer rather than a
// missed match.
func (r Renderer) hintOnEmptyMatch(original, queried any) {
	if r.Err == nil || queried != nil || original == nil {
		return
	}
	message := "query matched nothing"
	if object, ok := original.(map[string]any); ok && len(object) > 0 {
		message += "; top-level keys: " + Sanitize(strings.Join(sortedKeys(object), ", "))
	}
	fmt.Fprintln(r.Err, message)
}

// renderTable is the default output. A list response renders one row per item, a single
// resource renders as key/value rows, and a scalar prints alone.
func (r Renderer) renderTable(value any) error {
	if key, rows, meta, isList := unwrapList(value); isList {
		if err := r.renderRows(key, rows); err != nil {
			return err
		}
		r.hintPagination(meta)
		return nil
	}
	switch typed := unwrapSingle(value).(type) {
	case []any:
		return r.renderRows("", typed)
	case map[string]any:
		return r.renderPairs(typed)
	default:
		_, err := fmt.Fprintln(r.Out, cell(typed))
		return err
	}
}

// renderPairs prints a single resource as key/value rows. Null and blank values are
// omitted: an async endpoint such as `events send` answers with most fields unset, and
// five rows of nothing tell the operator less than the three that carry a value. When
// every value is empty the keys still print, so the output is never blank.
func (r Renderer) renderPairs(object map[string]any) error {
	keys := make([]string, 0, len(object))
	for _, key := range sortedKeys(object) {
		if !isBlank(object[key]) {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		keys = sortedKeys(object)
	}
	w := tabwriter.NewWriter(r.Out, 0, 4, 2, ' ', 0)
	for _, key := range keys {
		fmt.Fprintf(w, "%s\t%s\n", header(key), cell(object[key]))
	}
	return w.Flush()
}

func isBlank(value any) bool {
	if value == nil {
		return true
	}
	text, isText := value.(string)
	return isText && strings.TrimSpace(text) == ""
}

// renderRows prints one table row per item. key is the response wrapper the rows came
// from, which selects a declared column set when one exists.
func (r Renderer) renderRows(key string, rows []any) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(r.Out, "No results.")
		return err
	}
	objects := make([]map[string]any, 0, len(rows))
	for _, rowValue := range rows {
		row, ok := rowValue.(map[string]any)
		if !ok {
			for _, item := range rows {
				if _, err := fmt.Fprintln(r.Out, cell(item)); err != nil {
					return err
				}
			}
			return nil
		}
		objects = append(objects, row)
	}
	columns := columnsFor(key, objects)
	w := tabwriter.NewWriter(r.Out, 0, 4, 2, ' ', 0)
	for index, column := range columns {
		if index > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, header(column))
	}
	fmt.Fprintln(w)
	for _, row := range objects {
		for index, column := range columns {
			if index > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, cell(row[column]))
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

// unwrapList recognises the Lago list envelope: exactly one array-valued key, alone or
// beside a `meta` object. `{"customers": [...], "meta": {...}}` is a list of customers;
// `{"invoice": {...}, "meta": {...}}` is not, and `{"a": [...], "b": [...]}` is not either,
// because unwrapping would drop a field.
func unwrapList(value any) (key string, rows []any, meta map[string]any, ok bool) {
	object, isObject := value.(map[string]any)
	if !isObject || len(object) == 0 || len(object) > 2 {
		return "", nil, nil, false
	}
	for candidate, nested := range object {
		if candidate == "meta" {
			if len(object) == 1 {
				return "", nil, nil, false
			}
			meta, _ = nested.(map[string]any)
			if meta == nil {
				return "", nil, nil, false
			}
			continue
		}
		if key != "" {
			return "", nil, nil, false
		}
		rows, ok = nested.([]any)
		if !ok {
			return "", nil, nil, false
		}
		key = candidate
	}
	return key, rows, meta, key != ""
}

// hintPagination tells the operator, on stderr, that the table is one page of more.
// Table mode omits `meta` from stdout: a pagination object rendered as a row is noise for
// a reader and useless to a script, which reads --output json where meta is intact.
func (r Renderer) hintPagination(meta map[string]any) {
	if r.Err == nil || r.AllPages || meta == nil {
		return
	}
	totalPages, ok := metaInt(meta["total_pages"])
	if !ok || totalPages <= 1 {
		return
	}
	current, ok := metaInt(meta["current_page"])
	if !ok {
		current = 1
	}
	message := fmt.Sprintf("page %d of %d", current, totalPages)
	if total, ok := metaInt(meta["total_count"]); ok {
		message += fmt.Sprintf(" (%d total)", total)
	}
	fmt.Fprintf(r.Err, "%s; use --page N or --all\n", message)
}

func metaInt(value any) (int64, bool) {
	switch typed := value.(type) {
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	case float64:
		return int64(typed), true
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	default:
		return 0, false
	}
}

// identifierKeys are the fields an operator needs back from a write: the Lago ID to
// address the resource by, the external ID they chose, the human name or code they will
// recognise it by, and the status a state transition produced. Order is the render
// order, not a preference list.
var identifierKeys = []string{"lago_id", "external_id", "code", "name", "status"}

// renderIdentifiers prints only the identity of a created or updated resource.
//
// A create already echoes back every attribute the caller just sent; the one thing the
// caller does not have is the identifier Lago minted. Printing 40 rows to hide that one
// is the finding this renderer closes. Full detail stays one flag away: --output json.
//
// It never prints nothing: a response that carries no recognisable identifier falls
// back to the full table, because a blank terminal is worse than a verbose one.
func (r Renderer) renderIdentifiers(value any) error {
	if _, rows, meta, isList := unwrapList(value); isList {
		reduced, ok := reduceToIdentifiers(rows)
		if !ok {
			return r.renderTable(value)
		}
		if err := r.renderRows("", reduced); err != nil {
			return err
		}
		r.hintPagination(meta)
		return nil
	}
	switch typed := unwrapSingle(value).(type) {
	case map[string]any:
		fields := identifiersOf(typed)
		if len(fields) == 0 {
			return r.renderTable(value)
		}
		w := tabwriter.NewWriter(r.Out, 0, 4, 2, ' ', 0)
		for _, key := range fields {
			fmt.Fprintf(w, "%s\t%s\n", header(key), cell(typed[key]))
		}
		return w.Flush()
	case []any:
		reduced, ok := reduceToIdentifiers(typed)
		if !ok {
			return r.renderTable(value)
		}
		return r.renderRows("", reduced)
	default:
		return r.renderTable(value)
	}
}

// reduceToIdentifiers keeps only the identifier keys of every row. It reports false, so
// the caller falls back to the full table, when the list is empty, holds a non-object,
// or holds a row with no identifier at all.
func reduceToIdentifiers(rows []any) ([]any, bool) {
	if len(rows) == 0 {
		return nil, false
	}
	reduced := make([]any, 0, len(rows))
	for _, rowValue := range rows {
		row, ok := rowValue.(map[string]any)
		if !ok {
			return nil, false
		}
		fields := identifiersOf(row)
		if len(fields) == 0 {
			return nil, false
		}
		item := make(map[string]any, len(fields))
		for _, key := range fields {
			item[key] = row[key]
		}
		reduced = append(reduced, item)
	}
	return reduced, true
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
