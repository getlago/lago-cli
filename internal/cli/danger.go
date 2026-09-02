package cli

import (
	"net/http"
	"strings"

	"github.com/getlago/lago-cli/internal/generated"
	"github.com/getlago/lago-cli/internal/transport"
)

// classifyRequest decides whether a raw request (`lago api`, a fixture step) must go
// through the same confirmation gate as a generated command.
//
// The request is matched against the generated operation table by method and path
// template, so `POST /invoices/{id}/void` is gated because the spec classified it, not
// because a verb list happened to include it. A path no operation claims falls back to
// the same default-deny rule the generator applies: DELETE is always dangerous, and so
// is any path carrying a destructive segment. The matched operation ID is returned for
// messages and tests; it is empty when nothing matched.
func classifyRequest(method, path string) (dangerous bool, operationID string) {
	method = strings.ToUpper(strings.TrimSpace(method))
	actual := requestPathSegments(path)
	for _, operation := range generated.Operations {
		if operation.Method != method {
			continue
		}
		if matchOperationPath(operation.Path, actual) {
			return operation.Dangerous, operation.OperationID
		}
	}
	return method == http.MethodDelete || generated.MatchesDestructiveVocabulary(strings.Join(actual, "/")), ""
}

// requestPathSegments normalizes a request path the way the transport will send it
// (redundant /api/v1 stripped, query dropped) and splits it into segments.
func requestPathSegments(path string) []string {
	if index := strings.IndexAny(path, "?#"); index >= 0 {
		path = path[:index]
	}
	normalized := transport.NormalizeRequestPath(strings.TrimSpace(path))
	return splitPath(normalized)
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

// matchOperationPath reports whether a concrete path matches an OpenAPI path template.
// A `{param}` segment matches any single non-empty segment; every other segment must
// match exactly.
func matchOperationPath(template string, actual []string) bool {
	expected := splitPath(template)
	if len(expected) != len(actual) {
		return false
	}
	for index, segment := range expected {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			if actual[index] == "" {
				return false
			}
			continue
		}
		if segment != actual[index] {
			return false
		}
	}
	return true
}
