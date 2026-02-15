package base64_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/detro/spelunk"
	b64 "github.com/detro/spelunk/builtin/source/base64"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSourceBase64_Type(t *testing.T) {
	s := &b64.SecretSourceBase64{}
	require.Equal(t, "base64", s.Type())
}

func TestSecretSourceBase64_DigUp(t *testing.T) {
	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "simple value",
			coordStr: "base64://" + base64.StdEncoding.EncodeToString([]byte("my-secret")),
			want:     "my-secret",
		},
		{
			name:     "value with special chars",
			coordStr: "base64://" + base64.StdEncoding.EncodeToString([]byte("user:pass@host")),
			want:     "user:pass@host",
		},
		{
			name:     "invalid base64",
			coordStr: "base64://invalid-base64-string",
			errMatch: b64.ErrSecretSourceBase64FailedDecoding,
		},
	}

	spelunker := spelunk.NewSpelunker()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			got, err := spelunker.DigUp(context.Background(), coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
