package types

import "context"

// SecretModifier receives a secret and applies a modification to it.
type SecretModifier interface {
	// Type returns the unique identifier for the type of SecretModifier.
	Type() string

	// Modify applies a modification to the given secret value.
	Modify(ctx context.Context, secretValue string, mod string) (string, error)
}
