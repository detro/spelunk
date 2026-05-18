# Bitwarden Secret Source (`bw://`)

The **Bitwarden** secret source retrieves secrets directly from [Bitwarden Secrets Manager](https://bitwarden.com/products/secrets-manager/) using the official SDK.

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithBitwarden()`.

**Testing & Contributions**: Due to the lack of a free Bitwarden Secrets Manager tier and the fact that the SDK strictly requires a Machine Account Access Token (which Testcontainers or standard Personal Vaults do not support), **this plugin is currently not covered by integration tests**. 

We are actively looking for contributors willing to provide access to a test environment or who can test this thoroughly. If you can help, please reach out or open a PR!

## Dependencies

This plugin requires the official Bitwarden Go SDK:
- `github.com/bitwarden/sdk-go/v2`

**Note on CGO**: The official Bitwarden SDK relies on a Rust core library (`bitwarden_c`). Your Go environment must have CGO enabled (`CGO_ENABLED=1`) and a C compiler available to build any application using this plugin.

## Usage

To use the Bitwarden source, use the `bw://` scheme followed by the exact Secret ID.

### Syntax

```text
bw://<SECRET_ID>
```

### Examples

Retrieve a secret using its UUID:

```text
bw://f47ac10b-58cc-4372-a567-0e02b2c3d479
```

## Configuration

To use this source, you must initialize `spelunk` with a Bitwarden client:

```go
import (
    "os"
    "github.com/bitwarden/sdk-go/v2"
    "github.com/detro/spelunk"
    "github.com/detro/spelunk/plugin/source/bitwarden"
)

func main() {
    // 1. Create Bitwarden client
    token := os.Getenv("BWS_ACCESS_TOKEN")
    client, _ := sdk.NewBitwardenClient(nil, nil)
    _ = client.AccessTokenLogin(token, nil)

    // 2. Initialize Spelunker with the Bitwarden plugin
    s := spelunk.NewSpelunker(
        bitwarden.WithBitwarden(client),
    )

    // 3. Dig up secrets
    coord, _ := types.NewSecretCoord("bw://f47ac10b-58cc-4372-a567-0e02b2c3d479")
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing**: Validates that the location is a valid UUIDv4 using `github.com/google/uuid`.
2. **Retrieval**: Uses `client.Secrets().Get()` to fetch the specific Secret ID.
3. **Errors**:
    - Returns `types.ErrInvalidLocation` if the format is incorrect (e.g., not a valid UUIDv4).
    - Returns `ErrCouldNotFetchSecret` if the API call fails or the access token is invalid.

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates from Bitwarden Secrets Manager using Access Tokens.
