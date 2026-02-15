package env_test

import (
	"context"
	"os"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/builtin/source/env"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSourceEnv_Type(t *testing.T) {
	s := &env.SecretSourceEnv{}
	require.Equal(t, "env", s.Type())
}

func TestSecretSourceEnv_DigUp(t *testing.T) {
	// Setup test environment variables
	require.NoError(t, os.Setenv("TEST_SECRET_KEY", "super-secret-value"))
	require.NoError(t, os.Setenv("TEST_EMPTY_KEY", ""))
	require.NoError(t, os.Setenv("TEST_SECRET_KEY_WITH_WHITESPACES", "\nsecret\tword\r"))
	t.Cleanup(func() {
		os.Unsetenv("TEST_SECRET_KEY")
		os.Unsetenv("TEST_EMPTY_KEY")
		os.Unsetenv("TEST_SECRET_KEY_WITH_WHITESPACES")
	})

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "existing variable",
			coordStr: "env://TEST_SECRET_KEY",
			want:     "super-secret-value",
		},
		{
			name:     "empty variable",
			coordStr: "env://TEST_EMPTY_KEY",
			want:     "",
		},
		{
			name:     "variable with whitespace",
			coordStr: "env://TEST_SECRET_KEY_WITH_WHITESPACES",
			want:     "secret\tword",
		},
		{
			name:     "missing variable",
			coordStr: "env://NON_EXISTENT_VAR",
			errMatch: env.ErrSecretSourceEnvDoesNotExist,
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
