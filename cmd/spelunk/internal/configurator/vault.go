package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkvault "github.com/detro/spelunk/plugin/source/vault/v2"
	"github.com/detro/spelunk/v2"
	"github.com/hashicorp/vault/api"
)

type VaultConfigurator struct {
	Addr      string `name:"vault-addr"      env:"VAULT_ADDR"      help:"Vault Server Address (e.g. https://vault.example.com:8200)."`
	Token     string `name:"vault-token"     env:"VAULT_TOKEN"     help:"Vault Authentication Token."`
	Namespace string `name:"vault-namespace" env:"VAULT_NAMESPACE" help:"Vault Namespace."`
}

var _ internal.SecretSourceConfigurator = (*VaultConfigurator)(nil)

func (c *VaultConfigurator) Type() string {
	return spelunkvault.Type
}

func (c *VaultConfigurator) CredentialsDetected() bool {
	return c.Addr != "" || c.Token != "" || os.Getenv("VAULT_ADDR") != "" ||
		os.Getenv("VAULT_TOKEN") != ""
}

func (c *VaultConfigurator) newClient() (*api.Client, error) {
	cfg := api.DefaultConfig()
	if c.Addr != "" {
		cfg.Address = c.Addr
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		client.SetToken(c.Token)
	}
	if c.Namespace != "" {
		client.SetNamespace(c.Namespace)
	}
	return client, nil
}

func (c *VaultConfigurator) SpelunkerOption(ctx context.Context) (spelunk.SpelunkerOption, error) {
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

	return spelunkvault.WithVault(client), nil
}

func (c *VaultConfigurator) CredentialsValid(ctx context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	client, err := c.newClient()
	if err != nil {
		return err
	}
	if client.Token() != "" {
		_, err = client.Auth().Token().LookupSelfWithContext(ctx)
		return err
	}
	_, err = client.Sys().HealthWithContext(ctx)
	return err
}
