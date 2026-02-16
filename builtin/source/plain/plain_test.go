package plain_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/builtin/source/plain"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSourcePlain_Type(t *testing.T) {
	s := &plain.SecretSourcePlain{}
	if got := s.Type(); got != "plain" {
		t.Errorf("SecretSourcePlain.Type() = %v, want %v", got, "plain")
	}
}

func TestSecretSourcePlain_DigUp(t *testing.T) {
	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "simple value",
			coordStr: "plain://my-secret",
			want:     "my-secret",
		},
		{
			name:     "value with path",
			coordStr: "plain://my/nested/secret",
			want:     "my/nested/secret",
		},
		{
			name:     "value with special chars",
			coordStr: "plain://user:pass@host",
			want:     "user:pass@host",
		},
	}

	spelunker := spelunk.NewSpelunker(
		spelunk.WithSource(&plain.SecretSourcePlain{}),
	)

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
