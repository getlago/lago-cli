package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// fakeCatalog is an in-memory stand-in for the lago-api v2 catalog endpoints apply
// uses. Lists page two items at a time so every list goes through the cursor walk.
type fakeCatalog struct {
	t        *testing.T
	mutex    sync.Mutex
	objects  map[string]map[string]map[string]any // collection -> code -> object
	children map[string]map[string]map[string]any // "rate_cards/<code>/rates" -> identity -> child
	writes   []string
	bodies   map[string]map[string]any
	failOn   string // "METHOD /path" answered with a 422
}

var fakeWrappers = map[string]string{
	"billable_metrics": "billable_metric", "product_categories": "product_category", "products": "product",
	"rate_cards": "rate_card", "plans": "plan", "rates": "rate", "applied_rate_cards": "applied_rate_card",
}

func newFakeCatalog(t *testing.T) *fakeCatalog {
	catalog := &fakeCatalog{t: t, objects: map[string]map[string]map[string]any{}, children: map[string]map[string]map[string]any{}, bodies: map[string]map[string]any{}}
	for _, kind := range applyKinds {
		catalog.objects[kind.key] = map[string]map[string]any{}
	}
	return catalog
}

func (c *fakeCatalog) put(collection, code string, object map[string]any) {
	object["code"] = code
	c.objects[collection][code] = object
}

func (c *fakeCatalog) putChild(parentCollection, parent, collection, identity string, object map[string]any) {
	key := parentCollection + "/" + parent + "/" + collection
	if c.children[key] == nil {
		c.children[key] = map[string]map[string]any{}
	}
	if collection == "rates" {
		object["code"] = identity
		if object["status"] == nil {
			object["status"] = "active"
		}
	} else {
		object["rate_card_code"] = identity
	}
	c.children[key][identity] = object
}

func (c *fakeCatalog) server() string {
	server := jsonAPI(c.t, func(response http.ResponseWriter, request *http.Request) {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		path := strings.TrimPrefix(request.URL.Path, "/api/v2/")
		if path == request.URL.Path {
			c.t.Errorf("request outside /api/v2: %s", request.URL.Path)
		}
		line := request.Method + " /" + path
		if request.Method != http.MethodGet {
			c.writes = append(c.writes, line)
			if body, _ := io.ReadAll(request.Body); len(body) > 0 {
				var decoded map[string]any
				_ = json.Unmarshal(body, &decoded)
				c.bodies[line] = decoded
			}
		}
		if line == c.failOn {
			response.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = io.WriteString(response, `{"status":422,"error":"Unprocessable Entity","code":"validation_errors","error_details":{"name":["value_is_mandatory"]}}`)
			return
		}
		if path == "organizations" {
			_, _ = io.WriteString(response, `{"organization":{"lago_id":"org_1","name":"Test org"}}`)
			return
		}
		parts := strings.Split(path, "/")
		var store map[string]map[string]any
		var collection string
		var identity string
		switch len(parts) {
		case 1, 2:
			collection = parts[0]
			store = c.objects[collection]
			if len(parts) == 2 {
				identity = parts[1]
			}
		case 3, 4:
			collection = parts[2]
			key := parts[0] + "/" + parts[1] + "/" + parts[2]
			if c.children[key] == nil {
				c.children[key] = map[string]map[string]any{}
			}
			store = c.children[key]
			if len(parts) == 4 {
				identity = parts[3]
			}
		}
		if store == nil {
			response.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(response, `{"status":404,"error":"Not Found","code":"not_found"}`)
			return
		}
		wrapper := fakeWrappers[collection]
		switch request.Method {
		case http.MethodGet:
			codes := make([]string, 0, len(store))
			for code := range store {
				codes = append(codes, code)
			}
			sort.Strings(codes)
			start := 0
			if after := request.URL.Query().Get("after"); after != "" {
				_, _ = fmt.Sscan(after, &start)
			}
			end := min(start+2, len(codes))
			items := make([]any, 0, 2)
			for _, code := range codes[start:end] {
				items = append(items, store[code])
			}
			var next any
			if end < len(codes) {
				next = fmt.Sprint(end)
			}
			encoded, _ := json.Marshal(map[string]any{collection: items, "meta": map[string]any{"next_cursor": next}})
			_, _ = response.Write(encoded)
		case http.MethodPost:
			recorded, _ := c.bodies[line][wrapper].(map[string]any)
			fields := copyFields(recorded)
			key := "code"
			if collection == "applied_rate_cards" {
				key = "rate_card_code"
			}
			code, _ := fields[key].(string)
			if collection == "rate_cards" {
				rates, _ := fields["rates"].([]any)
				delete(fields, "rates")
				for _, raw := range rates {
					rate := raw.(map[string]any)
					c.putChild("rate_cards", code, "rates", rate["code"].(string), rate)
				}
			}
			if collection == "rates" {
				fields["status"] = "pending"
			}
			store[code] = fields
			encoded, _ := json.Marshal(map[string]any{wrapper: fields})
			_, _ = response.Write(encoded)
		case http.MethodPut:
			fields, _ := c.bodies[line][wrapper].(map[string]any)
			for key, value := range fields {
				store[identity][key] = value
			}
			encoded, _ := json.Marshal(map[string]any{wrapper: store[identity]})
			_, _ = response.Write(encoded)
		case http.MethodDelete:
			delete(store, identity)
			_, _ = io.WriteString(response, `{}`)
		}
	})
	profileAt(c.t, server.URL)
	return server.URL
}

func (c *fakeCatalog) takeWrites() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	writes := c.writes
	c.writes = nil
	return writes
}

func writeCatalogFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const catalogMetrics = `
billable_metrics:
  - code: billing_volume
    name: Billing volume
    aggregation_type: sum_agg
    field_name: total_usd_billing_volume
    recurring: false
`

const catalogProducts = `
product_categories:
  - code: usage
    name: Usage
products:
  - code: billing_volume
    name: Billing volume
    product_type: metered
    product_category_code: usage
    billable_metric_code: billing_volume
`

const catalogRateCards = `
rate_cards:
  - code: billing_volume_usd
    product_code: billing_volume
    name: Billing volume (USD)
    currency: USD
    billing_timing: arrears
    rates:
      - code: bv_2026_11
        effective_from: 2026-11-01T00:00:00Z
        rate_model: percentage
        rate_properties:
          rate: 0.7
        billing_interval_unit: month
        billing_interval_count: 1
`

const catalogPlans = `
plans:
  - code: launch
    name: Launch
    currency: USD
    applied_rate_cards:
      - rate_card_code: billing_volume_usd
        units: "1"
`

func catalogFiles() map[string]string {
	return map[string]string{
		"metrics.yml":           catalogMetrics,
		"catalog/products.yaml": catalogProducts,
		"catalog/rates.yml":     catalogRateCards,
		"plans.yml":             catalogPlans,
		"README.md":             "not a catalog file",
	}
}

// A first apply creates every object in dependency order, records them in the state,
// and a second apply finds nothing to do.
func TestApplyCreatesInDependencyOrderThenConverges(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	dir := writeCatalogFiles(t, catalogFiles())

	stdout, stderr, err := execute(t, "", "v2", "apply", dir)
	if err != nil {
		t.Fatalf("apply: %v\n%s", err, stderr)
	}
	want := []string{
		"POST /billable_metrics", "POST /product_categories", "POST /products", "POST /rate_cards", "POST /plans",
		"POST /plans/launch/applied_rate_cards",
	}
	if got := catalog.takeWrites(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("writes = %v, want %v", got, want)
	}
	if !strings.Contains(stdout, "Plan: 6 to create, 0 to update, 0 to delete, 0 unchanged.") || !strings.Contains(stdout, "with 1 rate(s)") {
		t.Fatalf("plan output:\n%s", stdout)
	}
	rateCard := catalog.bodies["POST /rate_cards"]["rate_card"].(map[string]any)
	rate := rateCard["rates"].([]any)[0].(map[string]any)
	if fmt.Sprint(rate["rate_properties"]) != "map[rate:0.7]" || rate["effective_from"] != "2026-11-01T00:00:00Z" {
		t.Fatalf("rate body = %v", rate)
	}

	stateData, err := os.ReadFile(filepath.Join(dir, ".lago-state", "org_1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state applyState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatal(err)
	}
	if state.OrganizationID != "org_1" || strings.Join(state.Managed["products"], ",") != "billing_volume" || strings.Join(state.Managed["plans"], ",") != "launch" {
		t.Fatalf("state = %+v", state)
	}

	// The API answers numbers and timestamps in its own format; they still match.
	catalog.children["rate_cards/billing_volume_usd/rates"]["bv_2026_11"]["effective_from"] = "2026-11-01T00:00:00.000000Z"
	catalog.children["rate_cards/billing_volume_usd/rates"]["bv_2026_11"]["rate_properties"] = map[string]any{"rate": "0.70", "pricing_group_keys": []any{}}
	catalog.children["plans/launch/applied_rate_cards"]["billing_volume_usd"]["units"] = "1.0"
	_, stderr, err = execute(t, "", "v2", "apply", dir)
	if err != nil {
		t.Fatal(err)
	}
	if writes := catalog.takeWrites(); len(writes) != 0 || !strings.Contains(stderr, "Nothing to apply.") {
		t.Fatalf("second apply wrote %v, stderr %s", writes, stderr)
	}
}

func TestApplyDryRunSendsNoWrite(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	dir := writeCatalogFiles(t, catalogFiles())
	stdout, stderr, err := execute(t, "", "--dry-run", "v2", "apply", dir)
	if err != nil {
		t.Fatal(err)
	}
	if writes := catalog.takeWrites(); len(writes) != 0 {
		t.Fatalf("dry run wrote %v", writes)
	}
	if !strings.Contains(stdout, "6 to create") || !strings.Contains(stderr, "Dry run: no request was sent.") {
		t.Fatalf("stdout=%s stderr=%s", stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, ".lago-state")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote the state: %v", err)
	}
}

// An existing object is adopted and updated with only the fields that differ.
func TestApplyUpdatesOnlyChangedFields(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.put("billable_metrics", "billing_volume", map[string]any{"name": "Old name", "aggregation_type": "sum_agg", "field_name": "total_usd_billing_volume", "recurring": false, "description": nil})
	catalog.put("product_categories", "usage", map[string]any{"name": "Usage"})
	catalog.server()
	dir := writeCatalogFiles(t, map[string]string{"metrics.yml": catalogMetrics, "categories.yml": "product_categories:\n  - code: usage\n    name: Usage\n"})

	stdout, _, err := execute(t, "", "v2", "apply", dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := catalog.takeWrites(); strings.Join(got, ",") != "PUT /billable_metrics/billing_volume" {
		t.Fatalf("writes = %v", got)
	}
	body := catalog.bodies["PUT /billable_metrics/billing_volume"]["billable_metric"].(map[string]any)
	if len(body) != 1 || body["name"] != "Billing volume" {
		t.Fatalf("update body = %v, want only the name", body)
	}
	if !strings.Contains(stdout, "adopted: already in Lago") || !strings.Contains(stdout, "unchanged") {
		t.Fatalf("plan:\n%s", stdout)
	}
}

func TestApplyRefusesInvalidFolders(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.put("product_categories", "usage", map[string]any{"name": "Usage"})
	catalog.put("billable_metrics", "billing_volume", map[string]any{"name": "Billing volume", "aggregation_type": "count_agg"})
	catalog.put("products", "billing_volume", map[string]any{"name": "Billing volume", "product_type": "fixed", "product_category_code": "usage"})
	catalog.server()

	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"immutable field", map[string]string{"p.yml": catalogProducts}, `product_type cannot change in place (Lago has "fixed", the files declare "metered")`},
		{"unknown field", map[string]string{"m.yml": "billable_metrics:\n  - code: x\n    colour: red\n"}, `unknown field "colour"`},
		{"unknown key", map[string]string{"m.yml": "coupons: []\n"}, `unknown key "coupons"`},
		{"not a list", map[string]string{"m.yml": "products: {}\n"}, "products must be a list"},
		{"not a map", map[string]string{"m.yml": "- a\n"}, "the document is not a map"},
		{"missing code", map[string]string{"m.yml": "products:\n  - name: x\n"}, "code is required"},
		{"duplicate", map[string]string{"a.yml": catalogMetrics, "b.yml": catalogMetrics}, `billable metric "billing_volume" is declared twice`},
		{"bad reference", map[string]string{"r.yml": strings.Replace(catalogRateCards, "product_code: billing_volume", "product_code: nope", 1)}, `product_code "nope" is neither declared nor in Lago`},
		{"bad plan reference", map[string]string{"p.yml": catalogPlans}, `rate card "billing_volume_usd" is neither declared nor in Lago`},
		{"odd number", map[string]string{"m.yml": "billable_metrics:\n  - code: x\n    rounding_precision: 1_000\n"}, "is not a plain decimal number"},
		{"repeated key", map[string]string{"m.yml": "billable_metrics:\n  - code: x\n    code: y\n"}, `key "code" is repeated`},
		{"bad yaml", map[string]string{"m.yml": "billable_metrics: [\n"}, "m.yml: yaml:"},
		{"child not a list", map[string]string{"p.yml": "plans:\n  - code: x\n    applied_rate_cards: {}\n"}, "applied_rate_cards must be a list"},
		{"child without identity", map[string]string{"p.yml": "plans:\n  - code: x\n    applied_rate_cards:\n      - units: 1\n"}, "rate_card_code is required"},
		{"child unknown field", map[string]string{"p.yml": "plans:\n  - code: x\n    applied_rate_cards:\n      - rate_card_code: a\n        rate_phases: []\n"}, `unknown field "rate_phases"`},
		{"child duplicate", map[string]string{"p.yml": "plans:\n  - code: x\n    applied_rate_cards:\n      - rate_card_code: a\n      - rate_card_code: a\n"}, "declared twice"},
		{"empty folder", map[string]string{"notes.txt": "x"}, "no .yml or .yaml file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeCatalogFiles(t, tc.files)
			_, _, err := execute(t, "", "v2", "apply", dir)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
			if writes := catalog.takeWrites(); len(writes) != 0 {
				t.Fatalf("an invalid folder wrote %v", writes)
			}
		})
	}
}

// --prune deletes the managed objects the files dropped, children first, and never an
// object the folder does not manage. Active rates are history and stay.
func TestApplyPruneDeletesOnlyManagedObjects(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	extra := `
billable_metrics:
  - code: tmp_metric
    name: Tmp
    aggregation_type: count_agg
products:
  - code: tmp_product
    name: Tmp
    product_type: metered
    product_category_code: usage
    billable_metric_code: tmp_metric
`
	files := catalogFiles()
	files["tmp.yml"] = extra
	dir := writeCatalogFiles(t, files)
	if _, _, err := execute(t, "", "v2", "apply", dir); err != nil {
		t.Fatal(err)
	}
	catalog.takeWrites()
	// Objects someone else made, and a pending rate added outside the folder.
	catalog.put("products", "someone_elses", map[string]any{"name": "Not ours", "product_type": "fixed"})
	catalog.putChild("rate_cards", "billing_volume_usd", "rates", "bv_2027_01", map[string]any{"status": "pending"})
	catalog.putChild("plans", "launch", "applied_rate_cards", "other_card", map[string]any{})

	if err := os.Remove(filepath.Join(dir, "tmp.yml")); err != nil {
		t.Fatal(err)
	}
	// The active rate is dropped from the file too: it must be kept.
	if err := os.WriteFile(filepath.Join(dir, "catalog", "rates.yml"), []byte(strings.SplitN(catalogRateCards, "    rates:", 2)[0]), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := execute(t, "", "v2", "apply", dir)
	if err != nil {
		t.Fatalf("%v %s", err, stderr)
	}
	if writes := catalog.takeWrites(); len(writes) != 0 {
		t.Fatalf("apply without --prune wrote %v", writes)
	}
	if strings.Count(stdout, "skipped: pass --prune to delete") != 4 || !strings.Contains(stdout, "4 deletion(s) skipped") {
		t.Fatalf("plan without --prune:\n%s", stdout)
	}

	if _, _, err := execute(t, "", "v2", "apply", dir, "--prune"); err == nil || !strings.Contains(err.Error(), "confirmation required for prune") {
		t.Fatalf("unconfirmed prune: %v", err)
	}
	if writes := catalog.takeWrites(); len(writes) != 0 {
		t.Fatalf("unconfirmed prune wrote %v", writes)
	}

	if _, stderr, err := execute(t, "", "--confirm", "prune", "v2", "apply", dir, "--prune"); err != nil {
		t.Fatalf("%v %s", err, stderr)
	}
	want := []string{
		"DELETE /plans/launch/applied_rate_cards/other_card",
		"DELETE /rate_cards/billing_volume_usd/rates/bv_2027_01",
		"DELETE /products/tmp_product",
		"DELETE /billable_metrics/tmp_metric",
	}
	if got := catalog.takeWrites(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("prune writes = %v, want %v", got, want)
	}
	if catalog.objects["products"]["someone_elses"] == nil || catalog.children["rate_cards/billing_volume_usd/rates"]["bv_2026_11"] == nil {
		t.Fatal("prune deleted an object the folder does not manage, or an active rate")
	}
}

// A plan has no delete route: dropping it releases it from the state.
func TestApplyReleasesDroppedPlan(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	files := catalogFiles()
	dir := writeCatalogFiles(t, files)
	if _, _, err := execute(t, "", "v2", "apply", dir); err != nil {
		t.Fatal(err)
	}
	catalog.takeWrites()
	if err := os.Remove(filepath.Join(dir, "plans.yml")); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := execute(t, "", "--confirm", "prune", "v2", "apply", dir, "--prune")
	if err != nil {
		t.Fatal(err)
	}
	if writes := catalog.takeWrites(); len(writes) != 0 {
		t.Fatalf("release wrote %v", writes)
	}
	if !strings.Contains(stdout, "the API cannot delete a plan") || !strings.Contains(stdout, "1 object(s) released") || !strings.Contains(stderr, "Nothing to apply.") {
		t.Fatalf("stdout=%s stderr=%s", stdout, stderr)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".lago-state", "org_1.json"))
	if strings.Contains(string(data), "launch") {
		t.Fatalf("released plan is still in the state: %s", data)
	}
}

// A failed write stops the apply, names what failed and how far it got, and the state
// keeps what was created so the rerun only does the rest.
func TestApplyStopsAtFirstFailureAndRecordsProgress(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.failOn = "POST /rate_cards"
	catalog.server()
	dir := writeCatalogFiles(t, catalogFiles())
	_, _, err := execute(t, "", "v2", "apply", dir)
	if err == nil || !strings.Contains(err.Error(), `create rate card "billing_volume_usd" failed after 3 applied change(s)`) {
		t.Fatalf("err = %v", err)
	}
	if got := catalog.takeWrites(); len(got) != 4 {
		t.Fatalf("writes = %v, want 3 creates and the failed one", got)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".lago-state", "org_1.json"))
	if !strings.Contains(string(data), `"billing_volume"`) || strings.Contains(string(data), "rate_cards") {
		t.Fatalf("state after failure = %s", data)
	}

	catalog.failOn = ""
	if _, _, err := execute(t, "", "v2", "apply", dir); err != nil {
		t.Fatal(err)
	}
	if got := catalog.takeWrites(); strings.Join(got, ",") != "POST /rate_cards,POST /plans,POST /plans/launch/applied_rate_cards" {
		t.Fatalf("rerun writes = %v", got)
	}
}

func TestApplyJSONOutputAndChildUpdates(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	dir := writeCatalogFiles(t, catalogFiles())
	if _, _, err := execute(t, "", "v2", "apply", dir); err != nil {
		t.Fatal(err)
	}
	catalog.takeWrites()
	plans := strings.Replace(catalogPlans, `units: "1"`, `units: "2"`, 1)
	rates := strings.Replace(catalogRateCards, "rate: 0.7", "rate: 0.6", 1) + `      - code: bv_2027_01
        effective_from: 2027-01-01T00:00:00Z
        rate_model: percentage
        rate_properties: {rate: "0.5"}
        billing_interval_unit: month
        billing_interval_count: 1
`
	for name, content := range map[string]string{"plans.yml": plans, filepath.Join("catalog", "rates.yml"): rates} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	stdout, _, err := execute(t, "", "--output", "json", "v2", "apply", dir)
	if err != nil {
		t.Fatal(err)
	}
	var plan struct {
		Changes []applyChange `json:"changes"`
	}
	if err := json.Unmarshal([]byte(stdout), &plan); err != nil {
		t.Fatalf("%v\n%s", err, stdout)
	}
	var summary []string
	for _, change := range plan.Changes {
		if change.Action != applyUnchanged {
			summary = append(summary, change.Action+" "+change.Kind+" "+change.Code+" "+strings.Join(change.Fields, "+"))
		}
	}
	want := "update rate bv_2026_11 rate_properties,create rate bv_2027_01 ,update applied rate card billing_volume_usd units"
	if strings.Join(summary, ",") != want {
		t.Fatalf("changes = %v, want %s", summary, want)
	}
	want = "PUT /rate_cards/billing_volume_usd/rates/bv_2026_11,POST /rate_cards/billing_volume_usd/rates,PUT /plans/launch/applied_rate_cards/billing_volume_usd"
	if got := catalog.takeWrites(); strings.Join(got, ",") != want {
		t.Fatalf("writes = %v", got)
	}
}

func TestApplyStateMustMatchOrganization(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	dir := writeCatalogFiles(t, map[string]string{"m.yml": catalogMetrics})
	statePath := filepath.Join(t.TempDir(), "state.json")
	for content, want := range map[string]string{
		`{"version":1,"organization_id":"org_2","managed":{}}`: "belongs to organization org_2",
		`{"version":9,"organization_id":"org_1"}`:              "state version 9",
		`not json`: "is not a valid apply state",
	} {
		if err := os.WriteFile(statePath, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := execute(t, "", "v2", "apply", dir, "--state", statePath); err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("state %s: err = %v, want %q", content, err, want)
		}
	}
	if err := os.WriteFile(statePath, []byte(`{"version":1,"organization_id":"org_1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute(t, "", "v2", "apply", dir, "--state", statePath); err != nil {
		t.Fatal(err)
	}
}

func TestApplyAcceptsASingleFile(t *testing.T) {
	catalog := newFakeCatalog(t)
	catalog.server()
	dir := writeCatalogFiles(t, map[string]string{"m.yml": catalogMetrics + "---\n# empty document\n"})
	if _, _, err := execute(t, "", "v2", "apply", filepath.Join(dir, "m.yml"), "--state", filepath.Join(dir, "state.json")); err != nil {
		t.Fatal(err)
	}
	if got := catalog.takeWrites(); strings.Join(got, ",") != "POST /billable_metrics" {
		t.Fatalf("writes = %v", got)
	}
	if _, _, err := execute(t, "", "v2", "apply", filepath.Join(dir, "missing")); err == nil {
		t.Fatal("a missing path was accepted")
	}
}

func TestApplyYAMLValues(t *testing.T) {
	dir := writeCatalogFiles(t, map[string]string{"m.yml": `
defaults: &defaults
  aggregation_type: count_agg
billable_metrics: []
`})
	if _, err := loadApplyDir(dir); err == nil || !strings.Contains(err.Error(), `unknown key "defaults"`) {
		t.Fatalf("err = %v", err)
	}
	dir = writeCatalogFiles(t, map[string]string{"m.yml": `
billable_metrics:
  - &metric
    code: a
    name: A
    recurring: true
    expression: null
    rounding_precision: 2
  - code: b
    name: B
    recurring: false
`})
	desired, err := loadApplyDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fields := desired["billable_metrics"][0].fields
	if fields["recurring"] != true || fields["expression"] != nil || fields["rounding_precision"] != json.Number("2") {
		t.Fatalf("fields = %#v", fields)
	}
}

func TestApplyValuesEqual(t *testing.T) {
	cases := []struct {
		existing, desired any
		equal             bool
	}{
		{"0.70", json.Number("0.7"), true},
		{json.Number("1"), "1.0", true},
		{"2026-11-01T00:00:00.000000Z", "2026-11-01T01:00:00+01:00", true},
		{"2026-11-01T00:00:00Z", "2026-11-02T00:00:00Z", false},
		{map[string]any{"amount": "1", "extra": true}, map[string]any{"amount": json.Number("1")}, true},
		{map[string]any{"amount": "1"}, map[string]any{"amount": "2"}, false},
		{"x", map[string]any{}, false},
		{[]any{"a"}, []any{"a"}, true},
		{[]any{"a"}, []any{"a", "b"}, false},
		{nil, nil, true},
		{"", nil, false},
		{"abc", json.Number("1"), false},
		{true, true, true},
		{"true", true, false},
		{"a", "a", true},
	}
	for _, tc := range cases {
		if got := applyValuesEqual(tc.existing, tc.desired); got != tc.equal {
			t.Errorf("applyValuesEqual(%#v, %#v) = %v, want %v", tc.existing, tc.desired, got, tc.equal)
		}
	}
}
