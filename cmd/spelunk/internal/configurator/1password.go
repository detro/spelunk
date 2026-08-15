package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/1password/onepassword-sdk-go"
	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkop "github.com/detro/spelunk/plugin/source/1password/v2"
	"github.com/detro/spelunk/v2"
)

type OnePasswordConfigurator struct {
	ServiceAccountToken string `name:"op-service-account-token" env:"OP_SERVICE_ACCOUNT_TOKEN" help:"1Password Service Account Token."`
	IntegrationName     string `name:"op-integration-name"                                     help:"1Password Integration Name."      default:"spelunk"`
	IntegrationVersion  string `name:"op-integration-version"                                  help:"1Password Integration Version."   default:"dev"`
}

var _ internal.SecretSourceConfigurator = (*OnePasswordConfigurator)(nil)

func (c *OnePasswordConfigurator) Type() string {
	return spelunkop.Type
}

func (c *OnePasswordConfigurator) CredentialsDetected() bool {
	return c.ServiceAccountToken != "" || os.Getenv("OP_SERVICE_ACCOUNT_TOKEN") != ""
}

func (c *OnePasswordConfigurator) newClient(ctx context.Context) (*onepassword.Client, error) {
	token := c.ServiceAccountToken
	if token == "" {
		token = os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	}
	integrationName := c.IntegrationName
	if integrationName == "" {
		integrationName = "spelunk"
	}
	integrationVersion := c.IntegrationVersion
	if integrationVersion == "" {
		integrationVersion = "dev"
	}
	return onepassword.NewClient(
		ctx,
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(integrationName, integrationVersion),
	)
}

func (c *OnePasswordConfigurator) SpelunkerOption(
	ctx context.Context,
) (spelunk.SpelunkerOption, error) {
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

	return spelunkop.With1Password(client), nil
}

func (c *OnePasswordConfigurator) CredentialsValid(ctx context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.Vaults().List(ctx)
	return err
}
