package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bitwarden/sdk-go/v2"
	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkbw "github.com/detro/spelunk/plugin/source/bitwarden/v2"
	"github.com/detro/spelunk/v2"
)

type BitwardenConfigurator struct {
	AccessToken string `name:"bws-access-token" env:"BWS_ACCESS_TOKEN" help:"Bitwarden Secrets Manager Access Token."`
	ServerURL   string `name:"bws-server-url"   env:"BWS_SERVER_URL"   help:"Bitwarden Secrets Manager API Server URL."`
}

var _ internal.SecretSourceConfigurator = (*BitwardenConfigurator)(nil)

func (c *BitwardenConfigurator) Type() string {
	return spelunkbw.Type
}

func (c *BitwardenConfigurator) CredentialsDetected() bool {
	return c.AccessToken != "" || os.Getenv("BWS_ACCESS_TOKEN") != ""
}

func (c *BitwardenConfigurator) newClient() (sdk.BitwardenClientInterface, error) {
	var serverURL *string
	if c.ServerURL != "" {
		serverURL = &c.ServerURL
	} else if envURL := os.Getenv("BWS_SERVER_URL"); envURL != "" {
		serverURL = &envURL
	}
	client, err := sdk.NewBitwardenClient(serverURL, nil)
	if err != nil {
		return nil, err
	}
	token := c.AccessToken
	if token == "" {
		token = os.Getenv("BWS_ACCESS_TOKEN")
	}
	if token != "" {
		if err := client.AccessTokenLogin(token, nil); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func (c *BitwardenConfigurator) SpelunkerOption(
	ctx context.Context,
) (spelunk.SpelunkerOption, error) {
	if !c.CredentialsDetected() {
		slog.Log(ctx, logger.LevelTrace, "skipped (no credentials detected)", "plugin", c.Type())
		return nil, nil
	}
	slog.Log(ctx, logger.LevelTrace, "detected credentials", "plugin", c.Type())

	client, err := c.newClient()
	if err != nil {
		return nil, err
	}
	slog.Log(ctx, logger.LevelTrace, "configured client", "plugin", c.Type())

	return spelunkbw.WithBitwarden(client), nil
}

func (c *BitwardenConfigurator) CredentialsValid(_ context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	_, err := c.newClient()
	return err
}
