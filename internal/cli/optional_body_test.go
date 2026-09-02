package cli

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/generated"
)

// QA L-2f, M-optional-body: `invoices void inv_1` was refused with "request body is
// required" although the spec marks the body optional and documents the bodiless call.
func TestQA_MOptionalBody_InvoicesVoidRunsWithoutBodyFlags(t *testing.T) {
	var mutex sync.Mutex
	var method, path, body string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		raw, _ := io.ReadAll(request.Body)
		mutex.Lock()
		method, path, body = request.Method, request.URL.Path, string(raw)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"invoice":{"lago_id":"inv_1","status":"voided"}}`))
	})
	profileAt(t, server.URL)

	if _, _, err := execute(t, "", "--output", "json", "invoices", "void", "inv_1", "--confirm", "inv_1"); err != nil {
		t.Fatalf("invoices void without body flags failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if method != http.MethodPost || path != "/api/v1/invoices/inv_1/void" {
		t.Errorf("request = %s %s", method, path)
	}
	if strings.TrimSpace(body) != "" {
		t.Errorf("bodiless void sent a body: %q", body)
	}
}

// A body the spec marks required is still guarded client-side, at exit 2. The guard is
// exercised on an operation whose body is required but whose fields are all optional,
// so the required-field check does not fire first.
func TestQA_MOptionalBody_RequiredBodyIsStillGuarded(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) {
		t.Error("a write with no body reached the API")
	})
	profileAt(t, server.URL)
	var picked *generated.Operation
	for _, operation := range generated.Operations {
		if operation.Body == nil || !operation.Body.Required || operation.Dangerous || operation.Method == http.MethodGet {
			continue
		}
		required := false
		for _, field := range operation.Body.Fields {
			required = required || field.Required
		}
		if !required && len(filterParameters(operation.Parameters, "path")) == 0 {
			operation := operation
			picked = &operation
			break
		}
	}
	if picked == nil {
		t.Skip("no operation has a required body with only optional fields")
	}
	_, _, err := execute(t, "", picked.Resource, picked.Action)
	if err == nil {
		t.Fatalf("%s with no body was accepted", picked.OperationID)
	}
	if apperr.ExitCode(err) != apperr.ExitUsage || !strings.Contains(err.Error(), "request body is required") {
		t.Errorf("%s: unexpected refusal: %v", picked.OperationID, err)
	}
}

// QA M-subscriptions-u: `subscriptions update` demanded --subscription-ending-at, a
// nullable field the API does not need.
func TestQA_MSubscriptionsU_UpdateWithoutEndingAtIsAccepted(t *testing.T) {
	var mutex sync.Mutex
	var body string
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		raw, _ := io.ReadAll(request.Body)
		mutex.Lock()
		body = string(raw)
		mutex.Unlock()
		_, _ = response.Write([]byte(`{"subscription":{"lago_id":"sub_1","external_id":"s1","name":"Renamed"}}`))
	})
	profileAt(t, server.URL)
	if _, _, err := execute(t, "", "--output", "json", "subscriptions", "update", "s1", "--subscription-name", "Renamed"); err != nil {
		t.Fatalf("subscriptions update without --subscription-ending-at failed: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if !strings.Contains(body, `"name":"Renamed"`) || strings.Contains(body, "ending_at") {
		t.Errorf("body = %s", body)
	}
}
