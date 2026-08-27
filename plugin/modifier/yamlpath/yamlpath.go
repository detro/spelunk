package yamlpath

import (
	"context"
	"fmt"

	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/common"
	"github.com/detro/spelunk/v2/types"
	"github.com/ohler55/ojg/jp"
	"gopkg.in/yaml.v3"
)

var (
	ErrYAMLPathInvalid        = fmt.Errorf("invalid YAML JSONPath expression")
	ErrYAMLPathFailed         = fmt.Errorf("failed to apply YAML JSONPath")
	ErrYAMLPathMatchingFailed = fmt.Errorf("failed to match YAML JSONPath")
	ErrSecretNotYAML          = fmt.Errorf("secret is not a valid YAML")
)

// SecretModifierYAMLPath is a modifier that can extract a specific field out of a YAML stored in a secret value.
// It parses the YAML into an object and applies standard JSONPath to it.
type SecretModifierYAMLPath struct{}

var _ types.SecretModifier = (*SecretModifierYAMLPath)(nil)

func (s *SecretModifierYAMLPath) Type() string {
	return "yp"
}

func (s *SecretModifierYAMLPath) Modify(
	_ context.Context,
	secretValue string,
	mod string,
) (string, error) {
	var data any
	if err := yaml.Unmarshal([]byte(secretValue), &data); err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretNotYAML, err)
	}

	compiledPath, err := jp.ParseString(mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrYAMLPathInvalid, mod, err)
	}

	matches := compiledPath.Get(data)
	if len(matches) == 0 {
		return "", fmt.Errorf("%w (%q): no match found", ErrYAMLPathFailed, mod)
	}

	// A single match is unwrapped, so that a matched list is post-processed as
	// the list it is, rather than as a list of matches.
	var res any = matches
	if len(matches) == 1 {
		res = matches[0]
	}

	strRes, err := common.PostProcessJSONPath(res)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrYAMLPathMatchingFailed, err)
	}

	return strRes, nil
}

// WithYAMLPath adds the YAML JSONPath modifier to a Spelunker.
func WithYAMLPath() spelunk.SpelunkerOption {
	return spelunk.WithModifier(&SecretModifierYAMLPath{})
}
