package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRendererJSONWithQuery(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := (Renderer{Mode: JSON, Query: "customers[0].name", Out: &output}).Render(map[string]any{"customers": []any{map[string]any{"name": "Example"}}})
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != "\"Example\"\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRendererTableAndErrors(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	value := map[string]any{"customers": []any{map[string]any{"external_id": "cust_1", "name": "Example"}}}
	if err := (Renderer{Mode: Table, Out: &output}).Render(value); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"EXTERNAL_ID", "cust_1", "Example"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("table %q lacks %q", output.String(), expected)
		}
	}
	if err := (Renderer{Mode: "xml", Out: &output}).Render(value); err == nil {
		t.Fatal("unsupported output unexpectedly succeeded")
	}
	if err := (Renderer{Mode: JSON, Query: "[", Out: &output}).Render(value); err == nil {
		t.Fatal("invalid query unexpectedly succeeded")
	}
}
