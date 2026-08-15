package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkgcp "github.com/detro/spelunk/plugin/source/gcp/v2"
	"github.com/detro/spelunk/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GCPConfigurator struct {
	CredentialsFile string `name:"gcp-credentials-file" env:"GOOGLE_APPLICATION_CREDENTIALS" help:"Path to GCP Service Account Credentials JSON file."`
}

var _ internal.SecretSourceConfigurator = (*GCPConfigurator)(nil)

func (c *GCPConfigurator) Type() string {
	return spelunkgcp.Type
}

func (c *GCPConfigurator) CredentialsDetected() bool {
	if c.CredentialsFile != "" {
		return true
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" ||
		os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" ||
		os.Getenv("SECRET_MANAGER_EMULATOR_HOST") != "" {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil {
		adcPath := filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
		if _, err := os.Stat(adcPath); err == nil {
			return true
		}
	}
	return false
}

func (c *GCPConfigurator) newClient(ctx context.Context) (*secretmanager.Client, error) {
	var opts []option.ClientOption
	if host := os.Getenv("SECRET_MANAGER_EMULATOR_HOST"); host != "" {
		conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithGRPCConn(conn))
	} else if c.CredentialsFile != "" {
		//nolint:staticcheck // Generic credentials file path
		opts = append(opts, option.WithCredentialsFile(c.CredentialsFile))
	}
	return secretmanager.NewClient(ctx, opts...)
}

func (c *GCPConfigurator) SpelunkerOption(ctx context.Context) (spelunk.SpelunkerOption, error) {
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

	return spelunkgcp.WithGCP(client), nil
}

func (c *GCPConfigurator) CredentialsValid(ctx context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	if host := os.Getenv("SECRET_MANAGER_EMULATOR_HOST"); host != "" {
		client, err := c.newClient(ctx)
		if err != nil {
			return err
		}
		return client.Close()
	}
	if c.CredentialsFile != "" {
		credsData, err := os.ReadFile(c.CredentialsFile)
		if err != nil {
			return err
		}
		//nolint:staticcheck // Generic credentials file verification
		creds, err := google.CredentialsFromJSONWithParams(
			ctx,
			credsData,
			google.CredentialsParams{
				Scopes: secretmanager.DefaultAuthScopes(),
			},
		)
		if err != nil {
			return err
		}
		_, err = creds.TokenSource.Token()
		return err
	}
	creds, err := google.FindDefaultCredentials(ctx, secretmanager.DefaultAuthScopes()...)
	if err != nil {
		return err
	}
	_, err = creds.TokenSource.Token()
	return err
}
