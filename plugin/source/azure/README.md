# Azure Key Vault Secret Source (`az://`)

The **Azure Key Vault** secret source retrieves secrets directly from [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault/).

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithAzure()`.

## Dependencies

This plugin requires the official Azure SDK for Go:
- `github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets`

## Usage

To use the Azure Key Vault source, use the `az://` scheme followed by the secret name. You can optionally specify a version.

### Syntax

```text
az://<SECRET_NAME>
az://<SECRET_NAME>/<VERSION>
az:///<SECRET_NAME>
```

**⚠️ Important NOTE regarding Vault URLs:**
Due to how the Azure SDK works, the `azsecrets.Client` is pre-configured with a specific `vaultURL`. Spelunk relies on this pre-configured client, so the vault name is **not** included in the Spelunk coordinate URI. 

### Examples

Retrieve the latest version of a secret:

```text
az://database-connection-string
```

Retrieve a specific version of a secret:

```text
az://api-key/7b1897b204e5485496a7981f44e13456
```

Using modifiers (e.g., extracting JSON path) safely ignores trailing slashes in the path:

```text
az://my-json-secret/?jp=$.password
```

## Configuration

To use this source, you must initialize `spelunk` with an Azure Key Vault Secrets client:

```go
import (
    "fmt"
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
    "github.com/detro/spelunk"
    spelunkazure "github.com/detro/spelunk/plugin/source/azure"
)

func main() {
    // 1. Create a credential using DefaultAzureCredential
    // (This works with Env Vars, Managed Identity, Azure CLI, etc.)
    cred, _ := azidentity.NewDefaultAzureCredential(nil)

    // Note: Azure Key Vault clients are usually bound to a specific Vault URL.
    // If you plan to read from multiple vaults, you'll need multiple Spelunker instances
    // or to implement a custom Spelunker logic.
    vaultURL := "https://my-vault-prod.vault.azure.net/"

    // 2. Create the Azure Key Vault secrets client
    azClient, _ := azsecrets.NewClient(vaultURL, cred, nil)

    // 3. Initialize Spelunker with the Azure plugin
    s := spelunk.NewSpelunker(
        spelunkazure.WithAzure(azClient),
    )

    // 4. Dig up secrets
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing & Cleaning**: 
   - Uses the location (hostname + path) as the Secret Name and optional Version.
   - If a leading slash is present (e.g., when using `az:///<SECRET_NAME>`), it is trimmed.
   - Any trailing slash (e.g., when the URI contains query parameters like `/?jp=$.password`) is stripped automatically.
2. **Validation**: The cleaned location is validated to extract the Secret Name and Version:
   - **Secret Name**: Must be 1-127 characters containing only alphanumeric characters and hyphens.
   - **Version**: Must be exactly 32 hexadecimal characters.
3. **Retrieval**: Uses `azClient.GetSecret(ctx, secretName, version, nil)` to fetch the secret.
4. **Extraction**: Returns the underlying string value of the secret (`*result.Value`).
5. **Errors**:
    - Returns `ErrSecretSourceAzureInvalidLocation` if the location does not match either the valid `<SECRET_NAME>` or `<SECRET_NAME>/<VERSION>` format.
    - Returns `ErrCouldNotFetchSecret` if the API call fails.
    - Returns `ErrSecretNotFound` if the secret does not exist (HTTP 404) or has a nil payload.

## Testing

Integration tests for this plugin are powered by [Testcontainers](https://golang.testcontainers.org/) using the `nagyesta/lowkey-vault` image. They are automatically skipped in short test mode (`go test -short` or `task test.short`).

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates managed by Azure Key Vault within Azure infrastructure (AKS, Azure Functions, App Service).
