package jsonpath

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/detro/spelunk/types"
	jp "github.com/oliveagle/jsonpath"
)

var (
	ErrSecretNotJSON          = fmt.Errorf("secret is not a valid JSON")
	ErrJSONPathFailed         = fmt.Errorf("failed to apply JSONPath")
	ErrJSONPathMatchingFailed = fmt.Errorf("failed to match JSONPath")
)

// SecretModifierJSONPath is a modifier that can extract a specific field out of a JSON stored in a secret value.
// After the secret has been dug-up, the modifier digs further at the provided JSONPath, and returns
// the value found there.
//
// To use it, append the modifier `jq` to the given secret coordinates string:
//
//	k8s://NAMESPACE/NAME/KEY?jp=$.kafka.brokers
//
// JSONPath (https://goessner.net/articles/JsonPath/) defines a string syntax for selecting and extracting
// JSON (RFC-8259) values from within a given JSON object.
//
// If a given JSONPath refers to multiple elements, only the first one is returned.
//
// JSONPath has been normalized as RFC-9535 (https://www.rfc-editor.org/rfc/rfc9535).
//
// See: https://github.com/oliveagle/jsonpath (underlying library).
type SecretModifierJSONPath struct{}

var _ types.SecretModifier = (*SecretModifierJSONPath)(nil)

func (s *SecretModifierJSONPath) Type() string {
	return "jp"
}

func (s *SecretModifierJSONPath) Modify(
	_ context.Context,
	secretValue string,
	mod string,
) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(secretValue), &data); err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretNotJSON, err)
	}

	res, err := jp.JsonPathLookup(data, mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrJSONPathFailed, mod, err)
	}

	// If the JSONPath matches multiple values, focus on the first one.
	// NOTE: user can always refine their JSONPath modifier if they need something else.
	if list, ok := res.([]interface{}); ok {
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
		return "", fmt.Errorf("%w: result is null", ErrJSONPathMatchingFailed)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf(
				"%w: failed to marshal result: %w",
				ErrJSONPathMatchingFailed,
				err,
			)
		}
		return string(b), nil
	}
}
