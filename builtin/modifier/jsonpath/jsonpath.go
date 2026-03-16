package jsonpath

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/detro/spelunk/internal/jsonpathutil"
	"github.com/detro/spelunk/types"
	jp "github.com/oliveagle/jsonpath"
)

var (
	ErrSecretNotJSON          = fmt.Errorf("secret is not a valid JSON")
	ErrJSONPathInvalid        = fmt.Errorf("invalid JSONPath expression")
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
	var data any
	if err := json.Unmarshal([]byte(secretValue), &data); err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretNotJSON, err)
	}

	compiledPath, err := jp.Compile(mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrJSONPathInvalid, mod, err)
	}

	res, err := compiledPath.Lookup(data)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrJSONPathFailed, mod, err)
	}

	strRes, err := jsonpathutil.PostProcessJSONPathResult(res)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrJSONPathMatchingFailed, err)
	}

	return strRes, nil
}
