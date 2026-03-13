package base64_encoder

import (
	"github.com/detro/spelunk/builtin/modifier/base64"
	"github.com/detro/spelunk/types"
)

// SecretModifierBase64Encoder is a modifier that encodes the secret value to a base64 string.
//
// To use it, append the modifier `b64e` to the given secret coordinates string:
//
//	plain://my-secret?b64e
//
// NOTE: This is an "alias" for the `?b64` encoder, implemented by base64.SecretModifierBase64.
// The modifier type `?b64e` is provided with symmetry with the decoder `?b64d`.
type SecretModifierBase64Encoder struct {
	base64.SecretModifierBase64
}

var _ types.SecretModifier = (*SecretModifierBase64Encoder)(nil)

func (s *SecretModifierBase64Encoder) Type() string {
	return "b64e"
}
