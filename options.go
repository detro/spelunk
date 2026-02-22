package spelunk

import (
	b64mod "github.com/detro/spelunk/builtin/modifier/base64"
	"github.com/detro/spelunk/builtin/modifier/jsonpath"
	b64src "github.com/detro/spelunk/builtin/source/base64"
	"github.com/detro/spelunk/builtin/source/env"
	"github.com/detro/spelunk/builtin/source/file"
	"github.com/detro/spelunk/builtin/source/plain"
	"github.com/detro/spelunk/types"
)

// options are the internal configuration used by an instance of Spelunker.
// They are set by client code using implementations of SpelunkerOption.
type options struct {
	trimValue bool
	sources   map[string]types.SecretSource
	modifiers map[string]types.SecretModifier
}

func (o *options) apply(opts ...SpelunkerOption) *options {
	if o.sources == nil {
		o.sources = make(map[string]types.SecretSource)
	}
	if o.modifiers == nil {
		o.modifiers = make(map[string]types.SecretModifier)
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func defaultOptions() []SpelunkerOption {
	return []SpelunkerOption{
		WithTrimValue(),
		WithSource(&plain.SecretSourcePlain{}),
		WithSource(&file.SecretSourceFile{}),
		WithSource(&env.SecretSourceEnv{}),
		WithSource(&b64src.SecretSourceBase64{}),
		WithModifier(&jsonpath.SecretModifierJSONPath{}),
		WithModifier(&b64mod.SecretModifierBase64{}),
	}
}

// SpelunkerOption options that can be provided to NewSpelunker.
type SpelunkerOption func(*options)

// WithTrimValue all leading and trailing (Unicode) white spaces
// of the secret value are removed.
// Enabled by Default.
func WithTrimValue() SpelunkerOption {
	return func(o *options) {
		o.trimValue = true
	}
}

// WithoutTrimValue all leading and trailing (Unicode) white spaces
// of the secret value are left alone.
func WithoutTrimValue() SpelunkerOption {
	return func(o *options) {
		o.trimValue = false
	}
}

// WithSource adds the given types.SecretSource to the set of sources
// a Spelunker can use to dig-up secrets.
func WithSource(source types.SecretSource) SpelunkerOption {
	return func(o *options) {
		o.sources[source.Type()] = source
	}
}

// WithModifier adds the given types.SecretModifier to the set of modifiers
// a Spelunker can apply to the value of secrets.
func WithModifier(modifier types.SecretModifier) SpelunkerOption {
	return func(o *options) {
		o.modifiers[modifier.Type()] = modifier
	}
}
