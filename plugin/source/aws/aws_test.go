package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/detro/spelunk"
	spelunkaws "github.com/detro/spelunk/plugin/source/aws"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

func TestSecretSourceAWS_Type(t *testing.T) {
	s := &spelunkaws.SecretSourceAWS{}
	require.Equal(t, "aws", s.Type())
}

const (
	secretName  = "my-app/secret"
	secretValue = `{"key":"value"}`
)

func TestSecretSourceAWS_DigUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	awsClient, err := setupAWSTestContainer(t)
	require.NoError(t, err)
	secret := createTestSecrets(t, awsClient)

	// Initialize Spelunker with AWS plugin
	spelunker := spelunk.NewSpelunker(spelunkaws.WithAWS(awsClient))

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "secret by name",
			coordStr: fmt.Sprintf("aws://%s", secretName),
			want:     secretValue,
		},
		{
			name:     "secret by exact ARN (with ///)",
			coordStr: fmt.Sprintf("aws:///%s", *secret.ARN),
			want:     secretValue,
		},
		{
			name:     "secret that does not exist",
			coordStr: "aws://missing/secret",
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "invalid location (empty)",
			coordStr: "aws:///",
			errMatch: spelunkaws.ErrSecretSourceAWSInvalidLocation,
		},
		{
			name:     "invalid location (spaces)",
			coordStr: "aws:///invalid name",
			errMatch: spelunkaws.ErrSecretSourceAWSInvalidLocation,
		},
		{
			name:     "invalid location (special chars)",
			coordStr: "aws://invalid!name",
			errMatch: spelunkaws.ErrSecretSourceAWSInvalidLocation,
		},
		{
			name:     "invalid location (name ends with hyphen and 6 characters)",
			coordStr: "aws://my-secret-12e4G6",
			errMatch: spelunkaws.ErrSecretSourceAWSInvalidNameSuffix,
		},
		{
			name:     "invalid location (bad arn format)",
			coordStr: "aws:///arn:aws:secretsmanager:us-east-1:123:secret:too-short-account-id",
			errMatch: spelunkaws.ErrSecretSourceAWSInvalidLocation,
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

func createTestSecrets(
	t *testing.T,
	client *secretsmanager.Client,
) *secretsmanager.CreateSecretOutput {
	out, err := client.CreateSecret(t.Context(), &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretValue),
	})
	require.NoError(t, err)
	return out
}

func setupAWSTestContainer(t *testing.T) (*secretsmanager.Client, error) {
	// See: https://hub.docker.com/r/localstack/localstack/tags
	localstackContainer, err := localstack.Run(t.Context(),
		"localstack/localstack:3.4.0",
	)
	testcontainers.CleanupContainer(t, localstackContainer)
	require.NoError(t, err)

	mappedPort, err := localstackContainer.MappedPort(t.Context(), "4566/tcp")
	require.NoError(t, err)
	hostIP, err := localstackContainer.Host(t.Context())
	require.NoError(t, err)
	mappedURL := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(mappedURL),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "test", SecretAccessKey: "test", SessionToken: "test",
				Source: "Hard-coded credentials; values are irrelevant for localstack",
			},
		}),
	)
	require.NoError(t, err)

	return secretsmanager.NewFromConfig(cfg), nil
}
