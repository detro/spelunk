package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/configurator"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	"github.com/detro/spelunk/v2"
)

// Configurators It aggregates all source configurations embedded into the CLI.
type Configurators struct {
	AWS         configurator.AWSConfigurator         `embed:"" group:"AWS Secrets Manager Configuration (https://aws.amazon.com/secrets-manager/):"`
	Azure       configurator.AzureConfigurator       `embed:"" group:"Azure Key Vault Configuration (https://azure.microsoft.com/en-us/products/key-vault/):"`
	GCP         configurator.GCPConfigurator         `embed:"" group:"GCP Secrets Manager Configuration (https://cloud.google.com/security/products/secret-manager):"`
	Vault       configurator.VaultConfigurator       `embed:"" group:"HashiCorp Vault Configuration (https://www.hashicorp.com/en/products/vault):"`
	Kubernetes  configurator.KubernetesConfigurator  `embed:"" group:"Kubernetes Secrets Configuration (https://kubernetes.io/docs/concepts/configuration/secret/):"`
	OnePassword configurator.OnePasswordConfigurator `embed:"" group:"1Password Configuration (https://developer.1password.com/docs/cli/):"`
	Keeper      configurator.KeeperConfigurator      `embed:"" group:"Keeper Configuration (https://docs.keeper.io/en/enterprise-guide/commander-cli):"`
}

// All returns all registered SecretSourceConfigurator instances.
func (c *Configurators) All() []internal.SecretSourceConfigurator {
	return []internal.SecretSourceConfigurator{
		&c.AWS,
		&c.Azure,
		&c.GCP,
		&c.Vault,
		&c.Kubernetes,
		&c.OnePassword,
		&c.Keeper,
	}
}

// SpelunkerOptions returns all SpelunkerOption for all internal.SecretSourceConfigurator detected.
func (c *Configurators) SpelunkerOptions(ctx context.Context) ([]spelunk.SpelunkerOption, error) {
	var opts []spelunk.SpelunkerOption
	for _, p := range c.All() {
		opt, err := p.SpelunkerOption(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to configure plugin %s: %w", p.Type(), err)
		}
		if opt != nil {
			opts = append(opts, opt)
		}
	}
	return opts, nil
}

// VerifyAll checks all registered internal.SecretSourceConfigurator, logging progress and errors,
// and returns a joined error if any detected credential failed verification.
func (c *Configurators) VerifyAll(ctx context.Context) error {
	// Detect and validate credentials for all configurators but, on error, accumulate and don't return yet
	var errs []error
	for _, p := range c.All() {
		if !p.CredentialsDetected() {
			slog.Log(
				ctx,
				logger.LevelTrace,
				"skipped plugin (no credentials detected)",
				"plugin",
				p.Type(),
			)
			continue
		}
		slog.Log(ctx, logger.LevelTrace, "detected credentials", "plugin", p.Type())

		err := p.CredentialsValid(ctx)
		if err != nil {
			slog.Error("credentials invalid", "plugin", p.Type(), "err", err)
			errs = append(
				errs,
				fmt.Errorf("plugin %q credential verification failed: %w", p.Type(), err),
			)
		} else {
			slog.Info("credentials valid", "plugin", p.Type())
		}
	}

	// If any error was accumulated, return it as a single join one
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
