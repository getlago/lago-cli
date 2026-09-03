package output

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Sanitize makes API-controlled text safe to print to a terminal.
//
// A customer name is whatever the caller sent Lago, and a table cell prints it raw. QA
// created a customer whose name held an ANSI colour sequence and a newline: the escape
// recoloured the terminal and the newline injected a fake row. Every C0 and C1 control
// character (ESC, CR, LF, TAB, NUL, DEL, 0x80-0x9f) is therefore replaced by its visible
// escape, `\x1b`, `\n`, `\r`, `\t`, `\xNN`, and an invalid UTF-8 byte by `\xNN`. Replacing
// rather than stripping keeps the evidence: an operator sees that the name contains an
// escape instead of a silently shortened string. Printable text, including non-ASCII, is
// untouched. --output json is unaffected; encoding/json escapes on its own.
func Sanitize(text string) string {
	if utf8.ValidString(text) && strings.IndexFunc(text, isControl) < 0 {
		return text
	}
	var builder strings.Builder
	builder.Grow(len(text) + 8)
	for index := 0; index < len(text); {
		r, size := utf8.DecodeRuneInString(text[index:])
		if r == utf8.RuneError && size == 1 {
			fmt.Fprintf(&builder, `\x%02x`, text[index])
			index++
			continue
		}
		switch {
		case r == 0x1b:
			builder.WriteString(`\x1b`)
		case r == '\n':
			builder.WriteString(`\n`)
		case r == '\r':
			builder.WriteString(`\r`)
		case r == '\t':
			builder.WriteString(`\t`)
		case isControl(r):
			fmt.Fprintf(&builder, `\x%02x`, r)
		default:
			builder.WriteRune(r)
		}
		index += size
	}
	return builder.String()
}

func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}

// cell is the one path every table value takes to the terminal: summarised, then
// sanitised. header does the same for column names and keys, which are also API data.
func cell(value any) string { return Sanitize(summary(value)) }

func header(key string) string { return Sanitize(strings.ToUpper(key)) }
