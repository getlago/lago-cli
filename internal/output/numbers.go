package output

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/getlago/lago-cli/internal/apperr"
	"gopkg.in/yaml.v3"
)

// API responses are decoded with json.Decoder.UseNumber so that monetary values keep
// their exact decimal literal all the way to the terminal. That exactness is correct
// and it makes json.Number, not float64, the numeric type flowing through the renderer.
// Two consumers cannot read json.Number, and each needs its own conversion:
//
//   - go-jmespath v0.4.0 type-checks numeric arguments against a concrete float64,
//     so `invoices[?amount_cents > ` + "`1000`" + `]` silently matched nothing.
//   - gopkg.in/yaml.v3 marshals json.Number as the string type it is, so `--output yaml`
//     emitted amount_cents: "150000" where --output json emitted 150000.

// queryValue returns a copy of value with every json.Number converted to float64 so
// JMESPath numeric filters and aggregations work.
//
// A number that cannot round-trip through float64 is refused rather than silently
// rounded: returning a slightly wrong invoice total is worse than declining to filter.
func queryValue(value any) (any, error) {
	switch typed := value.(type) {
	case json.Number:
		return exactFloat(typed)
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			converted, err := queryValue(nested)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			result[key] = converted
		}
		return result, nil
	case []any:
		result := make([]any, len(typed))
		for index, nested := range typed {
			converted, err := queryValue(nested)
			if err != nil {
				return nil, err
			}
			result[index] = converted
		}
		return result, nil
	default:
		return value, nil
	}
}

// exactFloat converts a JSON number for JMESPath, refusing conversions that would
// silently corrupt money.
//
// Per DECISIONS.md, monetary values are always integer minor units, so the guard is
// on integers: any integer beyond float64's exact range (2^53) would be filtered
// against a value that is provably not the one the API returned. Non-integer
// literals are rates, percentages and unit amounts, where the ordinary float64
// approximation is what every JSON consumer already uses; refusing those would make
// --query unusable against responses that merely contain a decimal.
func exactFloat(number json.Number) (float64, error) {
	approximate, err := number.Float64()
	if err != nil {
		return 0, fmt.Errorf("value %s is not a number", number)
	}
	exact, ok := new(big.Rat).SetString(number.String())
	if !ok {
		return 0, fmt.Errorf("value %s is not a number", number)
	}
	if !exact.IsInt() {
		return approximate, nil
	}
	if rounded := new(big.Rat).SetFloat64(approximate); rounded == nil || rounded.Cmp(exact) != 0 {
		return 0, fmt.Errorf("value %s cannot be filtered without losing precision", number)
	}
	return approximate, nil
}

// yamlValue returns a copy of value with every json.Number replaced by a YAML scalar
// node carrying the original literal, so YAML and JSON agree on both type and digits.
func yamlValue(value any) any {
	switch typed := value.(type) {
	case json.Number:
		tag := "!!int"
		if strings.ContainsAny(typed.String(), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: typed.String()}
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[key] = yamlValue(nested)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, nested := range typed {
			result[index] = yamlValue(nested)
		}
		return result
	default:
		return value
	}
}

func queryError(err error) error {
	return apperr.New(
		apperr.ExitUsage,
		fmt.Sprintf("cannot evaluate --query against this response: %v", err),
		"Re-run with --output json and no --query, then filter the exact values with jq.",
	)
}
