package bitwarden

import (
	"context"
	"fmt"
	"strings"

	"github.com/bitwarden/sdk-go/v2"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	"github.com/google/uuid"
)

var ErrSecretSourceBitwardenInvalidLocation = fmt.Errorf("invalid Bitwarden secret location format")

// SecretSourceBitwarden digs up secrets from Bitwarden Secrets Manager.
// The URI scheme for this source is "bw".
//
//	bw://SECRET_ID
//
// Where `SECRET_ID` is a UUIDv4, as documented in https://bitwarden.com/help/secrets-manager-cli/.
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceBitwarden struct {
	client sdk.BitwardenClientInterface
}

// WithBitwarden enables the SecretSourceBitwarden.
func WithBitwarden(client sdk.BitwardenClientInterface) spelunk.SpelunkerOption {
	source := &SecretSourceBitwarden{
		client: client,
	}
	return spelunk.WithSource(source)
}

var _ types.SecretSource = (*SecretSourceBitwarden)(nil)

func (s *SecretSourceBitwarden) Type() string {
	return "bw"
}

func (s *SecretSourceBitwarden) DigUp(
	_ context.Context,
	coord types.SecretCoord,
) (string, error) {
	secretID := strings.Trim(coord.Location, "/")
	id, err := uuid.Parse(secretID)
	if err != nil || id.Version() != 4 {
		return "", fmt.Errorf(
			"%w: expected SECRET_ID to be a valid UUIDv4, got %q",
			ErrSecretSourceBitwardenInvalidLocation,
			coord.Location,
		)
	}

	secret, err := s.client.Secrets().Get(secretID)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}

	return secret.Value, nil
}
