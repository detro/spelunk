package vault_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/source/vault"
	"github.com/detro/spelunk/types"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainersvault "github.com/testcontainers/testcontainers-go/modules/vault"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestSecretSourceVault_Type(t *testing.T) {
	s := &vault.SecretSourceVault{}
	require.Equal(t, "vault", s.Type())
}

const (
	kvSecretEngineV1Mount = "kvSecretsV1"
	kvSecretEngineV2Mount = "kvSecretsV2"

	v1SecPath = kvSecretEngineV1Mount + "/my-app/secr3t"
	v2SecPath = kvSecretEngineV2Mount + "/data" + "/my/Other/App/s3cret"
)

var secData = map[string]any{
	"string_value": "one",
	"intValue":     2,
	"float-value":  0.23,
	"map-value": map[string]any{
		"k1": "v1",
		"k2": "v2",
		"k3": 3,
	},
	"array-value": []string{"forza", "napoli", "sempre"},
}

func TestSecretSourceVault_DigUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	vaultClient, err := setupVaultTestContainer(t)
	require.NoError(t, err)
	createTestSecrets(t, vaultClient)

	// Initialize Spelunker with Vault plugin
	spelunker := spelunk.NewSpelunker(vault.WithVault(vaultClient))

	tests := []struct {
		name     string
		coordStr string
		want     string
		wantJson map[string]any
		errMatch error
	}{
		{
			name:     "key from v1 secret",
			coordStr: fmt.Sprintf("vault://%s/%s", v1SecPath, "string_value"),
			want:     "one",
		},
		{
			name:     "key from v2 secret",
			coordStr: fmt.Sprintf("vault://%s/%s", v2SecPath, "string_value"),
			want:     "one",
		},
		{
			name:     "secret that does not exist",
			coordStr: "vault://secret/data/missing-secret/key",
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "key that does not exist from v1 secret",
			coordStr: fmt.Sprintf("vault://%s/%s", v1SecPath, "does-notExist"),
			errMatch: types.ErrSecretKeyNotFound,
		},
		{
			name:     "key that does not exist from v2 secret",
			coordStr: fmt.Sprintf("vault://%s/%s", v2SecPath, "does-notExist"),
			errMatch: types.ErrSecretKeyNotFound,
		},
		{
			name:     "invalid location (just key)",
			coordStr: "vault://key",
			errMatch: vault.ErrSecretSourceVaultInvalidLocation,
		},
		{
			name:     "invalid location (mount and key, but no secret)",
			coordStr: "vault://mount/key",
			errMatch: vault.ErrSecretSourceVaultInvalidLocation,
		},
		{
			name:     "invalid location (just mount, no secret)",
			coordStr: "vault://mount/",
			errMatch: vault.ErrSecretSourceVaultInvalidLocation,
		},
		{
			name:     "whole v1 secret",
			coordStr: fmt.Sprintf("vault://%s/", v1SecPath),
			wantJson: secData,
		},
		{
			name:     "whole v2 secret",
			coordStr: fmt.Sprintf("vault://%s/", v2SecPath),
			wantJson: secData,
		},
		{
			name:     "key from v1 secret via JSONPath modifier",
			coordStr: fmt.Sprintf("vault://%s/?jp=$.%s", v1SecPath, "string_value"),
			want:     "one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			got, err := spelunker.DigUp(t.Context(), coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			if tt.wantJson != nil {
				wantJsonStr, err := json.Marshal(tt.wantJson)
				require.NoError(t, err)
				require.JSONEq(t, string(wantJsonStr), got)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func createTestSecrets(t *testing.T, client *api.Client) {
	_, err := client.Logical().Write(v1SecPath, secData)
	require.NoError(t, err)
	_, err = client.Logical().Write(v2SecPath, map[string]interface{}{
		"data": secData,
	})
	require.NoError(t, err)
}

func setupVaultTestContainer(t *testing.T) (*api.Client, error) {
	// Launch Vault container with 2 secrets engine: KV v1 and KV v2
	rootToken := rand.String(10)
	vaultContainer, err := testcontainersvault.Run(t.Context(),
		// See: https://hub.docker.com/r/hashicorp/vault.
		"hashicorp/vault:1.21",
		testcontainersvault.WithToken(rootToken),
		testcontainersvault.WithInitCommand(
			fmt.Sprintf("secrets enable -path %s -version=1 kv", kvSecretEngineV1Mount),
			fmt.Sprintf("secrets enable -path %s -version=2 kv", kvSecretEngineV2Mount),
		),
	)
	testcontainers.CleanupContainer(t, vaultContainer)
	require.NoError(t, err)

	// Work out mapped URL
	hostIP, err := vaultContainer.Host(t.Context())
	require.NoError(t, err)
	mappedPort, err := vaultContainer.MappedPort(t.Context(), "8200/tcp")
	require.NoError(t, err)
	mappedURL := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	// Setup client with root token
	config := api.DefaultConfig()
	config.Address = mappedURL
	config.Timeout = 5 * time.Second
	client, err := api.NewClient(config)
	require.NoError(t, err)
	client.SetToken(rootToken)

	return client, nil
}
