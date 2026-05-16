package onepassword

import (
	"context"
	"fmt"
	"strings"

	"github.com/1password/onepassword-sdk-go"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
)

var ErrSecretSource1PasswordInvalidLocation = fmt.Errorf("invalid 1Password secret location format")

// SecretSource1Password digs up secrets from 1Password.
// The URI scheme for this source is "op".
//
//	op://VAULT/ITEM/[SECTION]/FIELD
//
// In Spelunk, these "Secret Coordinates" are exactly the same as the "Secret Reference"
// that you can obtain by going to a 1Password vault item, selecting a field, and copying its "Secret Reference".
// See: https://developer.1password.com/docs/cli/secret-reference-syntax.
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSource1Password struct {
	client *onepassword.Client
}

// With1Password enables the SecretSource1Password.
func With1Password(client *onepassword.Client) spelunk.SpelunkerOption {
	return spelunk.WithSource(&SecretSource1Password{
		client: client,
	})
}

var _ types.SecretSource = (*SecretSource1Password)(nil)

func (s *SecretSource1Password) Type() string {
	return "op"
}

func (s *SecretSource1Password) DigUp(
	ctx context.Context,
	coord types.SecretCoord,
) (string, error) {
	parts := strings.Split(coord.Location, "/")
	if len(parts) < 3 || len(parts) > 4 {
		return "", fmt.Errorf(
			"%w: expected VAULT/ITEM/FIELD or VAULT/ITEM/SECTION/FIELD, got %q",
			ErrSecretSource1PasswordInvalidLocation,
			coord.Location,
		)
	}

	// 1Password expects the reference in the format op://vault/item/field
	opRef := "op://" + coord.Location

	secret, err := s.client.Secrets().Resolve(ctx, opRef)
	if err != nil {
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}

	return secret, nil
}
