package tomlpath

import (
	"context"
	"fmt"

	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/common"
	"github.com/detro/spelunk/v2/types"
	"github.com/ohler55/ojg/jp"
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

	compiledPath, err := jp.ParseString(mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrTOMLPathInvalid, mod, err)
	}

	matches := compiledPath.Get(data)
	if len(matches) == 0 {
		return "", fmt.Errorf("%w (%q): no match found", ErrTOMLPathFailed, mod)
	}

	// A single match is unwrapped, so that a matched list is post-processed as
	// the list it is, rather than as a list of matches.
	var res any = matches
	if len(matches) == 1 {
		res = matches[0]
	}

	strRes, err := common.PostProcessJSONPath(res)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTOMLPathMatchingFailed, err)
	}

	return strRes, nil
}

// WithTOMLPath adds the TOML JSONPath modifier to a Spelunker.
func WithTOMLPath() spelunk.SpelunkerOption {
	return spelunk.WithModifier(&SecretModifierTOMLPath{})
}
