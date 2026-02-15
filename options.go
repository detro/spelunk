package spelunk

import (
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
}

func (o *options) apply(opts ...SpelunkerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultOptions() []SpelunkerOption {
	return []SpelunkerOption{
		WithTrimValue(),
		WithSource(&plain.SecretSourcePlain{}),
		WithSource(&file.SecretSourceFile{}),
		WithSource(&env.SecretSourceEnv{}),
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

// WithSource adds the given SecretSource to the set of sources
// a Spelunker can use to dig-up secrets.
func WithSource(source types.SecretSource) SpelunkerOption {
	return func(o *options) {
		o.sources[source.Type()] = source
	}
}
