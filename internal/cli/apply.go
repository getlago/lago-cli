package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/output"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// `lago v2 apply DIR` reconciles the product catalog of one organization with a folder
// of YAML files: billable metrics, product categories, products, rate cards with their
// rates, and plans with their applied rate cards. Every object is identified by its
// code. See DECISIONS.md, "lago v2 apply".

// applyKind describes one top-level catalog collection. The field lists mirror the
// strong parameters of the lago-api v2 controllers, not the OpenAPI document: the
// document leaves `description` out of several update bodies that the API accepts.
type applyKind struct {
	key     string // YAML key, list response key and state key
	label   string // human name used in the plan and in errors
	wrapper string // request and response envelope
	path    string
	// createFields are the fields a file may declare. updateFields are the ones the API
	// lets apply change in place; a difference on any other field is refused.
	createFields []string
	updateFields []string
	// deletable is false when the API has no DELETE route for the kind.
	deletable bool
	// child names the nested collection reconciled with each object, if any.
	child *applyChildKind
}

// applyChildKind is a collection that lives under one parent object: the rates of a
// rate card, the applied rate cards of a plan.
type applyChildKind struct {
	key          string // YAML key inside the parent, and list response key
	label        string
	wrapper      string
	segment      string // path segment under the parent
	identity     string // field that identifies the child within its parent
	createFields []string
	updateFields []string
	// sentWithParent is true when the children are part of the parent's create body,
	// as rates are part of POST /rate_cards.
	sentWithParent bool
	// deletable reports whether an existing, undeclared child may be deleted. Rates are
	// an append-only timeline: only a pending rate can go; active and terminated rates
	// are history and are left alone.
	deletable func(existing map[string]any) bool
}

var (
	applyRates = &applyChildKind{
		key: "rates", label: "rate", wrapper: "rate", segment: "rates", identity: "code",
		createFields:   []string{"code", "effective_from", "rate_model", "rate_properties", "min_amount_cents", "billing_interval_count", "billing_interval_unit", "applied_pricing_unit_conversion_rate"},
		updateFields:   []string{"effective_from", "rate_model", "rate_properties", "min_amount_cents", "billing_interval_count", "billing_interval_unit", "applied_pricing_unit_conversion_rate"},
		sentWithParent: true,
		deletable: func(existing map[string]any) bool {
			status, _ := existing["status"].(string)
			return status == "pending"
		},
	}
	applyAppliedRateCards = &applyChildKind{
		key: "applied_rate_cards", label: "applied rate card", wrapper: "applied_rate_card", segment: "applied_rate_cards", identity: "rate_card_code",
		createFields: []string{"rate_card_code", "units"},
		updateFields: []string{"units"},
		deletable:    func(map[string]any) bool { return true },
	}

	// applyKinds is in dependency order: an object only references objects of an
	// earlier kind. Creates and updates run in this order, deletes in reverse.
	applyKinds = []*applyKind{
		{
			key: "billable_metrics", label: "billable metric", wrapper: "billable_metric", path: "/billable_metrics",
			createFields: []string{"code", "name", "description", "aggregation_type", "weighted_interval", "recurring", "field_name", "expression", "rounding_function", "rounding_precision", "filters"},
			updateFields: []string{"name", "description", "aggregation_type", "weighted_interval", "recurring", "field_name", "expression", "rounding_function", "rounding_precision", "filters"},
			deletable:    true,
		},
		{
			key: "product_categories", label: "product category", wrapper: "product_category", path: "/product_categories",
			createFields: []string{"code", "name", "description", "invoice_display_name"},
			updateFields: []string{"name", "description", "invoice_display_name"},
			deletable:    true,
		},
		{
			key: "products", label: "product", wrapper: "product", path: "/products",
			createFields: []string{"code", "name", "description", "invoice_display_name", "product_type", "product_category_code", "billable_metric_code"},
			updateFields: []string{"name", "description", "invoice_display_name", "product_category_code"},
			deletable:    true,
		},
		{
			key: "rate_cards", label: "rate card", wrapper: "rate_card", path: "/rate_cards",
			createFields: []string{"code", "product_code", "product_filter_code", "name", "description", "currency", "billing_timing", "proration", "display_on_invoice", "regroup_paid_fees", "applied_pricing_unit_code", "tax_codes"},
			updateFields: []string{"name", "description", "currency", "billing_timing", "proration", "display_on_invoice", "regroup_paid_fees", "applied_pricing_unit_code", "tax_codes"},
			deletable:    true,
			child:        applyRates,
		},
		{
			key: "plans", label: "plan", wrapper: "plan", path: "/plans",
			createFields: []string{"code", "name", "description", "invoice_display_name", "currency", "tax_codes"},
			updateFields: []string{"name", "description", "invoice_display_name", "currency", "tax_codes"},
			deletable:    false,
			child:        applyAppliedRateCards,
		},
	}
)

// applyReferences lists, per kind, the fields that name an object of another kind, so a
// typo fails in the plan instead of halfway through the apply.
var applyReferences = map[string]map[string]string{
	"products":   {"product_category_code": "product_categories", "billable_metric_code": "billable_metrics"},
	"rate_cards": {"product_code": "products"},
	"plans":      {},
}

const applyStateVersion = 1

// applyState records which objects a folder manages in one organization, so --prune
// deletes what the folder created and removed, never an object someone else made.
type applyState struct {
	Version        int                 `json:"version"`
	OrganizationID string              `json:"organization_id"`
	Organization   string              `json:"organization,omitempty"`
	Managed        map[string][]string `json:"managed"`
}

type applyObject struct {
	code     string
	fields   map[string]any
	children []applyChild
	source   string
}

type applyChild struct {
	identity string
	fields   map[string]any
}

type applyDesired map[string][]*applyObject

type applyRemote struct {
	objects  map[string]map[string]map[string]any            // kind -> code -> object
	children map[string]map[string]map[string]map[string]any // kind -> parent code -> identity -> child
}

// applyChange is one line of the plan. The exported fields are what --output json shows.
type applyChange struct {
	Action string   `json:"action"`
	Kind   string   `json:"kind"`
	Code   string   `json:"code"`
	Parent string   `json:"parent,omitempty"`
	Fields []string `json:"fields,omitempty"`
	Note   string   `json:"note,omitempty"`

	kind    *applyKind
	child   *applyChildKind
	method  string
	path    string
	body    map[string]any
	skipped bool
}

const (
	applyCreate    = "create"
	applyUpdate    = "update"
	applyUnchanged = "unchanged"
	applyDelete    = "delete"
	applyRelease   = "release"
)

func newV2ApplyCommand(app *App) *cobra.Command {
	var prune bool
	var statePath string
	command := &cobra.Command{
		Use:   "apply DIR",
		Short: "Reconcile the product catalog with a folder of YAML files",
		Long: "Read every .yml and .yaml file under DIR and make the organization's product catalog match it: " +
			"billable metrics, product categories, products, rate cards with their rates, and plans with their applied rate cards. " +
			"Objects are matched by code. apply prints the plan first, then creates and updates in dependency order.\n\n" +
			"Pass --dry-run to print the plan without changing anything.\n\n" +
			"Nothing is deleted unless you pass --prune. apply records the objects it manages in a state file, " +
			"DIR/.lago-state/<organization id>.json by default; commit it with the YAML. --prune deletes the managed objects " +
			"the files no longer declare, and never touches an object the folder did not create or adopt. " +
			"The applied rate cards of a declared plan and the pending rates of a declared rate card belong to that object, " +
			"so --prune also removes the ones the files dropped. Active and terminated rates are history and are kept. " +
			"Deleting asks for confirmation; pass --confirm prune in scripts.",
		Example: "  lago v2 apply billing/ --dry-run\n  lago v2 apply billing/\n  lago v2 apply billing/ --prune --confirm prune",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd.Context(), app, args[0], statePath, prune)
		},
	}
	command.Flags().BoolVar(&prune, "prune", false, "Delete managed objects the files no longer declare")
	command.Flags().StringVar(&statePath, "state", "", "State file path (default DIR/.lago-state/<organization id>.json)")
	return command
}

func runApply(ctx context.Context, app *App, dir, statePath string, prune bool) error {
	desired, err := loadApplyDir(dir)
	if err != nil {
		return err
	}
	client, err := app.Client(true)
	if err != nil {
		return err
	}
	organization, err := applyGet(ctx, client, "/organizations")
	if err != nil {
		return err
	}
	org := unwrapNamedObject(organization, "organization")
	orgID, _ := org["lago_id"].(string)
	orgName, _ := org["name"].(string)
	if orgID == "" {
		return apperr.New(apperr.ExitGeneral, "the organization response has no lago_id", "Check that the profile points at a Lago API that serves /api/v2/organizations.")
	}
	if statePath == "" {
		statePath = filepath.Join(dir, ".lago-state", orgID+".json")
	}
	state, err := readApplyState(statePath, orgID)
	if err != nil {
		return err
	}
	state.Organization = orgName

	remote, err := fetchApplyRemote(ctx, client, desired)
	if err != nil {
		return err
	}
	changes, err := planApply(desired, remote, state, prune)
	if err != nil {
		return err
	}
	if err := renderApplyPlan(app, changes, prune); err != nil {
		return err
	}
	if app.dryRun {
		fmt.Fprintln(app.Err, "Dry run: no request was sent.")
		return nil
	}
	pending := 0
	deletes := 0
	for _, change := range changes {
		if change.method == "" || change.skipped {
			continue
		}
		pending++
		if change.Action == applyDelete {
			deletes++
		}
	}
	if pending == 0 {
		fmt.Fprintln(app.Err, "Nothing to apply.")
		return writeApplyState(statePath, state, desired, remote, nil)
	}
	if deletes > 0 {
		if err := app.Confirm("prune"); err != nil {
			return err
		}
	}
	applied, applyErr := executeApply(ctx, app, changes)
	if err := writeApplyState(statePath, state, desired, remote, applied); err != nil && applyErr == nil {
		applyErr = err
	}
	if applyErr != nil {
		return applyErr
	}
	fmt.Fprintf(app.Err, "Applied %d change(s). State saved to %s.\n", len(applied), statePath)
	return nil
}

// --- loading ---------------------------------------------------------------------

func loadApplyDir(dir string) (applyDesired, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitUsage, fmt.Sprintf("read %s", dir), err)
	}
	var files []string
	if !info.IsDir() {
		files = []string{dir}
	} else {
		walkErr := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if path != dir && strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if extension := filepath.Ext(path); extension == ".yml" || extension == ".yaml" {
				files = append(files, path)
			}
			return nil
		})
		if walkErr != nil {
			return nil, apperr.Wrap(apperr.ExitUsage, fmt.Sprintf("read %s", dir), walkErr)
		}
		sort.Strings(files)
	}
	if len(files) == 0 {
		return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("no .yml or .yaml file in %s", dir), "Point apply at the folder that holds the catalog files.")
	}
	desired := applyDesired{}
	seen := map[string]string{} // kind/code -> file
	for _, file := range files {
		if err := loadApplyFile(file, desired, seen); err != nil {
			return nil, err
		}
	}
	return desired, nil
}

func loadApplyFile(file string, desired applyDesired, seen map[string]string) error {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return apperr.Wrap(apperr.ExitUsage, fmt.Sprintf("read %s", file), err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var node yaml.Node
		if err := decoder.Decode(&node); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s: %v", file, err), "Fix the YAML syntax.")
		}
		value, err := yamlNodeValue(&node)
		if err != nil {
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s: %v", file, err), "")
		}
		if value == nil {
			continue
		}
		document, ok := value.(map[string]any)
		if !ok {
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s: the document is not a map", file), applyKeysSuggestion())
		}
		for key, raw := range document {
			kind := findApplyKind(key)
			if kind == nil {
				return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s: unknown key %q", file, key), applyKeysSuggestion())
			}
			items, ok := raw.([]any)
			if !ok {
				return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s: %s must be a list", file, key), "")
			}
			for index, item := range items {
				object, err := parseApplyObject(kind, item)
				if err != nil {
					return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s: %s[%d]: %v", file, key, index, err), "")
				}
				object.source = file
				id := key + "/" + object.code
				if previous, exists := seen[id]; exists {
					return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s %q is declared twice: in %s and in %s", kind.label, object.code, previous, file), "Keep one declaration per code.")
				}
				seen[id] = file
				desired[key] = append(desired[key], object)
			}
		}
	}
}

func applyKeysSuggestion() string {
	keys := make([]string, 0, len(applyKinds))
	for _, kind := range applyKinds {
		keys = append(keys, kind.key)
	}
	return "A catalog file is a map whose keys are among: " + strings.Join(keys, ", ") + "."
}

func findApplyKind(key string) *applyKind {
	for _, kind := range applyKinds {
		if kind.key == key {
			return kind
		}
	}
	return nil
}

func parseApplyObject(kind *applyKind, raw any) (*applyObject, error) {
	fields, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("not a map")
	}
	code, _ := fields["code"].(string)
	if code == "" {
		return nil, errors.New("code is required and must be a string")
	}
	object := &applyObject{code: code, fields: map[string]any{}}
	for key, value := range fields {
		if kind.child != nil && key == kind.child.key {
			children, err := parseApplyChildren(kind.child, value)
			if err != nil {
				return nil, fmt.Errorf("%s %q: %w", kind.label, code, err)
			}
			object.children = children
			continue
		}
		if !contains(kind.createFields, key) {
			return nil, fmt.Errorf("%s %q: unknown field %q (allowed: %s)", kind.label, code, key, strings.Join(allowedApplyFields(kind), ", "))
		}
		object.fields[key] = value
	}
	return object, nil
}

func allowedApplyFields(kind *applyKind) []string {
	fields := append([]string{}, kind.createFields...)
	if kind.child != nil {
		fields = append(fields, kind.child.key)
	}
	return fields
}

func parseApplyChildren(child *applyChildKind, raw any) ([]applyChild, error) {
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a list", child.key)
	}
	children := make([]applyChild, 0, len(items))
	seen := map[string]bool{}
	for index, item := range items {
		fields, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s[%d] is not a map", child.key, index)
		}
		identity, _ := fields[child.identity].(string)
		if identity == "" {
			return nil, fmt.Errorf("%s[%d]: %s is required and must be a string", child.key, index, child.identity)
		}
		if seen[identity] {
			return nil, fmt.Errorf("%s %q is declared twice", child.label, identity)
		}
		seen[identity] = true
		for key := range fields {
			if !contains(child.createFields, key) {
				return nil, fmt.Errorf("%s %q: unknown field %q (allowed: %s)", child.label, identity, key, strings.Join(child.createFields, ", "))
			}
		}
		children = append(children, applyChild{identity: identity, fields: fields})
	}
	return children, nil
}

// jsonNumber matches what encoding/json accepts as a number, so a YAML scalar such as
// 1_000 or 0x10 is refused rather than sent as text the API would misread.
var jsonNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// yamlNodeValue converts a YAML node to the values encoding/json produces with
// UseNumber. Numbers keep their literal text as json.Number, so 0.00002 is sent as
// written and never goes through a float64.
func yamlNodeValue(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return yamlNodeValue(node.Content[0])
	case yaml.AliasNode:
		return yamlNodeValue(node.Alias)
	case yaml.MappingNode:
		object := make(map[string]any, len(node.Content)/2)
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			if keyNode.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("line %d: map keys must be strings", keyNode.Line)
			}
			if _, exists := object[keyNode.Value]; exists {
				return nil, fmt.Errorf("line %d: key %q is repeated", keyNode.Line, keyNode.Value)
			}
			value, err := yamlNodeValue(node.Content[index+1])
			if err != nil {
				return nil, err
			}
			object[keyNode.Value] = value
		}
		return object, nil
	case yaml.SequenceNode:
		items := make([]any, 0, len(node.Content))
		for _, item := range node.Content {
			value, err := yamlNodeValue(item)
			if err != nil {
				return nil, err
			}
			items = append(items, value)
		}
		return items, nil
	case yaml.ScalarNode:
		switch node.ShortTag() {
		case "!!null":
			return nil, nil
		case "!!bool":
			var value bool
			if err := node.Decode(&value); err != nil {
				return nil, fmt.Errorf("line %d: %w", node.Line, err)
			}
			return value, nil
		case "!!int", "!!float":
			if !jsonNumber.MatchString(node.Value) {
				return nil, fmt.Errorf("line %d: %q is not a plain decimal number; write it as one or quote it", node.Line, node.Value)
			}
			return json.Number(node.Value), nil
		default:
			// Strings, and timestamps kept exactly as written.
			return node.Value, nil
		}
	}
	return nil, fmt.Errorf("line %d: unsupported YAML node", node.Line)
}

// --- remote state ------------------------------------------------------------------

func applyGet(ctx context.Context, client *transport.Client, path string) (any, error) {
	response, err := client.Do(ctx, transport.Request{Method: http.MethodGet, Path: path, Idempotent: true, APIVersion: "v2"})
	if err != nil {
		return nil, err
	}
	value, err := transport.DecodeJSON(response.Body)
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitGeneral, "decode Lago API response", err)
	}
	return value, nil
}

// applyList walks a v2 cursor-paginated list and returns its items keyed by field.
func applyList(ctx context.Context, client *transport.Client, path, key, field string) (map[string]map[string]any, error) {
	items := map[string]map[string]any{}
	after := ""
	for {
		query := url.Values{"limit": {fmt.Sprint(maxCursorPageSize)}}
		if after != "" {
			query.Set("after", after)
		}
		response, err := client.Do(ctx, transport.Request{Method: http.MethodGet, Path: path, Query: query, Idempotent: true, APIVersion: "v2"})
		if err != nil {
			return nil, err
		}
		value, err := transport.DecodeJSON(response.Body)
		if err != nil {
			return nil, apperr.Wrap(apperr.ExitGeneral, "decode Lago API response", err)
		}
		object, _ := value.(map[string]any)
		list, _ := object[key].([]any)
		for _, raw := range list {
			if item, ok := raw.(map[string]any); ok {
				if identity, _ := item[field].(string); identity != "" {
					items[identity] = item
				}
			}
		}
		after = nextCursor(value)
		if after == "" {
			return items, nil
		}
	}
}

func fetchApplyRemote(ctx context.Context, client *transport.Client, desired applyDesired) (*applyRemote, error) {
	remote := &applyRemote{objects: map[string]map[string]map[string]any{}, children: map[string]map[string]map[string]map[string]any{}}
	for _, kind := range applyKinds {
		objects, err := applyList(ctx, client, kind.path, kind.key, "code")
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", kind.key, err)
		}
		remote.objects[kind.key] = objects
		if kind.child == nil {
			continue
		}
		remote.children[kind.key] = map[string]map[string]map[string]any{}
		for _, object := range desired[kind.key] {
			if _, exists := objects[object.code]; !exists {
				continue
			}
			path := kind.path + "/" + url.PathEscape(object.code) + "/" + kind.child.segment
			children, err := applyList(ctx, client, path, kind.child.key, kind.child.identity)
			if err != nil {
				return nil, fmt.Errorf("list %s of %s %q: %w", kind.child.key, kind.label, object.code, err)
			}
			remote.children[kind.key][object.code] = children
		}
	}
	return remote, nil
}

// --- planning ----------------------------------------------------------------------

func planApply(desired applyDesired, remote *applyRemote, state *applyState, prune bool) ([]*applyChange, error) {
	if err := checkApplyReferences(desired, remote); err != nil {
		return nil, err
	}
	var upserts, childUpserts, childDeletes, deletes []*applyChange
	for _, kind := range applyKinds {
		var kindChildDeletes []*applyChange
		existing := remote.objects[kind.key]
		managed := stringSet(state.Managed[kind.key])
		declared := map[string]bool{}
		for _, object := range desired[kind.key] {
			declared[object.code] = true
			current, exists := existing[object.code]
			path := kind.path + "/" + url.PathEscape(object.code)
			if !exists {
				body := copyFields(object.fields)
				change := &applyChange{Action: applyCreate, Kind: kind.label, Code: object.code, kind: kind, method: http.MethodPost, path: kind.path, body: body}
				if kind.child != nil && kind.child.sentWithParent {
					rates := make([]any, 0, len(object.children))
					for _, child := range object.children {
						rates = append(rates, child.fields)
					}
					body[kind.child.key] = rates
					if len(rates) > 0 {
						change.Note = fmt.Sprintf("with %d %s(s)", len(rates), kind.child.label)
					}
				}
				upserts = append(upserts, change)
			} else {
				fields, err := diffApplyFields(kind.label, object.code, current, object.fields, kind.updateFields)
				if err != nil {
					return nil, err
				}
				change := &applyChange{Kind: kind.label, Code: object.code, kind: kind, Fields: fields}
				if len(fields) == 0 {
					change.Action = applyUnchanged
				} else {
					change.Action, change.method, change.path, change.body = applyUpdate, http.MethodPut, path, pickFields(object.fields, fields)
				}
				if !managed[object.code] {
					change.Note = "adopted: already in Lago, now managed"
				}
				upserts = append(upserts, change)
			}
			if kind.child == nil || (!exists && kind.child.sentWithParent) {
				continue
			}
			created, removed, err := planApplyChildren(kind, object, remote.children[kind.key][object.code], exists, prune)
			if err != nil {
				return nil, err
			}
			childUpserts = append(childUpserts, created...)
			kindChildDeletes = append(kindChildDeletes, removed...)
		}
		// Children go in reverse kind order too: applied rate cards before rates.
		childDeletes = append(kindChildDeletes, childDeletes...)
		for _, code := range sortedStrings(managed) {
			if declared[code] {
				continue
			}
			if _, exists := existing[code]; !exists {
				continue // already gone; the state forgets it
			}
			if !kind.deletable {
				deletes = append(deletes, &applyChange{Action: applyRelease, Kind: kind.label, Code: code, kind: kind, Note: "the API cannot delete a " + kind.label + "; left in Lago and no longer managed"})
				continue
			}
			change := &applyChange{Action: applyDelete, Kind: kind.label, Code: code, kind: kind, method: http.MethodDelete, path: kind.path + "/" + url.PathEscape(code)}
			if !prune {
				change.skipped, change.Note = true, "skipped: pass --prune to delete"
			}
			deletes = append(deletes, change)
		}
	}
	// Reverse the top-level deletes so dependants go before what they reference.
	for left, right := 0, len(deletes)-1; left < right; left, right = left+1, right-1 {
		deletes[left], deletes[right] = deletes[right], deletes[left]
	}
	// Writes run in plan order: objects by kind, then children (applied rate cards
	// after the rate cards they name), then deletes.
	plan := append([]*applyChange{}, upserts...)
	plan = append(plan, childUpserts...)
	plan = append(plan, childDeletes...)
	plan = append(plan, deletes...)
	return plan, nil
}

func planApplyChildren(kind *applyKind, object *applyObject, existing map[string]map[string]any, parentExists, prune bool) ([]*applyChange, []*applyChange, error) {
	child := kind.child
	base := kind.path + "/" + url.PathEscape(object.code) + "/" + child.segment
	var upserts, deletes []*applyChange
	declared := map[string]bool{}
	for _, item := range object.children {
		declared[item.identity] = true
		current, exists := existing[item.identity]
		if !exists {
			upserts = append(upserts, &applyChange{Action: applyCreate, Kind: child.label, Code: item.identity, Parent: object.code, child: child, method: http.MethodPost, path: base, body: copyFields(item.fields)})
			continue
		}
		fields, err := diffApplyFields(child.label, item.identity, current, item.fields, child.updateFields)
		if err != nil {
			return nil, nil, err
		}
		change := &applyChange{Kind: child.label, Code: item.identity, Parent: object.code, child: child, Fields: fields}
		if len(fields) == 0 {
			change.Action = applyUnchanged
		} else {
			change.Action, change.method, change.path, change.body = applyUpdate, http.MethodPut, base+"/"+url.PathEscape(item.identity), pickFields(item.fields, fields)
		}
		upserts = append(upserts, change)
	}
	if !parentExists {
		return upserts, nil, nil
	}
	for _, identity := range sortedKeysOf(existing) {
		if declared[identity] || !child.deletable(existing[identity]) {
			continue
		}
		change := &applyChange{Action: applyDelete, Kind: child.label, Code: identity, Parent: object.code, child: child, method: http.MethodDelete, path: base + "/" + url.PathEscape(identity)}
		if !prune {
			change.skipped, change.Note = true, "skipped: pass --prune to delete"
		}
		deletes = append(deletes, change)
	}
	return upserts, deletes, nil
}

// diffApplyFields returns the declared fields whose value differs from Lago's. A field
// the API does not return (tax_codes, a rate's effective_from) cannot be compared and
// is only sent on create. A difference on a field the API will not update is an error:
// the object has to be replaced under a new code.
func diffApplyFields(label, code string, current, desired map[string]any, updatable []string) ([]string, error) {
	var changed []string
	for _, key := range sortedKeysOf(desired) {
		if key == "code" {
			continue
		}
		existing, observable := current[key]
		if !observable || applyValuesEqual(existing, desired[key]) {
			continue
		}
		if !contains(updatable, key) {
			return nil, apperr.New(apperr.ExitUsage,
				fmt.Sprintf("%s %q: %s cannot change in place (Lago has %s, the files declare %s)", label, code, key, applyValueText(existing), applyValueText(desired[key])),
				"Declare a new object under a new code, or restore the current value.")
		}
		changed = append(changed, key)
	}
	return changed, nil
}

func checkApplyReferences(desired applyDesired, remote *applyRemote) error {
	declared := map[string]map[string]bool{}
	for key, objects := range desired {
		declared[key] = map[string]bool{}
		for _, object := range objects {
			declared[key][object.code] = true
		}
	}
	known := func(kind, code string) bool {
		if declared[kind][code] {
			return true
		}
		_, exists := remote.objects[kind][code]
		return exists
	}
	for _, kind := range applyKinds {
		for _, object := range desired[kind.key] {
			for field, target := range applyReferences[kind.key] {
				code, _ := object.fields[field].(string)
				if code != "" && !known(target, code) {
					return apperr.New(apperr.ExitUsage, fmt.Sprintf("%s %q (%s): %s %q is neither declared nor in Lago", kind.label, object.code, object.source, field, code), "Declare it in the folder or fix the code.")
				}
			}
			if kind.child == applyAppliedRateCards {
				for _, child := range object.children {
					if !known("rate_cards", child.identity) {
						return apperr.New(apperr.ExitUsage, fmt.Sprintf("plan %q (%s): rate card %q is neither declared nor in Lago", object.code, object.source, child.identity), "Declare it in the folder or fix the code.")
					}
				}
			}
		}
	}
	return nil
}

// applyValuesEqual compares a value read from Lago with a declared one. Numbers compare
// as exact decimals whether they arrive as JSON numbers or strings ("0.7" and 0.70 are
// equal; the API returns rate amounts as strings). Maps compare as subsets: the API
// adds defaulted keys, such as empty pricing_group_keys, that a file need not repeat.
func applyValuesEqual(existing, desired any) bool {
	// Timestamps compare as instants: the API answers 2026-11-01T00:00:00.000000Z for a
	// declared 2026-11-01T00:00:00Z.
	if left, ok := applyTime(existing); ok {
		if right, ok := applyTime(desired); ok {
			return left.Equal(right)
		}
	}
	if left, ok := applyDecimal(existing); ok {
		if right, ok := applyDecimal(desired); ok {
			return left.Cmp(right) == 0
		}
	}
	switch typed := desired.(type) {
	case map[string]any:
		current, ok := existing.(map[string]any)
		if !ok {
			return false
		}
		for key, value := range typed {
			if !applyValuesEqual(current[key], value) {
				return false
			}
		}
		return true
	case []any:
		current, ok := existing.([]any)
		if !ok || len(current) != len(typed) {
			return false
		}
		for index := range typed {
			if !applyValuesEqual(current[index], typed[index]) {
				return false
			}
		}
		return true
	case nil:
		return existing == nil
	case json.Number:
		return false // existing is not a number, or applyDecimal above would have matched
	default:
		return fmt.Sprint(existing) == fmt.Sprint(desired) && sameScalarKind(existing, desired)
	}
}

func sameScalarKind(left, right any) bool {
	switch left.(type) {
	case bool:
		_, ok := right.(bool)
		return ok
	case string:
		_, ok := right.(string)
		return ok
	}
	return false
}

func applyTime(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	return parsed, err == nil
}

func applyDecimal(value any) (*big.Rat, bool) {
	var text string
	switch typed := value.(type) {
	case json.Number:
		text = typed.String()
	case string:
		text = typed
	default:
		return nil, false
	}
	if !jsonNumber.MatchString(text) {
		return nil, false
	}
	rat, ok := new(big.Rat).SetString(text)
	return rat, ok
}

func applyValueText(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(encoded)
}

// --- output ------------------------------------------------------------------------

func renderApplyPlan(app *App, changes []*applyChange, prune bool) error {
	counts := map[string]int{}
	for _, change := range changes {
		if change.skipped {
			counts["skipped"]++
			continue
		}
		counts[change.Action]++
	}
	summary := fmt.Sprintf("Plan: %d to create, %d to update, %d to delete, %d unchanged.", counts[applyCreate], counts[applyUpdate], counts[applyDelete], counts[applyUnchanged])
	if counts["skipped"] > 0 {
		summary += fmt.Sprintf(" %d deletion(s) skipped without --prune.", counts["skipped"])
	}
	if counts[applyRelease] > 0 {
		summary += fmt.Sprintf(" %d object(s) released.", counts[applyRelease])
	}
	mode := app.outputMode()
	if mode != output.Table {
		rows := make([]any, 0, len(changes))
		for _, change := range changes {
			rows = append(rows, change)
		}
		encoded, err := json.Marshal(map[string]any{"changes": rows, "dry_run": app.dryRun, "prune": prune})
		if err != nil {
			return apperr.Wrap(apperr.ExitGeneral, "encode plan", err)
		}
		value, err := transport.DecodeJSON(encoded)
		if err != nil {
			return apperr.Wrap(apperr.ExitGeneral, "encode plan", err)
		}
		if err := app.Renderer().Render(value); err != nil {
			return err
		}
		fmt.Fprintln(app.Err, summary)
		return nil
	}
	writer := tabwriter.NewWriter(app.Out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ACTION\tKIND\tCODE\tPARENT\tCHANGES\tNOTE")
	for _, change := range changes {
		action := change.Action
		if change.skipped {
			action = "(" + action + ")"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			output.Sanitize(action), output.Sanitize(change.Kind), output.Sanitize(change.Code),
			output.Sanitize(change.Parent), output.Sanitize(strings.Join(change.Fields, ",")), output.Sanitize(change.Note))
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(app.Out, summary)
	return nil
}

// --- execution ---------------------------------------------------------------------

// executeApply sends the planned writes in order and stops at the first failure,
// returning the changes that went through so the state can record them.
func executeApply(ctx context.Context, app *App, changes []*applyChange) ([]*applyChange, error) {
	var applied []*applyChange
	for _, change := range changes {
		if change.method == "" || change.skipped {
			continue
		}
		request := transport.Request{Method: change.method, Path: change.path, APIVersion: "v2"}
		if change.body != nil {
			wrapper := ""
			if change.kind != nil {
				wrapper = change.kind.wrapper
			} else {
				wrapper = change.child.wrapper
			}
			body, err := json.Marshal(map[string]any{wrapper: change.body})
			if err != nil {
				return applied, apperr.Wrap(apperr.ExitGeneral, "encode request", err)
			}
			request.Body = body
		}
		// Catalog writes are not replayed: lago-api reads no Idempotency-Key, so a
		// retried create could duplicate and a retried delete could hit a new object.
		if _, _, err := app.Request(ctx, request); err != nil {
			return applied, applyWriteError(change, err, len(applied))
		}
		applied = append(applied, change)
	}
	return applied, nil
}

func applyWriteError(change *applyChange, err error, done int) error {
	subject := fmt.Sprintf("%s %s %q", change.Action, change.Kind, change.Code)
	if change.Parent != "" {
		subject += fmt.Sprintf(" of %q", change.Parent)
	}
	prefix := fmt.Sprintf("%s failed after %d applied change(s)", subject, done)
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		copied := *appErr
		copied.Message = prefix + ": " + appErr.Message
		copied.Suggestion = "Fix the cause and rerun apply: the changes already made show as unchanged."
		return &copied
	}
	return apperr.Wrap(apperr.ExitGeneral, prefix, err)
}

// --- state -------------------------------------------------------------------------

func readApplyState(path, orgID string) (*applyState, error) {
	state := &applyState{Version: applyStateVersion, OrganizationID: orgID, Managed: map[string][]string{}}
	data, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, fs.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitGeneral, "read apply state", err)
	}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("%s is not a valid apply state: %v", path, err), "Restore it from version control, or delete it to start over: apply then adopts the declared objects and prunes nothing.")
	}
	if state.Version != applyStateVersion {
		return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("%s has state version %d; this CLI reads version %d", path, state.Version, applyStateVersion), "Upgrade the Lago CLI.")
	}
	if state.OrganizationID != orgID {
		return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("%s belongs to organization %s, not %s", path, state.OrganizationID, orgID), "Check the profile, or pass --state for this organization.")
	}
	if state.Managed == nil {
		state.Managed = map[string][]string{}
	}
	return state, nil
}

// writeApplyState records what the folder manages after a run: the previously managed
// objects that still exist and were not released or deleted, every declared object
// that exists in Lago, and every object this run created.
func writeApplyState(path string, state *applyState, desired applyDesired, remote *applyRemote, applied []*applyChange) error {
	removed := map[string]bool{}
	created := map[string]bool{}
	for _, change := range applied {
		if change.kind == nil {
			continue
		}
		switch change.Action {
		case applyDelete:
			removed[change.kind.key+"/"+change.Code] = true
		case applyCreate:
			created[change.kind.key+"/"+change.Code] = true
		}
	}
	managed := map[string][]string{}
	for _, kind := range applyKinds {
		set := map[string]bool{}
		for _, code := range state.Managed[kind.key] {
			if _, exists := remote.objects[kind.key][code]; exists {
				set[code] = true
			}
		}
		declared := map[string]bool{}
		for _, object := range desired[kind.key] {
			declared[object.code] = true
			if _, exists := remote.objects[kind.key][object.code]; exists {
				set[object.code] = true
			}
		}
		for code := range set {
			if removed[kind.key+"/"+code] || (!kind.deletable && !declared[code]) {
				delete(set, code)
			}
		}
		for id := range created {
			if prefix := kind.key + "/"; strings.HasPrefix(id, prefix) {
				set[strings.TrimPrefix(id, prefix)] = true
			}
		}
		if len(set) > 0 {
			managed[kind.key] = sortedStrings(set)
		}
	}
	state.Managed = managed
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return apperr.Wrap(apperr.ExitGeneral, "write apply state", err)
	}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return apperr.Wrap(apperr.ExitGeneral, "encode apply state", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return apperr.Wrap(apperr.ExitGeneral, "write apply state", err)
	}
	return nil
}

// --- helpers -----------------------------------------------------------------------

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func copyFields(fields map[string]any) map[string]any {
	copied := make(map[string]any, len(fields))
	for key, value := range fields {
		copied[key] = value
	}
	return copied
}

func pickFields(fields map[string]any, keys []string) map[string]any {
	picked := make(map[string]any, len(keys))
	for _, key := range keys {
		picked[key] = fields[key]
	}
	return picked
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func sortedStrings(set map[string]bool) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func sortedKeysOf[V any](object map[string]V) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
