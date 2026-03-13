package azure_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/source/azure"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSecretSourceAzure_Type(t *testing.T) {
	s := &azure.SecretSourceAzure{}
	require.Equal(t, "az", s.Type())
}

const (
	plainSecretName  = "my-secret"
	plainSecretValue = "top-secret-value"

	jsonSecretName  = "my-json-secret"
	jsonSecretValue = `{"key":"value"}`
)

func TestSecretSourceAzure_DigUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	azClient, err := setupAzureTestContainer(t, ctx)
	require.NoError(t, err)
	createTestSecrets(t, azClient)

	spelunker := spelunk.NewSpelunker(azure.WithAzure(azClient))

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "secret by name",
			coordStr: fmt.Sprintf("az://%s", plainSecretName),
			want:     plainSecretValue,
		},
		{
			name:     "secret by name via jp modifier",
			coordStr: fmt.Sprintf("az://%s/?jp=$.key", jsonSecretName),
			want:     "value",
		},
		{
			name:     "secret that does not exist",
			coordStr: "az://missing-secret",
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "invalid secret name format",
			coordStr: "az://invalid_secret_name", // underscores not allowed
			errMatch: azure.ErrSecretSourceAzureInvalidLocation,
		},
		{
			name: "invalid secret version format",
			coordStr: fmt.Sprintf(
				"az://%s/invalid-version-format",
				plainSecretName,
			), // versions are 32 hex chars
			errMatch: azure.ErrSecretSourceAzureInvalidLocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			got, err := spelunker.DigUp(ctx, coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}

func createTestSecrets(t *testing.T, client *azsecrets.Client) {
	_, err := client.SetSecret(
		t.Context(),
		plainSecretName,
		azsecrets.SetSecretParameters{Value: new(plainSecretValue)},
		nil,
	)
	require.NoError(t, err)

	_, err = client.SetSecret(
		t.Context(),
		jsonSecretName,
		azsecrets.SetSecretParameters{Value: new(jsonSecretValue)},
		nil,
	)
	require.NoError(t, err)
}

func setupAzureTestContainer(t *testing.T, ctx context.Context) (*azsecrets.Client, error) {
	azContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			// See: https://hub.docker.com/r/nagyesta/lowkey-vault
			Image:        "nagyesta/lowkey-vault:7.1.32",
			ExposedPorts: []string{"8443/tcp"},
			Env: map[string]string{
				"LOWKEY_ARGS": "--LOWKEY_VAULT_RELAXED_PORTS=true",
			},
			WaitingFor: wait.ForLog("Started LowkeyVaultApp").
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, azContainer)

	mappedPort, err := azContainer.MappedPort(ctx, "8443/tcp")
	require.NoError(t, err)
	hostIP, err := azContainer.Host(ctx)
	require.NoError(t, err)

	lowkeyVaultURL := fmt.Sprintf("https://%s:%s", hostIP, mappedPort.Port())

	// Need a custom HTTP client that skips TLS verification
	client, err := azsecrets.NewClient(lowkeyVaultURL, &fakeCredential{}, &azsecrets.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Transport: &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}},
		},
		DisableChallengeResourceVerification: true,
	})
	require.NoError(t, err)
	return client, nil
}

// fakeCredential implements azcore.TokenCredential for emulator testing
type fakeCredential struct{}

func (f *fakeCredential) GetToken(
	_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}
