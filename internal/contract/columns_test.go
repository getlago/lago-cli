package contract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getlago/lago-cli/internal/output"
	"gopkg.in/yaml.v3"
)

// listSchemas names the response object schema behind each list wrapper the table
// renderer declares columns for. CustomerObject and InvoiceObject are allOf over a base
// object, so both halves are listed.
var listSchemas = map[string][]string{
	"customers":     {"CustomerObject", "CustomerBaseObject"},
	"invoices":      {"InvoiceObject", "InvoiceBaseObject"},
	"subscriptions": {"SubscriptionObject"},
	"plans":         {"PlanObject"},
}

// Every column the renderer declares for a resource must exist on that resource's
// response schema in the pinned spec. A spec rename otherwise prints an empty column
// in every list, and nothing else would notice.
func TestAllowlistedColumnsExistInSpec(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("..", "..", "spec", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]any `yaml:"properties"`
				AllOf      []struct {
					Properties map[string]any `yaml:"properties"`
				} `yaml:"allOf"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	for wrapper, columns := range output.ResourceColumns() {
		names, ok := listSchemas[wrapper]
		if !ok {
			t.Errorf("declared columns for %q but no response schema is mapped here", wrapper)
			continue
		}
		properties := map[string]bool{}
		for _, name := range names {
			schema, exists := document.Components.Schemas[name]
			if !exists {
				t.Errorf("schema %s for %q is not in the spec", name, wrapper)
				continue
			}
			for property := range schema.Properties {
				properties[property] = true
			}
			for _, part := range schema.AllOf {
				for property := range part.Properties {
					properties[property] = true
				}
			}
		}
		for _, column := range columns {
			if !properties[column] {
				t.Errorf("%s column %q is not a property of %v in the pinned spec", wrapper, column, names)
			}
		}
	}
}
