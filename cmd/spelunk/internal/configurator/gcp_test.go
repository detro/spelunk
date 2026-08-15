package configurator_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	gcpProjectID       = "test-project"
	gcpPlainSecretName = "my-secret"
	gcpPlainSecretVal  = "super-secret-value"

	gcpJsonSecretName = "my-json-secret"
	gcpJsonSecretVal  = `{"password":"super-secret-value"}`
)

func setupGCPTestContainer(t *testing.T) (*secretmanager.Client, string, error) {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/blackwell-systems/gcp-secret-manager-emulator:1.3.0",
		ExposedPorts: []string{"9090/tcp"},
		WaitingFor:   wait.ForListeningPort("9090/tcp"),
	}
	container, err := testcontainers.GenericContainer(
		t.Context(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		return nil, "", err
	}
	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(t.Context())
	if err != nil {
		return nil, "", err
	}
	port, err := container.MappedPort(t.Context(), "9090")
	if err != nil {
		return nil, "", err
	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, "", err
	}

	client, err := secretmanager.NewClient(t.Context(), option.WithGRPCConn(conn))
	if err != nil {
		return nil, "", err
	}

	return client, addr, nil
}

func createGCPTestSecrets(t *testing.T, client *secretmanager.Client) {
	t.Helper()
	parent := fmt.Sprintf("projects/%s", gcpProjectID)
	secret, err := client.CreateSecret(t.Context(), &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: gcpPlainSecretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	require.NoError(t, err)

	_, err = client.AddSecretVersion(t.Context(), &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(gcpPlainSecretVal),
		},
	})
	require.NoError(t, err)

	jsonSecret, err := client.CreateSecret(t.Context(), &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: gcpJsonSecretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	require.NoError(t, err)

	_, err = client.AddSecretVersion(t.Context(), &secretmanagerpb.AddSecretVersionRequest{
		Parent: jsonSecret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(gcpJsonSecretVal),
		},
	})
	require.NoError(t, err)
}

func TestPlugin_GCP_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GCP E2E integration test in short mode")
	}

	bin := buildCLI(t)
	client, addr, err := setupGCPTestContainer(t)
	require.NoError(t, err)
	createGCPTestSecrets(t, client)

	env := cleanEnv(t,
		fmt.Sprintf("SECRET_MANAGER_EMULATOR_HOST=%s", addr),
	)

	ctx := context.Background()

	t.Run("dig base64 decoded secret", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"dig",
			fmt.Sprintf("gcp://projects/%s/secrets/%s?b64d", gcpProjectID, gcpPlainSecretName),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, gcpPlainSecretVal, res.Stdout)
	})

	t.Run("dig with jsonpath modifier", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"dig",
			fmt.Sprintf(
				"gcp://projects/%s/secrets/%s/?b64d&jp=$.password",
				gcpProjectID,
				gcpJsonSecretName,
			),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, "super-secret-value", res.Stdout)
	})

	t.Run("dig raw secret is base64 encoded", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			fmt.Sprintf("gcp://projects/%s/secrets/%s", gcpProjectID, gcpPlainSecretName),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, base64.StdEncoding.EncodeToString([]byte(gcpPlainSecretVal)), res.Stdout)
	})

	t.Run("exist secret found", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exists",
			fmt.Sprintf("gcp://projects/%s/secrets/%s", gcpProjectID, gcpPlainSecretName),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})

	t.Run("exist secret missing", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exists",
			fmt.Sprintf("gcp://projects/%s/secrets/missing-secret", gcpProjectID),
		)
		require.NotEqual(t, 0, res.ExitCode)
	})

	t.Run("creds check", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "creds")
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})
}
