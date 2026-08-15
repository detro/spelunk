package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkkeeper "github.com/detro/spelunk/plugin/source/keeper/v2"
	"github.com/detro/spelunk/v2"
	ksm "github.com/keeper-security/secrets-manager-go/core"
)

type KeeperConfigurator struct {
	KsmConfig string `name:"ksm-config" env:"KSM_CONFIG" help:"Keeper Secrets Manager configuration (Base64 string or file path)."`
}

var _ internal.SecretSourceConfigurator = (*KeeperConfigurator)(nil)

func (c *KeeperConfigurator) Type() string {
	return spelunkkeeper.Type
}

func (c *KeeperConfigurator) CredentialsDetected() bool {
	return c.KsmConfig != "" || os.Getenv("KSM_CONFIG") != ""
}

func (c *KeeperConfigurator) newClient() (*ksm.SecretsManager, error) {
	configStr := c.KsmConfig
	if configStr == "" {
		configStr = os.Getenv("KSM_CONFIG")
	}
	opts := &ksm.ClientOptions{}
	if _, err := os.Stat(configStr); err == nil {
		opts.Config = ksm.NewFileKeyValueStorage(configStr)
	} else {
		opts.Config = ksm.NewMemoryKeyValueStorage(configStr)
	}
	return ksm.NewSecretsManager(opts), nil
}

func (c *KeeperConfigurator) SpelunkerOption(ctx context.Context) (spelunk.SpelunkerOption, error) {
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

	return spelunkkeeper.WithKeeper(client), nil
}

func (c *KeeperConfigurator) CredentialsValid(_ context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	client, err := c.newClient()
	if err != nil {
		return err
	}
	_, err = client.GetSecrets([]string{})
	return err
}
