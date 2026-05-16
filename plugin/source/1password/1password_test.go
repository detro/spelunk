package onepassword_test

import (
	"context"
	"os"
	"testing"

	"github.com/1password/onepassword-sdk-go"
	"github.com/detro/spelunk"
	spelunkop "github.com/detro/spelunk/plugin/source/1password"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSource1Password_Type(t *testing.T) {
	s := &spelunkop.SecretSource1Password{}
	require.Equal(t, "op", s.Type())
}

// TestSecretSource1Password_DigUp_Integration because there is no way to simulate locally a 1Password installation,
// tests here are designed to be executed by maintainers with access to a specific 1Password vault and with the
// environment variable `OP_SERVICE_ACCOUNT_TOKEN` set.
// These tests will not be executed in CI, unfortunately.
func TestSecretSource1Password_DigUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	saToken, found := os.LookupEnv("OP_SERVICE_ACCOUNT_TOKEN")
	if !found {
		t.Skip("Skipping 1Password integration test: OP_SERVICE_ACCOUNT_TOKEN not set")
	}

	client, err := onepassword.NewClient(
		context.Background(),
		onepassword.WithServiceAccountToken(saToken),
		onepassword.WithIntegrationInfo("Spelunk Integration Tests", "dev"),
	)
	require.NoError(t, err)

	spelunker := spelunk.NewSpelunker(spelunkop.With1Password(client))

	tests := []struct {
		name     string
		coordStr string
		expected string
		errMatch error
	}{
		{
			name:     "invalid location (not enough parts)",
			coordStr: "op://my-vault/my-item",
			errMatch: spelunkop.ErrSecretSource1PasswordInvalidLocation,
		},
		{
			name:     "invalid location (too many parts)",
			coordStr: "op://my-vault/my-item/my-section/my-field/extra",
			errMatch: spelunkop.ErrSecretSource1PasswordInvalidLocation,
		},
		{
			name:     "secret that does not exist",
			coordStr: "op://non-existent-vault/item/password",
			errMatch: types.ErrCouldNotFetchSecret,
		},
		{
			name:     "secret that does not exist (with section)",
			coordStr: "op://non-existent-vault/item/section/password",
			errMatch: types.ErrCouldNotFetchSecret,
		},
		{
			name:     "valid secret (standard field)",
			coordStr: "op://spelunk-integration-tests/Integrations Tests Account/password",
			expected: "spelunker-test-password",
		},
		{
			name:     "valid secret (field inside section)",
			coordStr: "op://spelunk-integration-tests/Integrations Tests Account/test-section/test-email",
			expected: "spelunker-integration-test@test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			val, err := spelunker.DigUp(t.Context(), coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, val)
			}
		})
	}
}
