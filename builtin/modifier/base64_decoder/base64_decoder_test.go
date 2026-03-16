package base64_decoder_test

import (
	"context"
	b64 "encoding/base64"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/builtin/modifier/base64_decoder"
	"github.com/detro/spelunk/builtin/modifier/jsonpath"
	"github.com/detro/spelunk/internal/testutil"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretModifierBase64Decoder_Type(t *testing.T) {
	s := &base64_decoder.SecretModifierBase64Decoder{}
	require.Equal(t, "b64d", s.Type())
}

func TestSecretModifierBase64Decoder_Modify(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		val      string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "decode plain string",
			val:      b64.StdEncoding.EncodeToString([]byte("hello world")),
			coordStr: "test://loc?b64d",
			want:     "hello world",
		},
		{
			name:     "decode with modArg ignored",
			val:      b64.StdEncoding.EncodeToString([]byte("ignore me")),
			coordStr: "test://loc?b64d=decode",
			want:     "ignore me",
		},
		{
			name:     "decode empty string",
			val:      "",
			coordStr: "test://loc?b64d",
			want:     "",
		},
		{
			name:     "apply jsonpath then base64 decode",
			val:      `{"key": "bXlzZWNyZXQ="}`, // bXlzZWNyZXQ= is "mysecret" in base64
			coordStr: "test://loc?jp=$.key&b64d",
			want:     "mysecret",
		},
		{
			name:     "invalid base64",
			val:      "invalid base64 string",
			coordStr: "test://loc?b64d",
			errMatch: base64_decoder.ErrSecretModifierBase64DecoderFailedDecoding,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(&testutil.MockSource{Typ: "test", Val: tt.val}),
				spelunk.WithModifier(&jsonpath.SecretModifierJSONPath{}),
				spelunk.WithModifier(&base64_decoder.SecretModifierBase64Decoder{}),
			)

			got, err := spelunker.DigUp(ctx, coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
