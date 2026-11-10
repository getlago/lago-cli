package cli

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/getlago/lago-cli/internal/generated"
)

// Every v2 command must reach /api/v2, never the v1 endpoint of the same name.
func TestV2CommandsCallAPIV2(t *testing.T) {
	var mutex sync.Mutex
	var paths []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		paths = append(paths, request.Method+" "+request.URL.Path)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"product":{"lago_id":"p_1","code":"compute"}}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "v2", "products", "get", "compute"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute(t, "", "v2", "products", "create", "--code", "compute", "--name", "Compute", "--product-type", "fixed"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute(t, "", "v2", "rate-card-rates", "list", "card_1"); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	want := []string{"GET /api/v2/products/compute", "POST /api/v2/products", "GET /api/v2/rate_cards/card_1/rates"}
	if strings.Join(paths, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %v, want %v", paths, want)
	}
}

// --all must follow meta.next_cursor until it is null, ask for the largest page, and
// drop include_total_count after the first page because v2 rejects it there.
func TestV2AllFollowsNextCursor(t *testing.T) {
	var mutex sync.Mutex
	var queries []string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		mutex.Lock()
		queries = append(queries, query.Encode())
		mutex.Unlock()
		if query.Get("limit") != "100" {
			t.Errorf("--all did not request a full page: %s", request.URL.RawQuery)
		}
		switch query.Get("after") {
		case "":
			_, _ = response.Write([]byte(`{"products":[{"code":"p1"}],"meta":{"next_cursor":"c2","prev_cursor":null,"total_count":3}}`))
		case "c2":
			if query.Has("include_total_count") {
				t.Errorf("include_total_count was sent after the first page: %s", request.URL.RawQuery)
			}
			_, _ = response.Write([]byte(`{"products":[{"code":"p2"}],"meta":{"next_cursor":"c3","prev_cursor":"c1"}}`))
		default:
			_, _ = response.Write([]byte(`{"products":[{"code":"p3"}],"meta":{"next_cursor":null,"prev_cursor":"c2"}}`))
		}
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "--output", "json", "v2", "products", "list", "--all", "--include-total-count", "true")
	if err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if len(queries) != 3 {
		t.Fatalf("requests = %v, want 3", queries)
	}
	for _, want := range []string{"p1", "p2", "p3"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("output is missing %s:\n%s", want, stdout)
		}
	}
}

// A server that keeps returning the same cursor must not loop the CLI forever.
func TestV2AllStopsOnARepeatedCursor(t *testing.T) {
	var mutex sync.Mutex
	requests := 0
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		requests++
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"products":[],"meta":{"next_cursor":"same","prev_cursor":null}}`))
	})
	profileAt(t, server.URL)
	if _, _, err := execute(t, "", "--output", "json", "v2", "products", "list", "--all"); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestV2AllRefusesBefore(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) { t.Error("request sent") })
	profileAt(t, server.URL)
	if _, _, err := execute(t, "", "v2", "products", "list", "--all", "--before", "c1"); err == nil || !strings.Contains(err.Error(), "--before") {
		t.Fatalf("err = %v", err)
	}
}

// A v2 validation error must print the field codes, not the raw JSON body.
func TestV2ErrorIsReadable(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = response.Write([]byte(`{"code":"validation_errors","details":[{"code":"value_is_mandatory","field":"name","meta":{}}]}`))
	})
	profileAt(t, server.URL)
	_, _, err := execute(t, "", "v2", "products", "create", "--code", "c", "--name", "n", "--product-type", "fixed")
	if err == nil || !strings.Contains(err.Error(), "name: value_is_mandatory") {
		t.Fatalf("err = %v", err)
	}
}

// `lago api` on a v2 path must be classified against the v2 operation table, so a
// live-mode delete is gated and a v2-only read is not mistaken for anything else.
func TestRawV2RequestsUseTheV2DangerTable(t *testing.T) {
	t.Parallel()
	dangerous, operationID := classifyRequest(http.MethodDelete, "/api/v2/products/compute")
	if !dangerous || operationID != "destroyProduct" {
		t.Errorf("DELETE /api/v2/products/compute = %v %q", dangerous, operationID)
	}
	dangerous, operationID = classifyRequest(http.MethodGet, "/api/v2/rate_cards/card_1/rates?limit=5")
	if dangerous || operationID == "" {
		t.Errorf("GET rate card rates = %v %q", dangerous, operationID)
	}
	if _, operationID = classifyRequest(http.MethodGet, "/customers"); operationID != "findAllCustomers" {
		t.Errorf("v1 path matched %q", operationID)
	}
}

// The v2 table is generated from its own pinned document. It must be non-empty, carry
// the product catalog, and mark every operation as v2 with cursor lists detected.
func TestV2OperationTable(t *testing.T) {
	t.Parallel()
	if generated.SpecV2Version == "" || len(generated.OperationsV2) == 0 {
		t.Fatal("v2 operation table is empty")
	}
	resources := map[string]bool{}
	cursor := 0
	for _, operation := range generated.OperationsV2 {
		if operation.APIVersion != "v2" {
			t.Errorf("%s has APIVersion %q", operation.OperationID, operation.APIVersion)
		}
		if operation.Paginated {
			t.Errorf("%s is page-paginated; v2 lists use cursors", operation.OperationID)
		}
		if operation.CursorPaginated {
			cursor++
		}
		resources[operation.Resource] = true
	}
	for _, resource := range []string{"products", "product-categories", "product-filters", "rate-cards", "rate-card-rates", "catalog-plans", "contracts"} {
		if !resources[resource] {
			t.Errorf("v2 table is missing %s", resource)
		}
	}
	if cursor == 0 {
		t.Error("no v2 list was detected as cursor-paginated")
	}
	for _, operation := range generated.Operations {
		if operation.APIVersion != "" || operation.CursorPaginated {
			t.Errorf("v1 operation %s carries v2 metadata", operation.OperationID)
		}
	}
}

func TestV2ExampleNamesTheNamespace(t *testing.T) {
	t.Parallel()
	for _, operation := range generated.OperationsV2 {
		example := generatedExample(operation, filterParameters(operation.Parameters, "path"))
		if !strings.HasPrefix(example, fmt.Sprintf("  lago v2 %s %s", operation.Resource, operation.Action)) {
			t.Fatalf("example for %s = %q", operation.OperationID, example)
		}
	}
}
