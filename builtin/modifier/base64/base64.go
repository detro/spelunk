package base64

import (
	"context"
	b64 "encoding/base64"

	"github.com/detro/spelunk/types"
)

// SecretModifierBase64 is a modifier that encodes the secret value to a base64 string.
//
// To use it, append the modifier `b64` to the given secret coordinates string:
//
//	plain://my-secret?b64
type SecretModifierBase64 struct{}

var _ types.SecretModifier = (*SecretModifierBase64)(nil)

func (s *SecretModifierBase64) Type() string {
	return "b64"
}

func (s *SecretModifierBase64) Modify(
	_ context.Context,
	secretValue string,
	_ string,
) (string, error) {
	return b64.StdEncoding.EncodeToString([]byte(secretValue)), nil
}
