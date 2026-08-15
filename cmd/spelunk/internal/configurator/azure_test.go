package configurator_test

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
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	azurePlainSecretName  = "my-secret"
	azurePlainSecretValue = "top-secret-value"

	azureJsonSecretName  = "my-json-secret"
	azureJsonSecretValue = `{"key":"value"}`
)

type fakeAzureCredential struct{}

func (f *fakeAzureCredential) GetToken(
	_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

func setupAzureTestContainer(t *testing.T, ctx context.Context) (*azsecrets.Client, string, error) {
	t.Helper()
	azContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
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
	if err != nil {
		return nil, "", err
	}
	testcontainers.CleanupContainer(t, azContainer)

	mappedPort, err := azContainer.MappedPort(ctx, "8443/tcp")
	if err != nil {
		return nil, "", err
	}
	hostIP, err := azContainer.Host(ctx)
	if err != nil {
		return nil, "", err
	}

	lowkeyVaultURL := fmt.Sprintf("https://%s:%s", hostIP, mappedPort.Port())

	client, err := azsecrets.NewClient(
		lowkeyVaultURL,
		&fakeAzureCredential{},
		&azsecrets.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				APIVersion: "7.4",
				Transport: &http.Client{Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}},
			},
			DisableChallengeResourceVerification: true,
		},
	)
	if err != nil {
		return nil, "", err
	}

	return client, lowkeyVaultURL, nil
}

func createAzureTestSecrets(t *testing.T, client *azsecrets.Client) {
	t.Helper()
	_, err := client.SetSecret(
		t.Context(),
		azurePlainSecretName,
		azsecrets.SetSecretParameters{Value: new(azurePlainSecretValue)},
		nil,
	)
	require.NoError(t, err)

	_, err = client.SetSecret(
		t.Context(),
		azureJsonSecretName,
		azsecrets.SetSecretParameters{Value: new(azureJsonSecretValue)},
		nil,
	)
	require.NoError(t, err)
}

func TestPlugin_Azure_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Azure E2E integration test in short mode")
	}

	bin := buildCLI(t)
	ctx := context.Background()
	client, vaultURL, err := setupAzureTestContainer(t, ctx)
	require.NoError(t, err)
	createAzureTestSecrets(t, client)

	env := cleanEnv(t,
		fmt.Sprintf("AZURE_KEYVAULT_URL=%s", vaultURL),
		"AZURE_TESTING_LOWKEY_VAULT=true",
	)

	t.Run("dig secret by name", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"dig",
			fmt.Sprintf("az://%s", azurePlainSecretName),
			"--azure-insecure-skip-tls-verify",
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, azurePlainSecretValue, res.Stdout)
	})

	t.Run("dig with jsonpath modifier", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"dig",
			fmt.Sprintf("az://%s/?jp=$.key", azureJsonSecretName),
			"--azure-insecure-skip-tls-verify",
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, "value", res.Stdout)
	})

	t.Run("exist secret found", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exists",
			fmt.Sprintf("az://%s", azurePlainSecretName),
			"--azure-insecure-skip-tls-verify",
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})

	t.Run("exist secret missing", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exists",
			"az://missing-secret",
			"--azure-insecure-skip-tls-verify",
		)
		require.NotEqual(t, 0, res.ExitCode)
	})

	t.Run("creds check", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "creds", "--azure-insecure-skip-tls-verify")
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})
}
