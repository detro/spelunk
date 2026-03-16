package jsonpathutil

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// PostProcessJSONPathResult takes the result of a jsonpath extraction
// and formats it into a string following common rules for spelunk modifiers
// (jp, yp, tp). It extracts the first element of lists, converts floats
// appropriately, returns errors for null, and JSON-marshals other complex types.
func PostProcessJSONPathResult(res any) (string, error) {
	// If the JSONPath matches multiple values, focus on the first one.
	// NOTE: user can always refine their JSONPath modifier if they need something else.
	if list, ok := res.([]any); ok {
		if len(list) > 0 {
			res = list[0]
		}
	}

	switch v := res.(type) {
	case string:
		return v, nil
	case float64:
		// Convert float to string, removing trailing zeros
		// (e.g. 1.500000 -> 1.5, 10.000000 -> 10).
		//
		// We use 'f' format to avoid scientific notation for large numbers,
		// and -1 precision to use the smallest number of digits necessary.
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return fmt.Sprintf("%v", v), nil
	case nil:
		return "", fmt.Errorf("result is null")
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(b), nil
	}
}
