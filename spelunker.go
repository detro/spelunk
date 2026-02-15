package spelunk

import (
	"context"
	"fmt"
	"strings"

	"github.com/detro/spelunk/types"
)

var (
	ErrUnsupportedSecretSourceType = fmt.Errorf("unsupported secret source type")
	ErrFailedToDigUpSecret         = fmt.Errorf("failed to dig-up secret")
)

// Spelunker digs up secrets from given SecretCoord.
type Spelunker struct {
	opts options
}

// NewSpelunker creates a new Spelunker.
// It can be configured providing one or more SpelunkerOption.
func NewSpelunker(opts ...SpelunkerOption) *Spelunker {
	s := &Spelunker{
		opts: options{
			sources: make(map[string]types.SecretSource),
		},
	}

	s.opts.apply(defaultOptions()...)
	s.opts.apply(opts...)

	return s
}

// DigUp digs up a secret using the given *SecretCoord.
func (s *Spelunker) DigUp(ctx context.Context, coord *types.SecretCoord) (string, error) {
	source, found := s.opts.sources[coord.Type]
	if !found {
		return "", fmt.Errorf("%w: %q", ErrUnsupportedSecretSourceType, coord.Type)
	}

	val, err := source.DigUp(ctx, *coord)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToDigUpSecret, err)
	}

	// TODO implement support for modifiers here

	if s.opts.trimValue {
		val = strings.TrimSpace(val)
	}

	return val, nil
}
