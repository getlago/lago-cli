package generated

import "strings"

// DestructiveVocabulary lists the path and action segments that mark an irreversible
// or money-moving Lago operation. The generator uses it to set Operation.Dangerous, and
// the CLI reuses it for requests that do not come from a generated command (`lago api`,
// fixture steps) whose path matches no known operation.
//
// Kept here, outside the generated JSON, so the generator, the runtime, and the parity
// test all read one list. internal/contract/parity_test.go carries an independent copy
// on purpose: it is the oracle that fails if this one drifts.
var DestructiveVocabulary = []string{
	"void", "finalize", "retry", "terminate", "delete", "destroy", "refund",
}

// MatchesDestructiveVocabulary reports whether value contains any destructive segment.
func MatchesDestructiveVocabulary(value string) bool {
	value = strings.ToLower(value)
	for _, segment := range DestructiveVocabulary {
		if strings.Contains(value, segment) {
			return true
		}
	}
	return false
}
