package keeper_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/plugin/modifier/jsonpath/v2"
	spelunkkeeper "github.com/detro/spelunk/plugin/source/keeper/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	ksm "github.com/keeper-security/secrets-manager-go/core"
	"github.com/stretchr/testify/require"
)

func TestSecretSourceKeeper_Type(t *testing.T) {
	s := &spelunkkeeper.SecretSourceKeeper{}
	require.Equal(t, "kp", s.Type())
}

func TestSecretSourceKeeper_DigUp_Parsing(t *testing.T) {
	// Initialize a dummy client with in-memory storage so no client-config.json
	// file is created on disk during tests.
	dummyClient := ksm.NewSecretsManager(&ksm.ClientOptions{
		Token:  "US:dummy-token",
		Config: ksm.NewMemoryKeyValueStorage(),
	})

	spelunker := spelunk.NewSpelunker(
		spelunkkeeper.WithKeeper(dummyClient),
		jsonpath.WithJSONPath(),
	)

	tests := []struct {
		name     string
		coordStr string
		errMatch error
	}{
		{
			name:     "invalid location (just a slash)",
			coordStr: "kp:///",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (missing record UID but has field)",
			coordStr: "kp:///password",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (UID too short - 21 chars)",
			coordStr: "kp://123456789012345678901/password",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (UID too long - 23 chars)",
			coordStr: "kp://12345678901234567890123/password",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (UID contains invalid chars)",
			coordStr: "kp://invalid_uid_with_!@#$/password",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "valid 22-char UID coordinate without field (attempts fetch)",
			coordStr: "kp://abcdefghijklmnopqrstuv",
			errMatch: types.ErrCouldNotFetchSecret,
		},
		{
			name:     "valid 22-char UID coordinate with trailing slash (attempts fetch)",
			coordStr: "kp://abcdefghijklmnopqrstuv/",
			errMatch: types.ErrCouldNotFetchSecret,
		},
		{
			name:     "valid 22-char UID coordinate with standard field (attempts fetch)",
			coordStr: "kp://abcdefghijklmnopqrstuv/password",
			errMatch: types.ErrCouldNotFetchSecret,
		},
		{
			name:     "valid 22-char UID coordinate with custom field (attempts fetch)",
			coordStr: "kp://abcdefghijklmnopqrstuv/custom_field",
			errMatch: types.ErrCouldNotFetchSecret,
		},
		{
			name:     "valid 22-char UID containing base64url characters _ and - (attempts fetch)",
			coordStr: "kp://_bcdefghijklmnopqrstu-/password",
			errMatch: types.ErrCouldNotFetchSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			_, err = spelunker.DigUp(context.Background(), coord)
			require.ErrorIs(t, err, tt.errMatch)
		})
	}
}
