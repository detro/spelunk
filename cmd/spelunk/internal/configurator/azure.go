package configurator

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkaz "github.com/detro/spelunk/plugin/source/azure/v2"
	"github.com/detro/spelunk/v2"
)

type AzureConfigurator struct {
	VaultURL              string `name:"azure-vault-url"                env:"AZURE_KEYVAULT_URL"  help:"Azure Key Vault URL (e.g. https://<vault-name>.vault.azure.net)."`
	TenantID              string `name:"azure-tenant-id"                env:"AZURE_TENANT_ID"     help:"Azure Tenant ID."`
	ClientID              string `name:"azure-client-id"                env:"AZURE_CLIENT_ID"     help:"Azure Client ID."`
	ClientSecret          string `name:"azure-client-secret"            env:"AZURE_CLIENT_SECRET" help:"Azure Client Secret."`
	InsecureSkipTLSVerify bool   `name:"azure-insecure-skip-tls-verify"                           help:"Skip TLS verification for Azure Key Vault (useful for local emulators)."`
}

var _ internal.SecretSourceConfigurator = (*AzureConfigurator)(nil)

func (c *AzureConfigurator) Type() string {
	return spelunkaz.Type
}

func (c *AzureConfigurator) CredentialsDetected() bool {
	return c.VaultURL != "" || os.Getenv("AZURE_KEYVAULT_URL") != ""
}

type lowkeyVaultCredential struct{}

func (e *lowkeyVaultCredential) GetToken(
	_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     "emulator-token",
		ExpiresOn: time.Now().Add(24 * time.Hour),
	}, nil
}

func (c *AzureConfigurator) newClient() (*azsecrets.Client, error) {
	vaultURL := c.VaultURL
	if vaultURL == "" {
		vaultURL = os.Getenv("AZURE_KEYVAULT_URL")
	}

	var cred azcore.TokenCredential
	var err error
	if c.TenantID != "" && c.ClientID != "" && c.ClientSecret != "" {
		cred, err = azidentity.NewClientSecretCredential(
			c.TenantID,
			c.ClientID,
			c.ClientSecret,
			nil,
		)
	} else if os.Getenv("AZURE_TESTING_LOWKEY_VAULT") == "true" {
		cred = &lowkeyVaultCredential{}
	} else {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	}
	if err != nil {
		return nil, err
	}

	var clientOpts azsecrets.ClientOptions
	if c.InsecureSkipTLSVerify {
		clientOpts.ClientOptions = azcore.ClientOptions{
			APIVersion: "7.4",
			Transport: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		}
		clientOpts.DisableChallengeResourceVerification = true
	}

	return azsecrets.NewClient(vaultURL, cred, &clientOpts)
}

func (c *AzureConfigurator) SpelunkerOption(ctx context.Context) (spelunk.SpelunkerOption, error) {
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

	return spelunkaz.WithAzure(client), nil
}

func (c *AzureConfigurator) CredentialsValid(ctx context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	client, err := c.newClient()
	if err != nil {
		return err
	}
	pager := client.NewListSecretPropertiesPager(nil)
	if pager.More() {
		_, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
