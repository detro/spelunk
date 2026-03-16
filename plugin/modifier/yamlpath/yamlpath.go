package yamlpath

import (
	"context"
	"fmt"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/internal/jsonpathutil"
	"github.com/detro/spelunk/types"
	jp "github.com/oliveagle/jsonpath"
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

	compiledPath, err := jp.Compile(mod)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrYAMLPathInvalid, mod, err)
	}

	res, err := compiledPath.Lookup(data)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", ErrYAMLPathFailed, mod, err)
	}

	strRes, err := jsonpathutil.PostProcessJSONPathResult(res)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrYAMLPathMatchingFailed, err)
	}

	return strRes, nil
}

// WithYAMLPath adds the YAML JSONPath modifier to a Spelunker.
func WithYAMLPath() spelunk.SpelunkerOption {
	return spelunk.WithModifier(&SecretModifierYAMLPath{})
}
