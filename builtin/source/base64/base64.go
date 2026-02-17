package base64

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/detro/spelunk/types"
)

var ErrSecretSourceBase64FailedDecoding = fmt.Errorf("failed to decode base64 secret")

// SecretSourceBase64 digs up secrets that are base64-encoded in the URI-coordinates.
// The URI scheme for this source is "base64".
//
//	base64://BASE64_ENCODED_SECRET
//
// This types.SecretSource is built-in to spelunker.Spelunker.
type SecretSourceBase64 struct{}

var _ types.SecretSource = (*SecretSourceBase64)(nil)

func (s *SecretSourceBase64) Type() string {
	return "base64"
}

func (s *SecretSourceBase64) DigUp(_ context.Context, coord types.SecretCoord) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(coord.Location)
	if err != nil {
		return "", fmt.Errorf(
			"%w (%q): %w",
			ErrSecretSourceBase64FailedDecoding,
			coord.Location,
			err,
		)
	}
	return string(decoded), nil
}
