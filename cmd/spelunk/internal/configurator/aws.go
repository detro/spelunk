package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkaws "github.com/detro/spelunk/plugin/source/aws/v2"
	"github.com/detro/spelunk/v2"
)

type AWSConfigurator struct {
	Region      string `name:"aws-region"       env:"AWS_REGION"                      help:"AWS Region."`
	Profile     string `name:"aws-profile"      env:"AWS_PROFILE"                     help:"AWS Profile."`
	EndpointURL string `name:"aws-endpoint-url" env:"AWS_ENDPOINT_URL_SECRETSMANAGER" help:"AWS Secrets Manager Endpoint URL."`
}

var _ internal.SecretSourceConfigurator = (*AWSConfigurator)(nil)

func (c *AWSConfigurator) Type() string {
	return spelunkaws.Type
}

func (c *AWSConfigurator) CredentialsDetected() bool {
	if c.Region != "" || c.Profile != "" || c.EndpointURL != "" {
		return true
	}
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_SECRET_ACCESS_KEY") != "" ||
		os.Getenv("AWS_SESSION_TOKEN") != "" || os.Getenv("AWS_DEFAULT_REGION") != "" ||
		os.Getenv("AWS_REGION") != "" || os.Getenv("AWS_PROFILE") != "" ||
		os.Getenv("AWS_ENDPOINT_URL_SECRETSMANAGER") != "" ||
		os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" ||
		os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI") != "" ||
		os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI") != "" {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil {
		if _, err := os.Stat(filepath.Join(home, ".aws", "credentials")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(home, ".aws", "config")); err == nil {
			return true
		}
	}
	return false
}

func (c *AWSConfigurator) newClient(ctx context.Context) (*secretsmanager.Client, error) {
	var optFns []func(*config.LoadOptions) error
	if c.Region != "" {
		optFns = append(optFns, config.WithRegion(c.Region))
	}
	if c.Profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(c.Profile))
	}
	if c.EndpointURL != "" {
		optFns = append(optFns, config.WithBaseEndpoint(c.EndpointURL))
	}
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}
	return secretsmanager.NewFromConfig(cfg), nil
}

func (c *AWSConfigurator) SpelunkerOption(ctx context.Context) (spelunk.SpelunkerOption, error) {
	if !c.CredentialsDetected() {
		slog.Log(ctx, logger.LevelTrace, "skipped (no credentials detected)", "plugin", c.Type())
		return nil, nil
	}
	slog.Log(ctx, logger.LevelTrace, "detected credentials", "plugin", c.Type())

	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}
	slog.Log(ctx, logger.LevelTrace, "configured client", "plugin", c.Type())

	return spelunkaws.WithAWS(client), nil
}

func (c *AWSConfigurator) CredentialsValid(ctx context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int32(1),
	})
	return err
}
