package spelunk

import (
	"context"
)

// SecretSource is a source of secrets.
// It has a Type(), matching a SecretCoord.Type, and
// it can be asked to DigUp() a secret using the given SecretCoord.
type SecretSource interface {
	Type() string
	DigUp(context.Context, SecretCoord) (string, error)
}
