package bitwarden_test

import (
	"context"
	"testing"

	"github.com/detro/spelunk/plugin/modifier/jsonpath/v2"
	spelunkbw "github.com/detro/spelunk/plugin/source/bitwarden/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/stretchr/testify/require"
)

func TestSecretSourceBitwarden_Type(t *testing.T) {
	s := &spelunkbw.SecretSourceBitwarden{}
	require.Equal(t, "bw", s.Type())
}

func TestSecretSourceBitwarden_DigUp_Parsing(t *testing.T) {
	// We can pass nil for the client because (for now) we are only testing
	// the coordinate parsing logic which happens before client invocation.
	spelunker := spelunk.NewSpelunker(
		spelunkbw.WithBitwarden(nil),
		jsonpath.WithJSONPath(),
	)

	tests := []struct {
		name     string
		coordStr string
		errMatch error
	}{
		{
			name:     "invalid location (just a slash)",
			coordStr: "bw:///",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (not a uuid)",
			coordStr: "bw://just-a-string",
			errMatch: types.ErrInvalidLocation,
		},
		{
			name:     "invalid location (not uuidv4 - this is a v1)",
			coordStr: "bw://5120353c-1b70-11ee-be56-0242ac120002",
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
