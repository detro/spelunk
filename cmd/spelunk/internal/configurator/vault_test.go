package configurator_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainersvault "github.com/testcontainers/testcontainers-go/modules/vault"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	vaultKvV1Mount = "kvSecretsV1"
	vaultKvV2Mount = "kvSecretsV2"

	vaultV1SecPath = vaultKvV1Mount + "/my-app/secr3t"
	vaultV2SecPath = vaultKvV2Mount + "/data/my/Other/App/s3cret"
)

var vaultSecData = map[string]any{
	"string_value": "one",
	"intValue":     2,
}

func setupVaultTestContainer(t *testing.T) (*api.Client, string, string, error) {
	t.Helper()
	rootToken := rand.String(10)
	vaultContainer, err := testcontainersvault.Run(t.Context(),
		"hashicorp/vault:1.21",
		testcontainersvault.WithToken(rootToken),
		testcontainersvault.WithInitCommand(
			fmt.Sprintf("secrets enable -path %s -version=1 kv", vaultKvV1Mount),
			fmt.Sprintf("secrets enable -path %s -version=2 kv", vaultKvV2Mount),
		),
	)
	if err != nil {
		return nil, "", "", err
	}
	testcontainers.CleanupContainer(t, vaultContainer)

	hostIP, err := vaultContainer.Host(t.Context())
	if err != nil {
		return nil, "", "", err
	}
	mappedPort, err := vaultContainer.MappedPort(t.Context(), "8200/tcp")
	if err != nil {
		return nil, "", "", err
	}
	mappedURL := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	cfg := api.DefaultConfig()
	cfg.Address = mappedURL
	cfg.Timeout = 5 * time.Second
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, "", "", err
	}
	client.SetToken(rootToken)

	return client, mappedURL, rootToken, nil
}

func createVaultTestSecrets(t *testing.T, client *api.Client) {
	t.Helper()
	_, err := client.Logical().Write(vaultV1SecPath, vaultSecData)
	require.NoError(t, err)
	_, err = client.Logical().Write(vaultV2SecPath, map[string]any{
		"data": vaultSecData,
	})
	require.NoError(t, err)
}

func TestPlugin_Vault_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Vault E2E integration test in short mode")
	}

	bin := buildCLI(t)
	client, vaultAddr, vaultToken, err := setupVaultTestContainer(t)
	require.NoError(t, err)
	createVaultTestSecrets(t, client)

	env := cleanEnv(t,
		fmt.Sprintf("VAULT_ADDR=%s", vaultAddr),
		fmt.Sprintf("VAULT_TOKEN=%s", vaultToken),
	)

	ctx := context.Background()

	t.Run("dig v1 secret key", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "dig", fmt.Sprintf("vault://%s/string_value", vaultV1SecPath))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, "one", res.Stdout)
	})

	t.Run("dig v2 secret key", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "dig", fmt.Sprintf("vault://%s/string_value", vaultV2SecPath))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, "one", res.Stdout)
	})

	t.Run("default dig whole v1 secret as json with jsonpath", func(t *testing.T) {
		res := runCLI(ctx, bin, env, fmt.Sprintf("vault://%s/?jp=$.intValue", vaultV1SecPath))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, "2", res.Stdout)
	})

	t.Run("exist secret found", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exist",
			fmt.Sprintf("vault://%s/string_value", vaultV1SecPath),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})

	t.Run("exist secret missing", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "exist", "vault://missing/secret/key")
		require.NotEqual(t, 0, res.ExitCode)
	})

	t.Run("creds check", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "creds")
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})
}
