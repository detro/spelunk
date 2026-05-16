# 1Password Secret Source (`op://`)

The **1Password** secret source retrieves secrets directly from 1Password using the new official [1Password Go SDK](https://developer.1password.com/docs/sdks/go/).

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `With1Password()`.

**Testing**: Because the 1Password SDK relies on a Rust core to communicate directly with production servers, it does not support local mocking or Testcontainers. Integration tests require a valid 1Password Service Account and are thus skipped in standard CI environments without the `OP_SERVICE_ACCOUNT_TOKEN` environment variable set.

## Dependencies

This plugin requires the official 1Password Go SDK:
- `github.com/1password/onepassword-sdk-go`

## Usage

To use the 1Password source, use the `op://` scheme followed by the Vault, Item, optional Section, and Field you want to retrieve.

In Spelunk, these "Secret Coordinates" are exactly the same as the "Secret Reference" that you can obtain by going to a 1Password vault item, selecting a field, and copying its "Secret Reference".
See https://developer.1password.com/docs/cli/secret-reference-syntax.

### Syntax

```text
op://<VAULT>/<ITEM>/[<SECTION>]/<FIELD>
```

### Examples

Retrieve the `password` field from the `Database` item in the `Production` vault:

```text
op://Production/Database/password
```

Retrieve the `token` field from the `API` section of the `Stripe` item in the `Shared` vault:

```text
op://Shared/Stripe/API/token
```

## Configuration

To use this source, you must initialize `spelunk` with a 1Password client. The [1Password Go SDK supports two authentication methods](https://github.com/1Password/onepassword-sdk-go/blob/main/README.md#authentication):
1. **Service Account**: Uses a token (`OP_SERVICE_ACCOUNT_TOKEN`) to authenticate. Best for CI/CD, servers, and automated environments.
2. **Local App**: Communicates with the 1Password desktop app running on the same machine. Best for local development.

### Example using a Service Account

```go
import (
    "context"
    "os"
    "github.com/1password/onepassword-sdk-go"
    "github.com/detro/spelunk"
    "github.com/detro/spelunk/plugin/source/1password"
)

func main() {
    // 1. Create 1Password client
    token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
    client, _ := onepassword.NewClient(
        context.Background(),
        onepassword.WithServiceAccountToken(token), // Or use onepassword.WithDesktopAppIntegration("account-name") for Local App
        onepassword.WithIntegrationInfo("My App", "v1.0.0"),
    )

    // 2. Initialize Spelunker with the 1Password plugin
    s := spelunk.NewSpelunker(
        onepassword.With1Password(client),
    )

    // 3. Dig up secrets
    coord, _ := types.NewSecretCoord("op://Production/Database/password")
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing**: Validates the location strictly matches the format `VAULT/ITEM/FIELD` or `VAULT/ITEM/SECTION/FIELD`.
2. **Retrieval**: Uses `client.Secrets().Resolve()` with the official `op://` reference syntax.
3. **Errors**:
    - Returns `ErrSecretSource1PasswordInvalidLocation` if the format is incorrect.
    - Returns `ErrCouldNotFetchSecret` if the API call fails, authentication is invalid, or the item/field doesn't exist (the SDK currently lacks strongly typed error differentiation for "not found").

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates from a centralized 1Password Vault.
- Local development without storing `.env` files (using the 1Password desktop app integration).
- Secure secret injection in CI/CD pipelines (using Service Accounts).
