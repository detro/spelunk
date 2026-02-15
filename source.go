package spelunk

import (
	"context"
)

// SecretSource is a source of secrets.
type SecretSource interface {
	// Type returns the unique identifier for the type of SecretSource.
	Type() string

	// DigUp returns the secret pointed at by the given SecretCoord.
	DigUp(context.Context, SecretCoord) (string, error)
}
