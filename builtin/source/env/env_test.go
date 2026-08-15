package env_test

import (
	"context"
	"os"
	"testing"

	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/builtin/source/env"
	"github.com/detro/spelunk/v2/types"
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
	require.NoError(t, os.Setenv("my_secret_var", "lowercase-value"))
	require.NoError(t, os.Setenv("USER@VAR_NAME", "userinfo-value"))
	t.Cleanup(func() {
		_ = os.Unsetenv("TEST_SECRET_KEY")
		_ = os.Unsetenv("TEST_EMPTY_KEY")
		_ = os.Unsetenv("TEST_SECRET_KEY_WITH_WHITESPACES")
		_ = os.Unsetenv("my_secret_var")
		_ = os.Unsetenv("USER@VAR_NAME")
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
			name:     "lowercase and underscore variable",
			coordStr: "env://my_secret_var",
			want:     "lowercase-value",
		},
		{
			name:     "variable with userinfo format",
			coordStr: "env://USER@VAR_NAME",
			want:     "userinfo-value",
		},
		{
			name:     "missing variable with leading slash",
			coordStr: "env:///TEST_SECRET_KEY",
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "missing variable",
			coordStr: "env://NON_EXISTENT_VAR",
			errMatch: types.ErrSecretNotFound,
		},
	}

	spelunker := spelunk.NewSpelunker(
		spelunk.WithSource(&env.SecretSourceEnv{}),
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
