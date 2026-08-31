package contract_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/cli"
	"github.com/getlago/lago-cli/internal/generated"
	"gopkg.in/yaml.v3"
)

func TestEveryOpenAPIOperationHasExactlyOneCommand(t *testing.T) {
	t.Parallel()
	root := cli.NewRoot(cli.NewApp(nil, io.Discard, io.Discard, "test"))
	seenCommands := map[string]string{}
	seenOperations := map[string]bool{}
	for _, operation := range generated.Operations {
		if seenOperations[operation.OperationID] {
			t.Fatalf("duplicate operation ID %s", operation.OperationID)
		}
		seenOperations[operation.OperationID] = true
		key := operation.Resource + " " + operation.Action
		if owner := seenCommands[key]; owner != "" {
			t.Fatalf("command %s maps both %s and %s", key, owner, operation.OperationID)
		}
		seenCommands[key] = operation.OperationID
		command, _, err := root.Find([]string{operation.Resource, operation.Action})
		if err != nil || command == nil || command.Name() != operation.Action {
			t.Fatalf("operation %s has no command %s: %v", operation.OperationID, key, err)
		}
		if command.Example == "" {
			t.Errorf("generated command %s has no example", key)
		}
	}

	specPath := filepath.Join("..", "..", "spec", "openapi.yaml")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, path := range spec.Paths {
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options"} {
			if _, exists := path[method]; exists {
				count++
			}
		}
	}
	if count != len(generated.Operations) {
		t.Fatalf("OpenAPI has %d operations but command manifest has %d", count, len(generated.Operations))
	}
	if count != 217 {
		t.Fatalf("review the pinned spec operation-count change: got %d, previous contract is 217", count)
	}
}

func TestGeneratedManifestEmbedsPinnedSpecIdentity(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("..", "..", "spec", "operations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Version string `json:"spec_version"`
		SHA     string `json:"spec_sha256"`
		Count   int    `json:"operation_count"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Version != generated.SpecVersion || manifest.SHA != generated.SpecSHA256 || manifest.Count != len(generated.Operations) {
		t.Fatalf("manifest identity is stale: %#v", manifest)
	}
}

// destructiveSegments are the trailing path segments that identify an irreversible
// or money-moving Lago operation. HTTP method semantics do not describe billing
// semantics: PUT /invoices/{id}/finalize is idempotent by RFC and irreversible by
// accounting. Every operation matching this vocabulary must be confirmation-gated
// and must never be retried without an explicit idempotency key.
var destructiveSegments = []string{
	"void", "finalize", "retry", "retry_payment", "terminate", "delete", "destroy",
}

func TestDestructiveOperationsAreGatedAndNeverAutoRetried(t *testing.T) {
	t.Parallel()
	for _, operation := range generated.Operations {
		if !isDestructiveOperation(operation) {
			continue
		}
		if !operation.Dangerous {
			t.Errorf("%s (%s %s) is destructive but is not confirmation-gated", operation.OperationID, operation.Method, operation.Path)
		}
		if operation.Idempotent {
			t.Errorf("%s (%s %s) is destructive but is marked auto-retryable", operation.OperationID, operation.Method, operation.Path)
		}
	}
}

func isDestructiveOperation(operation generated.Operation) bool {
	haystack := strings.ToLower(operation.Path + " " + operation.Action)
	for _, segment := range destructiveSegments {
		if strings.Contains(haystack, segment) {
			return true
		}
	}
	return operation.Method == http.MethodDelete
}
