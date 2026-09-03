package output

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// Pair is one labelled value in a hand-ordered key/value table.
type Pair struct {
	Key   string
	Value string
}

// WritePairs prints pairs in the order given, as the table renderer prints a resource:
// upper-case key, two-space gap, sanitised value. Blank values are skipped. Built-in
// commands whose output is a short identity block (`whoami`) use it so the order is the
// one a reader expects rather than alphabetical.
func WritePairs(out io.Writer, pairs []Pair) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for _, pair := range pairs {
		if pair.Value == "" {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\n", header(pair.Key), Sanitize(pair.Value))
	}
	return w.Flush()
}
