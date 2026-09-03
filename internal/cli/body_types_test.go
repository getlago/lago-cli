package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/generated"
)

// QA type coercion: the moneycheck-style gate for request bodies. Every generated write
// is run with a synthetic value for each scalar field, and the JSON kind that reaches
// the server must match the field's spec type: an integer field is a JSON number, not
// "1"; a boolean is true/false, not "false". A stringified amount is exactly the kind of
// drift that passes a schema-less server and corrupts a ledger later.
func TestQA_TypeCoercion_GeneratedBodiesSendSpecTypes(t *testing.T) {
	t.Parallel()
	type expectation struct {
		operationID string
		body        *generated.Body
	}
	expectations := make(chan expectation, 1)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		expected := <-expectations
		raw, _ := io.ReadAll(request.Body)
		var body map[string]any
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&body); err != nil {
			t.Errorf("%s: body is not JSON: %v", expected.operationID, err)
		}
		for _, field := range expected.body.Fields {
			if field.Complex {
				continue
			}
			value, present := lookupPath(body, expected.body.Wrapper, field.Path)
			if !present {
				continue // a flag that collides with a parameter name is not set by this test
			}
			switch field.Type {
			case "integer", "number":
				if _, ok := value.(json.Number); !ok {
					t.Errorf("%s --%s (%s) was sent as %T %v, want a JSON number", expected.operationID, field.Flag, field.Type, value, value)
				}
			case "boolean":
				if _, ok := value.(bool); !ok {
					t.Errorf("%s --%s (boolean) was sent as %T %v", expected.operationID, field.Flag, value, value)
				}
			case "string":
				if _, ok := value.(string); !ok {
					t.Errorf("%s --%s (string) was sent as %T %v", expected.operationID, field.Flag, value, value)
				}
			}
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{}`)
	}))
	defer server.Close()

	checked := 0
	for _, operation := range generated.Operations {
		if operation.Body == nil || len(operation.Body.Fields) == 0 || operation.Method == http.MethodGet {
			continue
		}
		operation := operation
		t.Run(operation.OperationID, func(t *testing.T) {
			app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
			app.loaded = true
			app.resolved = config.Resolved{Name: "types", Profile: config.Profile{Region: config.RegionSelf, APIURL: server.URL, APIKey: "fake-key", Mode: config.ModeTest, Insecure: true}}
			command := newGeneratedCommand(app, operation)
			arguments := make([]string, 0)
			for _, parameter := range filterParameters(operation.Parameters, "path") {
				arguments = append(arguments, "fake_"+parameter.Name)
			}
			parameterFlags := map[string]bool{}
			for _, parameter := range operation.Parameters {
				parameterFlags[parameter.Flag] = true
				if parameter.In == "query" && parameter.Required {
					arguments = append(arguments, "--"+parameter.Flag, syntheticValue(parameter.Type, parameter.Enum))
				}
			}
			for _, field := range operation.Body.Fields {
				if parameterFlags[field.Flag] || (field.Complex && !field.Required) {
					continue
				}
				value := syntheticValue(field.Type, field.Enum)
				if operation.Resource == "events" && field.Flag == "timestamp" {
					// The CLI parses this field client-side (Unix seconds or RFC 3339)
					// before sending, so a placeholder string never reaches the wire.
					value = "1788338088"
				}
				arguments = append(arguments, "--"+field.Flag, value)
			}
			if operation.Dangerous {
				identifier := operation.Resource
				if pathParameters := filterParameters(operation.Parameters, "path"); len(pathParameters) > 0 {
					identifier = "fake_" + pathParameters[len(pathParameters)-1].Name
				}
				app.confirm = identifier
			}
			expectations <- expectation{operationID: operation.OperationID, body: operation.Body}
			command.SetArgs(arguments)
			command.SetOut(io.Discard)
			command.SetErr(io.Discard)
			if err := command.Execute(); err != nil {
				<-expectations
				t.Fatalf("%s failed: %v (args %v)", operation.OperationID, err, arguments)
			}
			checked++
		})
	}
	if checked < 60 {
		t.Fatalf("checked only %d operations with body fields", checked)
	}
}

// lookupPath walks a decoded body to a field, through the wrapper key when present.
func lookupPath(body map[string]any, wrapper string, path []string) (any, bool) {
	var current any = body
	if wrapper != "" {
		current = body[wrapper]
	}
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
