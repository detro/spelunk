package base64_decoder

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/detro/spelunk/types"
)

var ErrSecretModifierBase64DecoderFailedDecoding = fmt.Errorf("failed to decode base64 secret")

// SecretModifierBase64Decoder is a modifier that decodes a base64 string to a secret value.
//
// To use it, append the modifier `b64d` to the given secret coordinates string:
//
//	plain://bXktc2VjcmV0?b64d
type SecretModifierBase64Decoder struct{}

var _ types.SecretModifier = (*SecretModifierBase64Decoder)(nil)

func (s *SecretModifierBase64Decoder) Type() string {
	return "b64d"
}

func (s *SecretModifierBase64Decoder) Modify(
	_ context.Context,
	secretValue string,
	_ string,
) (string, error) {
	decoded, err := b64.StdEncoding.DecodeString(secretValue)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretModifierBase64DecoderFailedDecoding, err)
	}
	return string(decoded), nil
}
