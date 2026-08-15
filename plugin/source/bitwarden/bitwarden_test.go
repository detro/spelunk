package bitwarden_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bitwarden/sdk-go/v2"
	"github.com/detro/spelunk/plugin/modifier/jsonpath/v2"
	spelunkbw "github.com/detro/spelunk/plugin/source/bitwarden/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockSecrets struct {
	sdk.SecretsInterface
	secrets map[string]*sdk.SecretResponse
	err     error
}

func (m *mockSecrets) Get(id string) (*sdk.SecretResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if s, ok := m.secrets[id]; ok {
		return s, nil
	}
	return nil, errors.New("secret not found in mock")
}

type mockBitwardenClient struct {
	sdk.BitwardenClientInterface
	secrets *mockSecrets
}

func (m *mockBitwardenClient) Secrets() sdk.SecretsInterface {
	return m.secrets
}

func TestSecretSourceBitwarden_Type(t *testing.T) {
	s := &spelunkbw.SecretSourceBitwarden{}
	require.Equal(t, "bw", s.Type())
}

func TestSecretSourceBitwarden_DigUp(t *testing.T) {
	validUUID := uuid.NewString()
	mockClient := &mockBitwardenClient{
		secrets: &mockSecrets{
			secrets: map[string]*sdk.SecretResponse{
				validUUID: {
					Value: "my-bitwarden-secret",
				},
			},
		},
	}

	spelunker := spelunk.NewSpelunker(
		spelunkbw.WithBitwarden(mockClient),
		jsonpath.WithJSONPath(),
	)

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "valid uuidv4 coordinate",
			coordStr: fmt.Sprintf("bw://%s", validUUID),
			want:     "my-bitwarden-secret",
		},
		{
			name:     "valid uuidv4 coordinate with trailing slash",
			coordStr: fmt.Sprintf("bw://%s/", validUUID),
			want:     "my-bitwarden-secret",
		},
		{
			name:     "valid uuidv4 coordinate with leading slash",
			coordStr: fmt.Sprintf("bw:///%s", validUUID),
			want:     "my-bitwarden-secret",
		},
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
		{
			name:     "secret retrieval error",
			coordStr: fmt.Sprintf("bw://%s", uuid.NewString()),
			errMatch: types.ErrCouldNotFetchSecret,
		},
	}

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
