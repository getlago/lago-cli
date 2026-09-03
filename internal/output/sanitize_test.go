package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestSanitizeEscapesEveryControlClass(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ in, want string }{
		{"\x1b[31mred\x1b[0m", `\x1b[31mred\x1b[0m`},
		{"line\nbreak", `line\nbreak`},
		{"carriage\rreturn", `carriage\rreturn`},
		{"tab\tstop", `tab\tstop`},
		{"nul\x00byte", `nul\x00byte`},
		{"del\x7f", `del\x7f`},
		{"c1\u0085next", `c1\x85next`},
		{"bad\xffbyte", `bad\xffbyte`},
		{"café 日本 ✓", "café 日本 ✓"},
		{"", ""},
	} {
		if got := Sanitize(tc.in); got != tc.want {
			t.Errorf("Sanitize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// QA S-22: a customer name holding an ANSI escape and a newline recoloured the terminal
// and injected a fake row. No raw control byte may reach a table.
func TestQA_S22_TableEscapeSanitization(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		value    any
		wantRows int
		want     []string
	}{
		{
			name:     "ANSI escape in a list cell",
			value:    map[string]any{"customers": []any{map[string]any{"lago_id": "cus_1", "name": "\x1b[31mred\x1b[0m"}}},
			wantRows: 2,
			want:     []string{`\x1b[31mred\x1b[0m`},
		},
		{
			name:     "newline cannot inject a fake row",
			value:    map[string]any{"customers": []any{map[string]any{"lago_id": "cus_1", "name": "Alice\nLAGO_ID  evil\n"}}},
			wantRows: 2,
			want:     []string{`Alice\nLAGO_ID  evil\n`},
		},
		{
			name:     "carriage return in a key/value table and an escape in a key",
			value:    map[string]any{"customer": map[string]any{"name": "over\rwrite", "\x1bkey": "v"}},
			wantRows: 2,
			want:     []string{`over\rwrite`, `\x1bKEY`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := render(t, Table, "", tc.value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.ContainsAny(out, "\x1b\r") {
				t.Errorf("raw control byte reached the table:\n%q", out)
			}
			if rows := strings.Count(out, "\n"); rows != tc.wantRows {
				t.Errorf("rows = %d, want %d:\n%q", rows, tc.wantRows, out)
			}
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Errorf("output missing visible escape %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestQA_S22_IdentifierBlockIsSanitized(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	value := map[string]any{"customer": map[string]any{"lago_id": "cus_1", "name": "Alice\x1b[2J\nNAME  fake"}}
	if err := (Renderer{Mode: Table, Out: &buffer, Identifiers: true}).Render(value); err != nil {
		t.Fatal(err)
	}
	out := buffer.String()
	if strings.Contains(out, "\x1b") || strings.Count(out, "\n") != 2 {
		t.Errorf("identifier block leaked a control byte or a fake row:\n%q", out)
	}
	if !strings.Contains(out, `Alice\x1b[2J\nNAME  fake`) {
		t.Errorf("visible escape missing:\n%s", out)
	}
}

// The empty-match hint names response keys, which are API data too.
func TestEmptyMatchHintIsSanitized(t *testing.T) {
	t.Parallel()
	var out, errOut bytes.Buffer
	if err := (Renderer{Mode: JSON, Query: "missing", Out: &out, Err: &errOut}).Render(map[string]any{"a\x1bb": 1}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errOut.String(), "\x1b") || !strings.Contains(errOut.String(), `a\x1bb`) {
		t.Errorf("hint leaked a control byte: %q", errOut.String())
	}
}
