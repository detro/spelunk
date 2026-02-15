package env

import (
	"context"
	"fmt"
	"os"

	"github.com/detro/spelunk/types"
)

var (
	ErrSecretSourceEnvDoesNotExist = fmt.Errorf("environment variable does not exist")
)

// SecretSourceEnv digs up secrets from environment variables.
// The URI scheme for this source is "env".
//
//	env://MY_API_KEY
//
// This types.SecretSource is built-in to spelunker.Spelunker.
type SecretSourceEnv struct{}

var _ types.SecretSource = (*SecretSourceEnv)(nil)

func (s *SecretSourceEnv) Type() string {
	return "env"
}

func (s *SecretSourceEnv) DigUp(_ context.Context, coord types.SecretCoord) (string, error) {
	val, exists := os.LookupEnv(coord.Location)
	if !exists {
		return "", fmt.Errorf("%w: %q", ErrSecretSourceEnvDoesNotExist, coord.Location)
	}

	return val, nil
}
