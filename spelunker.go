package spelunk

import (
	"context"
	"fmt"
	"strings"

	"github.com/detro/spelunk/types"
)

var (
	ErrUnsupportedSecretSourceType   = fmt.Errorf("unsupported secret source type")
	ErrUnsupportedSecretModifierType = fmt.Errorf("unsupported secret modifier type")
	ErrFailedToDigUpSecret           = fmt.Errorf("failed to dig-up secret")
	ErrFailedToApplyModifier         = fmt.Errorf("failed to apply modifier")
)

// Spelunker digs up secrets from given SecretCoord.
type Spelunker struct {
	opts options
}

// NewSpelunker creates a new Spelunker.
// It can be configured providing one or more SpelunkerOption.
func NewSpelunker(opts ...SpelunkerOption) *Spelunker {
	s := &Spelunker{}

	s.opts.
		apply(defaultOptions()...).
		apply(opts...)

	return s
}

// DigUp digs up a secret using the given *SecretCoord.
func (s *Spelunker) DigUp(ctx context.Context, coord *types.SecretCoord) (string, error) {
	// Identify the source of the secret
	source, found := s.opts.sources[coord.Type]
	if !found {
		return "", fmt.Errorf("%w: %q", ErrUnsupportedSecretSourceType, coord.Type)
	}

	// Dig-up the secret from the source
	val, err := source.DigUp(ctx, *coord)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToDigUpSecret, err)
	}

	// Apply modifiers, if any
	for _, mod := range coord.Modifiers {
		modifier, found := s.opts.modifiers[mod[0]]
		if !found {
			return "", fmt.Errorf("%w: %q", ErrUnsupportedSecretModifierType, mod[0])
		}

		val, err = modifier.Modify(ctx, val, mod[1])
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrFailedToApplyModifier, err)
		}
	}

	// Lastly, trim value if configured
	if s.opts.trimValue {
		val = strings.TrimSpace(val)
	}

	return val, nil
}
