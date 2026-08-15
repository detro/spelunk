package plain_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/builtin/source/plain"
	"github.com/detro/spelunk/v2/types"
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
		{
			name:     "value with leading slash",
			coordStr: "plain:///path/to/secret",
			want:     "/path/to/secret",
		},
		{
			name:     "value with root slash",
			coordStr: "plain:///",
			want:     "/",
		},
		{
			name:     "value with percent-encoded characters",
			coordStr: "plain:///hello%20world%21",
			want:     "/hello world!",
		},
		{
			name:     "value with whitespace and multiline",
			coordStr: "plain:///line1%20word%0Aline2",
			want:     "/line1 word\nline2",
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
