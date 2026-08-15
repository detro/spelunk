package configurator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

const (
	awsJsonSecretName  = "my-app/my-jsonsecret"
	awsJsonSecretValue = `{"key":"value"}`

	awsPlainSecretName  = "my-app/my-plainsecret"
	awsPlainSecretValue = "simple-plain-secret"
)

func setupAWSTestContainer(t *testing.T) (*secretsmanager.Client, string, error) {
	t.Helper()
	localstackContainer, err := localstack.Run(t.Context(),
		"localstack/localstack:3.4.0",
	)
	if err != nil {
		return nil, "", err
	}
	testcontainers.CleanupContainer(t, localstackContainer)

	mappedPort, err := localstackContainer.MappedPort(t.Context(), "4566/tcp")
	if err != nil {
		return nil, "", err
	}
	hostIP, err := localstackContainer.Host(t.Context())
	if err != nil {
		return nil, "", err
	}
	mappedURL := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(mappedURL),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "test", SecretAccessKey: "test", SessionToken: "test",
				Source: "Hard-coded credentials for localstack",
			},
		}),
	)
	if err != nil {
		return nil, "", err
	}

	return secretsmanager.NewFromConfig(cfg), mappedURL, nil
}

func createAWSTestSecrets(t *testing.T, client *secretsmanager.Client) {
	t.Helper()
	_, err := client.CreateSecret(t.Context(), &secretsmanager.CreateSecretInput{
		Name:         aws.String(awsJsonSecretName),
		SecretString: aws.String(awsJsonSecretValue),
	})
	require.NoError(t, err)

	_, err = client.CreateSecret(t.Context(), &secretsmanager.CreateSecretInput{
		Name:         aws.String(awsPlainSecretName),
		SecretString: aws.String(awsPlainSecretValue),
	})
	require.NoError(t, err)
}

func TestPlugin_AWS_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping AWS E2E integration test in short mode")
	}

	bin := buildCLI(t)
	client, mappedURL, err := setupAWSTestContainer(t)
	require.NoError(t, err)
	createAWSTestSecrets(t, client)

	env := cleanEnv(t,
		"AWS_ACCESS_KEY_ID=test",
		"AWS_SECRET_ACCESS_KEY=test",
		"AWS_REGION=us-east-1",
		fmt.Sprintf("AWS_ENDPOINT_URL_SECRETSMANAGER=%s", mappedURL),
	)

	ctx := context.Background()

	t.Run("dig with jsonpath modifier", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "dig", fmt.Sprintf("aws://%s?jp=$.key", awsJsonSecretName))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, "value", res.Stdout)
	})

	t.Run("default dig without subcommand", func(t *testing.T) {
		res := runCLI(ctx, bin, env, fmt.Sprintf("aws://%s", awsPlainSecretName))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, awsPlainSecretValue, res.Stdout)
	})

	t.Run("exist secret found", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "exist", fmt.Sprintf("aws://%s", awsPlainSecretName))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})

	t.Run("exist secret missing", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "exist", "aws://missing/secret")
		require.NotEqual(t, 0, res.ExitCode)
	})

	t.Run("creds check", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "creds")
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})
}
