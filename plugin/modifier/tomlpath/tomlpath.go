package tomlpath

import (
	"context"
	"fmt"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/internal/jsonpathutil"
	"github.com/detro/spelunk/types"
	jp "github.com/oliveagle/jsonpath"
	"github.com/pelletier/go-toml/v2"
)

var (
	ErrTOMLPathInvalid        = fmt.Errorf("invalid TOML JSONPath expression")
	ErrTOMLPathFailed         = fmt.Errorf("failed to apply TOML JSONPath")
	ErrTOMLPathMatchingFailed = fmt.Errorf("failed to match TOML JSONPath")
	ErrSecretNotTOML          = fmt.Errorf("secret is not a valid TOML")
)

// SecretModifierTOMLPath is a modifier that can extract a specific field out of a TOML stored in a secret value.
// It parses the TOML into an object and applies standard JSONPath to it.
type SecretModifierTOMLPath struct{}

var _ types.SecretModifier = (*SecretModifierTOMLPath)(nil)

func (s *SecretModifierTOMLPath) Type() string {
	return "tp"
}

func (s *SecretModifierTOMLPath) Modify(
	_ context.Context,
	secretValue string,
	mod string,
) (string, error) {
	var data any
	if err := toml.Unmarshal([]byte(secretValue), &data); err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretNotTOML, err)
	}

	compiledPath, err := jp.Compile(mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrTOMLPathInvalid, mod, err)
	}

	res, err := compiledPath.Lookup(data)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrTOMLPathFailed, mod, err)
	}

	strRes, err := jsonpathutil.PostProcessJSONPathResult(res)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTOMLPathMatchingFailed, err)
	}

	return strRes, nil
}

// WithTOMLPath adds the TOML JSONPath modifier to a Spelunker.
func WithTOMLPath() spelunk.SpelunkerOption {
	return spelunk.WithModifier(&SecretModifierTOMLPath{})
}
