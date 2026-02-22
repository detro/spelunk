package base64_test

import (
	"context"
	b64 "encoding/base64"
	"testing"

	"github.com/detro/spelunk"
	b64mod "github.com/detro/spelunk/builtin/modifier/base64"
	"github.com/detro/spelunk/builtin/modifier/jsonpath"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

// mockSource implements spelunk.SecretSource for testing
type mockSource struct {
	typ string
	val string
	err error
}

func (m *mockSource) Type() string {
	return m.typ
}

func (m *mockSource) DigUp(_ context.Context, _ types.SecretCoord) (string, error) {
	return m.val, m.err
}

func TestSecretModifierBase64_Type(t *testing.T) {
	s := &b64mod.SecretModifierBase64{}
	require.Equal(t, "b64", s.Type())
}

func TestSecretModifierBase64_Modify(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		val      string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "encode plain string",
			val:      "hello world",
			coordStr: "test://loc?b64",
			want:     b64.StdEncoding.EncodeToString([]byte("hello world")),
		},
		{
			name:     "encode with modArg ignored",
			val:      "ignore me",
			coordStr: "test://loc?b64=encode",
			want:     b64.StdEncoding.EncodeToString([]byte("ignore me")),
		},
		{
			name:     "encode empty string",
			val:      "",
			coordStr: "test://loc?b64",
			want:     "",
		},
		{
			name:     "apply jsonpath then base64",
			val:      `{"key": "mysecret"}`,
			coordStr: "test://loc?jp=$.key&b64",
			want:     b64.StdEncoding.EncodeToString([]byte("mysecret")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			spelunker := spelunk.NewSpelunker(
				spelunk.WithSource(&mockSource{typ: "test", val: tt.val}),
				spelunk.WithModifier(&jsonpath.SecretModifierJSONPath{}),
				spelunk.WithModifier(&b64mod.SecretModifierBase64{}),
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
