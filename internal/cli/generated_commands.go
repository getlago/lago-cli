package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func addGenerated(root *cobra.Command, app *App, operations []generated.Operation) {
	resources := map[string]*cobra.Command{}
	for _, operation := range operations {
		resource := resources[operation.Resource]
		if resource == nil {
			resource = &cobra.Command{
				Use:   operation.Resource,
				Short: fmt.Sprintf("Manage Lago %s", strings.ReplaceAll(operation.Resource, "-", " ")),
			}
			resources[operation.Resource] = resource
			root.AddCommand(resource)
		}
		resource.AddCommand(newGeneratedCommand(app, operation))
	}
}

func newGeneratedCommand(app *App, operation generated.Operation) *cobra.Command {
	pathParameters := filterParameters(operation.Parameters, "path")
	use := operation.Action
	for _, parameter := range pathParameters {
		use += " <" + parameter.Name + ">"
	}
	cmd := &cobra.Command{
		Use:     use,
		Short:   firstNonBlank(operation.Summary, operation.OperationID),
		Long:    operation.Description,
		Args:    cobra.ExactArgs(len(pathParameters)),
		Example: generatedExample(operation, pathParameters),
	}

	flagValues := map[string]*string{}
	addFlag := func(name, description string) {
		if _, exists := flagValues[name]; exists {
			return
		}
		value := ""
		flagValues[name] = &value
		cmd.Flags().StringVar(&value, name, "", description)
	}
	for _, parameter := range operation.Parameters {
		if parameter.In != "path" {
			addFlag(parameter.Flag, flagDescription(parameter.Description, parameter.Type, parameter.Enum))
		}
	}
	if operation.Body != nil {
		addFlag("input", "Complete JSON request body or @file.json")
		for _, field := range operation.Body.Fields {
			addFlag(field.Flag, flagDescription(field.Description, field.Type, field.Enum))
		}
	}
	if operation.Paginated {
		addFlag("limit", "Results per page (1-1000)")
		addFlag("page", "Page number, starting at 1")
		all := false
		cmd.Flags().BoolVar(&all, "all", false, "Fetch every page")
		_ = all
	}
	var bulkFile string
	var concurrency int
	if operation.Resource == "events" && operation.Action == "send" {
		cmd.Flags().StringVar(&bulkFile, "file", "", "Stream newline-delimited JSON events from a file, or - for stdin")
		cmd.Flags().IntVar(&concurrency, "concurrency", 4, "Concurrent bulk event requests (1-64)")
	}
	var idempotencyKey string
	var watch bool
	var watchInterval time.Duration
	if operation.Method == http.MethodGet {
		cmd.Flags().BoolVar(&watch, "watch", false, "Poll and re-render when the response changes")
		cmd.Flags().DurationVar(&watchInterval, "watch-interval", 2*time.Second, "Polling interval used with --watch")
	}
	if !isIdempotentMethod(operation.Method) {
		cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key for safe mutation retries")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if bulkFile != "" {
			return runEventStream(cmd, app, bulkFile, concurrency)
		}
		path := operation.Path
		for index, parameter := range pathParameters {
			path = strings.ReplaceAll(path, "{"+parameter.Name+"}", url.PathEscape(args[index]))
		}
		query := make(url.Values)
		for _, parameter := range operation.Parameters {
			if parameter.In != "query" || !cmd.Flags().Changed(parameter.Flag) {
				continue
			}
			for _, value := range queryValues(*flagValues[parameter.Flag], parameter.Type) {
				query.Add(parameter.Name, value)
			}
		}
		if operation.Paginated {
			if cmd.Flags().Changed("limit") {
				limit, err := parseLimit(*flagValues["limit"])
				if err != nil {
					return err
				}
				query.Set("per_page", limit)
			}
			if cmd.Flags().Changed("page") {
				page, err := parsePage(*flagValues["page"])
				if err != nil {
					return err
				}
				query.Set("page", page)
			}
			all, _ := cmd.Flags().GetBool("all")
			if all {
				return runAllPages(cmd, app, operation, path, query)
			}
		}
		// Read before the generated-ID block below, which sets the flag and marks it changed.
		callerChoseTransactionID := cmd.Flags().Changed("transaction-id")
		if operation.Resource == "events" && operation.Action == "send" && operation.Body != nil && !cmd.Flags().Changed("input") {
			if transactionID, exists := flagValues["transaction-id"]; exists && !cmd.Flags().Changed("transaction-id") {
				generatedID := defaultIdempotencyKey()
				if err := cmd.Flags().Set("transaction-id", generatedID); err != nil {
					return apperr.Wrap(apperr.ExitGeneral, "set event transaction ID", err)
				}
				*transactionID = generatedID
				if idempotencyKey == "" {
					idempotencyKey = generatedID
				}
			}
		}
		body, err := generatedBody(app.In, cmd, operation.Body, flagValues)
		if err != nil {
			return err
		}
		if operation.Resource == "events" && operation.Action == "send" && (callerChoseTransactionID || cmd.Flags().Changed("input")) {
			warnRetryUnsafeBody(app, body)
		}
		if operation.Body != nil && operation.Body.Required && len(body) == 0 && operation.Method != http.MethodGet && operation.Method != http.MethodDelete {
			return apperr.New(apperr.ExitUsage, "request body is required", "Pass generated field flags or --input @payload.json.")
		}
		if operation.Dangerous {
			identifier := operation.Resource
			if len(args) > 0 {
				identifier = args[len(args)-1]
			}
			if err := app.Confirm(identifier); err != nil {
				return err
			}
		}
		headers := make(http.Header)
		if idempotencyKey != "" {
			headers.Set("Idempotency-Key", idempotencyKey)
		}
		idempotent := operation.Idempotent || idempotencyKey != ""
		if watch {
			if watchInterval < 500*time.Millisecond {
				return apperr.New(apperr.ExitUsage, "watch interval must be at least 500ms", "Pass --watch-interval 2s or a longer duration.")
			}
			return runWatch(cmd, app, operation, path, query, watchInterval)
		}
		subjects := identifierSubjects(operation, pathParameters, args, cmd, flagValues)
		value, response, err := app.Request(cmd.Context(), transport.Request{Method: operation.Method, Path: path, Query: query, Headers: headers, Body: body, Idempotent: idempotent, Subjects: subjects})
		if err != nil {
			return err
		}
		if operation.Mutation {
			return app.RenderMutation(value, response)
		}
		return app.Render(value, response)
	}
	return cmd
}

func runWatch(cmd *cobra.Command, app *App, operation generated.Operation, path string, query url.Values, interval time.Duration) error {
	var previous []byte
	render := func() error {
		value, response, err := app.Request(cmd.Context(), transport.Request{Method: operation.Method, Path: path, Query: query, Idempotent: true})
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return apperr.Wrap(apperr.ExitGeneral, "compare watched response", err)
		}
		if bytes.Equal(encoded, previous) {
			return nil
		}
		previous = encoded
		return app.Render(value, response)
	}
	if err := render(); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case <-ticker.C:
			if err := render(); err != nil {
				return err
			}
		}
	}
}

func filterParameters(parameters []generated.Parameter, location string) []generated.Parameter {
	filtered := make([]generated.Parameter, 0)
	for _, parameter := range parameters {
		if parameter.In == location {
			filtered = append(filtered, parameter)
		}
	}
	return filtered
}

func generatedBody(in interface{ Read([]byte) (int, error) }, cmd *cobra.Command, body *generated.Body, values map[string]*string) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	if input, exists := values["input"]; exists && cmd.Flags().Changed("input") {
		return readData(in, *input)
	}
	root := map[string]any{}
	target := root
	if body.Wrapper != "" {
		target = map[string]any{}
		root[body.Wrapper] = target
	}
	set := false
	missing := make([]string, 0)
	for _, field := range body.Fields {
		if !cmd.Flags().Changed(field.Flag) {
			if field.Required {
				missing = append(missing, "--"+field.Flag)
			}
			continue
		}
		value, err := parseGeneratedValue(*values[field.Flag], field.Type, field.Complex)
		if err != nil {
			return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid --%s value: %v", field.Flag, err), "Check the field type in `--help`; objects and arrays use JSON syntax.")
		}
		setNested(target, field.Path, value)
		set = true
	}
	if len(missing) > 0 {
		return nil, apperr.New(apperr.ExitUsage, "missing required request fields: "+strings.Join(missing, ", "), "Pass each required flag or provide the complete payload with --input.")
	}
	if !set {
		return nil, nil
	}
	encoded, err := json.Marshal(root)
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitGeneral, "encode generated request body", err)
	}
	return encoded, nil
}

func parseGeneratedValue(raw, valueType string, complex bool) (any, error) {
	if complex || valueType == "object" || valueType == "array" {
		var value any
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		return value, nil
	}
	switch valueType {
	case "boolean":
		return strconv.ParseBool(raw)
	case "integer":
		return strconv.ParseInt(raw, 10, 64)
	case "number":
		if _, err := strconv.ParseInt(strings.ReplaceAll(raw, ".", ""), 10, 64); err != nil {
			return nil, fmt.Errorf("expected an exact decimal string")
		}
		return json.Number(raw), nil
	default:
		return raw, nil
	}
}

func setNested(root map[string]any, path []string, value any) {
	if len(path) == 0 {
		return
	}
	current := root
	for _, component := range path[:len(path)-1] {
		next, ok := current[component].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[component] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
}

func queryValues(raw, valueType string) []string {
	if valueType != "array" {
		return []string{raw}
	}
	var values []string
	if json.Unmarshal([]byte(raw), &values) == nil {
		return values
	}
	parts := strings.Split(raw, ",")
	values = values[:0]
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}

// maxPerPage bounds --limit client-side. QA F-19: `--limit 0` reached the API as
// per_page=0 and lago-api answered 500, because its paginator hands per_page straight to
// Kaminari (BaseQuery#paginate: `scope.page(page).per(limit)`) with no max_per_page, and
// a zero page size divides by zero in total_pages. The spec declares no maximum either,
// so the upper bound here is a sanity bound, not a server contract: nobody reads a
// thousand-row table, and --all exists for the rest.
const maxPerPage = 1000

// defaultAllPageSize is the page size --all requests when --limit is not given.
const defaultAllPageSize = 100

// parseLimit validates --limit as an integer in 1..maxPerPage and returns it as the
// query-string value.
func parseLimit(raw string) (string, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 1 || value > maxPerPage {
		return "", apperr.New(apperr.ExitUsage, fmt.Sprintf("--limit must be an integer between 1 and %d, got %q", maxPerPage, raw), "Pass --limit 100 for the largest common page, or --all to walk every page.")
	}
	return strconv.Itoa(value), nil
}

// parsePage validates --page as a positive integer and returns it as the query-string
// value. It guards the single-page path as well as --all, which validated it already.
func parsePage(raw string) (string, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 1 {
		return "", apperr.New(apperr.ExitUsage, fmt.Sprintf("--page must be a positive integer, got %q", raw), "Pass --page 1 or omit it.")
	}
	return strconv.Itoa(value), nil
}

func runAllPages(cmd *cobra.Command, app *App, operation generated.Operation, path string, query url.Values) error {
	if app.query != "" {
		return apperr.New(apperr.ExitUsage, "--all cannot safely buffer a full-collection JMESPath query", "Use a per-page --query without --all, or stream JSON pages and process them with jq.")
	}
	page := 1
	if raw := query.Get("page"); raw != "" {
		validated, err := parsePage(raw)
		if err != nil {
			return err
		}
		page, _ = strconv.Atoi(validated)
	}
	if query.Get("per_page") == "" {
		query.Set("per_page", strconv.Itoa(defaultAllPageSize))
	}
	for {
		query.Set("page", strconv.Itoa(page))
		value, response, err := app.Request(cmd.Context(), transport.Request{Method: operation.Method, Path: path, Query: query, Idempotent: true})
		if err != nil {
			return err
		}
		if err := app.Render(value, response); err != nil {
			return err
		}
		more, err := hasNextPage(value, page)
		if err != nil || !more {
			return err
		}
		page++
	}
}

func hasNextPage(value any, current int) (bool, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return false, nil
	}
	meta, _ := object["meta"].(map[string]any)
	if meta == nil {
		return false, nil
	}
	totalPages, ok := jsonInteger(meta["total_pages"])
	if !ok {
		return false, nil
	}
	return int64(current) < totalPages, nil
}

func jsonInteger(value any) (int64, bool) {
	switch typed := value.(type) {
	case json.Number:
		result, err := typed.Int64()
		return result, err == nil
	case float64:
		return int64(typed), typed == float64(int64(typed))
	case int64:
		return typed, true
	default:
		return 0, false
	}
}

// identifierSubjects collects the identifiers a command was given, so a 404 can name
// what was not found instead of returning a bare "Not Found".
//
// QA passed a plan code as --external-subscription-id and read the resulting empty error
// as "no usage" rather than "no such subscription". Path arguments are always
// identifiers; among flags only the ones whose names are identifier-shaped are, so a
// --name or an --amount-cents never ends up in the message.
func identifierSubjects(operation generated.Operation, pathParameters []generated.Parameter, args []string, cmd *cobra.Command, values map[string]*string) []transport.Subject {
	subjects := make([]transport.Subject, 0, len(args)+2)
	for index, parameter := range pathParameters {
		if index < len(args) && args[index] != "" {
			subjects = append(subjects, transport.Subject{Kind: subjectKind(parameter.Name, operation.Resource), Value: args[index]})
		}
	}
	for _, parameter := range operation.Parameters {
		if parameter.In == "path" || !isIdentifierName(parameter.Name) || !cmd.Flags().Changed(parameter.Flag) {
			continue
		}
		if value := *values[parameter.Flag]; value != "" {
			subjects = append(subjects, transport.Subject{Kind: subjectKind(parameter.Name, operation.Resource), Value: value})
		}
	}
	if operation.Body != nil {
		for _, field := range operation.Body.Fields {
			name := field.Path[len(field.Path)-1]
			if !isIdentifierName(name) || !cmd.Flags().Changed(field.Flag) {
				continue
			}
			if value := *values[field.Flag]; value != "" {
				subjects = append(subjects, transport.Subject{Kind: subjectKind(name, operation.Resource), Value: value})
			}
		}
	}
	return subjects
}

// identifierNames are the field-name shapes that address an existing resource. A create
// payload's `name` or `amount_cents` is data; `external_subscription_id` is a lookup.
func isIdentifierName(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "id", "code":
		return true
	}
	return strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "_code")
}

// subjectKind turns an identifier field name into the resource type it addresses:
// `external_subscription_id` and `subscription_external_id` both become "subscription",
// `plan_code` becomes "plan", and a bare `id` or `code` falls back to the command's own
// resource so the message still names something.
func subjectKind(name, resource string) string {
	lower := strings.ToLower(name)
	if lower == "id" || lower == "code" {
		return singularResource(resource)
	}
	lower = strings.TrimSuffix(strings.TrimSuffix(lower, "_id"), "_code")
	lower = strings.TrimPrefix(lower, "external_")
	lower = strings.TrimSuffix(lower, "_external")
	lower = strings.TrimPrefix(lower, "lago_")
	if lower == "" || lower == "external" || lower == "lago" {
		lower = singularResource(resource)
	}
	return strings.ReplaceAll(lower, "_", " ")
}

// singularResource turns a plural command group into the noun for one of its members.
func singularResource(resource string) string {
	resource = strings.ReplaceAll(resource, "-", " ")
	switch {
	case strings.HasSuffix(resource, "ies"):
		return strings.TrimSuffix(resource, "ies") + "y"
	// -sses, -xes and -ches take "es", so trimming a bare "s" would leave "taxe".
	case strings.HasSuffix(resource, "sses"), strings.HasSuffix(resource, "xes"), strings.HasSuffix(resource, "ches"):
		return strings.TrimSuffix(resource, "es")
	case strings.HasSuffix(resource, "s"):
		return strings.TrimSuffix(resource, "s")
	default:
		return resource
	}
}

func generatedExample(operation generated.Operation, pathParameters []generated.Parameter) string {
	var command strings.Builder
	command.WriteString("  lago ")
	command.WriteString(operation.Resource)
	command.WriteByte(' ')
	command.WriteString(operation.Action)
	for _, parameter := range pathParameters {
		command.WriteString(" <")
		command.WriteString(parameter.Name)
		command.WriteByte('>')
	}
	if operation.Body != nil {
		command.WriteString(" --input @payload.json")
	}
	if operation.Mutation {
		// The terse default is the whole point of the identifier renderer, so the
		// example has to teach the escape hatch or the operator will think the
		// remaining attributes were lost.
		full := command.String()
		return full + "\n" + full + " --output json  # full resource"
	}
	return command.String()
}

func flagDescription(description, valueType string, enum []string) string {
	description = strings.ReplaceAll(strings.TrimSpace(description), "`", "'")
	if description == "" {
		description = "API field (" + valueType + ")"
	}
	if len(enum) > 0 {
		description += "; one of: " + strings.Join(enum, ", ")
	}
	return description
}

func defaultIdempotencyKey() string { return uuid.NewString() }

var _ = bytes.MinRead
