package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/generated"
)

func TestEveryGeneratedCommandRoundTripsMethodPathAndJSONTypes(t *testing.T) {
	t.Parallel()
	type expectation struct {
		method string
		path   string
		body   *generated.Body
	}
	expectations := make(chan expectation, 1)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		expected := <-expectations
		if request.Method != expected.method {
			t.Errorf("method = %s, want %s", request.Method, expected.method)
		}
		if request.URL.Path != "/api/v1"+expected.path {
			t.Errorf("path = %s, want %s", request.URL.Path, "/api/v1"+expected.path)
		}
		if request.Header.Get("Authorization") != "Bearer fake-key" {
			t.Error("generated command omitted authentication")
		}
		if expected.body != nil {
			var body map[string]any
			decoder := json.NewDecoder(request.Body)
			decoder.UseNumber()
			if err := decoder.Decode(&body); err != nil {
				t.Errorf("request body is not JSON: %v", err)
			}
			if expected.body.Wrapper != "" {
				if _, ok := body[expected.body.Wrapper].(map[string]any); !ok {
					t.Errorf("request body lacks object wrapper %q: %#v", expected.body.Wrapper, body)
				}
			}
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{}`)
	}))
	defer server.Close()

	for _, operation := range generated.Operations {
		operation := operation
		t.Run(operation.OperationID, func(t *testing.T) {
			app := NewApp(strings.NewReader(""), io.Discard, io.Discard, "test")
			app.loaded = true
			app.resolved = config.Resolved{Name: "contract", Profile: config.Profile{Region: config.RegionSelf, APIURL: server.URL, APIKey: "fake-key", Mode: config.ModeTest, Insecure: true}}
			command := newGeneratedCommand(app, operation)
			arguments := make([]string, 0)
			expectedPath := operation.Path
			for _, parameter := range filterParameters(operation.Parameters, "path") {
				value := "fake_" + parameter.Name
				arguments = append(arguments, value)
				expectedPath = strings.ReplaceAll(expectedPath, "{"+parameter.Name+"}", url.PathEscape(value))
			}
			for _, parameter := range operation.Parameters {
				if parameter.In == "query" && parameter.Required {
					arguments = append(arguments, "--"+parameter.Flag, syntheticValue(parameter.Type, parameter.Enum))
				}
			}
			if operation.Body != nil {
				added := false
				for _, field := range operation.Body.Fields {
					if field.Required {
						arguments = append(arguments, "--"+field.Flag, syntheticValue(field.Type, field.Enum))
						added = true
					}
				}
				if !added {
					input := map[string]any{}
					if operation.Body.Wrapper != "" {
						input[operation.Body.Wrapper] = map[string]any{}
					}
					encoded, err := json.Marshal(input)
					if err != nil {
						t.Fatal(err)
					}
					arguments = append(arguments, "--input", string(encoded))
				}
			}
			if operation.Dangerous {
				identifier := operation.Resource
				pathParameters := filterParameters(operation.Parameters, "path")
				if len(pathParameters) > 0 {
					identifier = "fake_" + pathParameters[len(pathParameters)-1].Name
				}
				app.confirm = identifier
			}
			expectations <- expectation{method: operation.Method, path: expectedPath, body: operation.Body}
			command.SetArgs(arguments)
			if err := command.ExecuteContext(context.Background()); err != nil {
				select {
				case <-expectations:
				default:
				}
				t.Fatalf("%s %s failed: %v; args=%v", operation.Resource, operation.Action, err, arguments)
			}
		})
	}
}

func syntheticValue(valueType string, enum []string) string {
	for _, value := range enum {
		if value != "" {
			return value
		}
	}
	switch valueType {
	case "boolean":
		return "false"
	case "integer", "number":
		return "1"
	case "array":
		return "[]"
	case "object":
		return "{}"
	default:
		return fmt.Sprintf("fake_%s", valueType)
	}
}
