package spelunk

import (
	"context"
)

// SecretSourcePlain digs up secrets that can be found with URI-coordinates in the format:
//
//	plain://MY_SECRET_VALUE
//
// This SecretSource is built-in to Spelunker.
//
// Note that `MY_SECRET` can also be in the form `MY/SECRET_VALUE` or `MY/SECRET/VALUE`:
// the whole combination of URI _authority_ and _path_ is returned.
type SecretSourcePlain struct{}

func (s *SecretSourcePlain) Type() string {
	return "plain"
}

func (s *SecretSourcePlain) DigUp(_ context.Context, coord SecretCoord) (string, error) {
	return coord.Location, nil
}
