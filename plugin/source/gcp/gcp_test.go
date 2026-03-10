package gcp_test

import (
	"context"
	"fmt"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/source/gcp"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSecretSourceGCP_Type(t *testing.T) {
	s := &gcp.SecretSourceGCP{}
	require.Equal(t, "gcp", s.Type())
}

const (
	projectID   = "test-project"
	secretName  = "my-secret"
	secretValue = "super-secret-value"

	jsonSecretName  = "my-json-secret"
	jsonSecretValue = `{"password":"super-secret-value"}`
)

func TestSecretSourceGCP_DigUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test
	ctx := context.Background()
	client, err := setupGCPTestContainer(t, ctx)
	require.NoError(t, err)
	createTestSecrets(t, err, client, ctx)

	// Initialize Spelunker with GCP plugin
	spelunker := spelunk.NewSpelunker(gcp.WithGCP(client))

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "valid secret latest",
			coordStr: fmt.Sprintf("gcp://projects/%s/secrets/%s", projectID, secretName),
			want:     secretValue,
		},
		{
			name:     "valid secret specific version",
			coordStr: fmt.Sprintf("gcp://projects/%s/secrets/%s/versions/1", projectID, secretName),
			want:     secretValue,
		},
		{
			name: "valid secret specific latest version",
			coordStr: fmt.Sprintf(
				"gcp://projects/%s/secrets/%s/versions/latest",
				projectID,
				secretName,
			),
			want: secretValue,
		},
		{
			name: "valid secret via jp modifier",
			coordStr: fmt.Sprintf(
				"gcp://projects/%s/secrets/%s/?jp=$.password",
				projectID,
				jsonSecretName,
			),
			want: "super-secret-value",
		},
		{
			name:     "secret not found",
			coordStr: fmt.Sprintf("gcp://projects/%s/secrets/missing-secret", projectID),
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "invalid location missing project",
			coordStr: "gcp://secrets/secret-name",
			errMatch: gcp.ErrSecretSourceGCPInvalidLocation,
		},
		{
			name:     "invalid project name",
			coordStr: "gcp://projects/pr/secrets/secret-name",
			errMatch: gcp.ErrSecretSourceGCPInvalidLocation,
		},
		{
			name:     "invalid latest version",
			coordStr: fmt.Sprintf("gcp://projects/%s/secrets/missing-secret/versions/latesttypo", projectID),
			errMatch: gcp.ErrSecretSourceGCPInvalidLocation,
		},
		{
			name:     "invalid version",
			coordStr: fmt.Sprintf("gcp://projects/%s/secrets/missing-secret/versions/123b", projectID),
			errMatch: gcp.ErrSecretSourceGCPInvalidLocation,
		},
		{
			name:     "missing version",
			coordStr: fmt.Sprintf("gcp://projects/%s/secrets/missing-secret/versions/", projectID),
			errMatch: gcp.ErrSecretSourceGCPInvalidLocation,
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

func createTestSecrets(t *testing.T, err error, client *secretmanager.Client, ctx context.Context) {
	// Create secret in GCP Secret Manager Emulator
	parent := fmt.Sprintf("projects/%s", projectID)
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: secretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}
	secret, err := client.CreateSecret(ctx, createSecretReq)
	require.NoError(t, err)

	addVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}
	_, err = client.AddSecretVersion(ctx, addVersionReq)
	require.NoError(t, err)

	createJsonSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: jsonSecretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}
	jsonSecret, err := client.CreateSecret(ctx, createJsonSecretReq)
	require.NoError(t, err)
	_, err = client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: jsonSecret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(jsonSecretValue),
		},
	})
	require.NoError(t, err)
}

func setupGCPTestContainer(t *testing.T, ctx context.Context) (*secretmanager.Client, error) {
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/blackwell-systems/gcp-secret-manager-emulator:1.3.0",
		ExposedPorts: []string{"9090/tcp"},
		WaitingFor:   wait.ForListeningPort("9090/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "9090")
	require.NoError(t, err)

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client, err := secretmanager.NewClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	return client, nil
}
