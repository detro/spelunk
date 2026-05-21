package keeper_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk"
	spelunkkeeper "github.com/detro/spelunk/plugin/source/keeper"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSourceKeeper_Type(t *testing.T) {
	s := &spelunkkeeper.SecretSourceKeeper{}
	require.Equal(t, "kp", s.Type())
}

func TestSecretSourceKeeper_DigUp_Parsing(t *testing.T) {
	// We can pass nil for the client because (for now) we are only testing
	// the coordinate parsing logic which happens before client invocation.
	spelunker := spelunk.NewSpelunker(spelunkkeeper.WithKeeper(nil))

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
			name:     "invalid location (UID too short)",
			coordStr: "kp://short_uid/password",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (UID contains invalid chars)",
			coordStr: "kp://invalid_uid_with_!@#$/password",
			errMatch: types.ErrInvalidLocation,
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
