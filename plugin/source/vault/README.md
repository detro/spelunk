# Vault Secret Source (`vault://`)

The **Vault** secret source retrieves secrets directly from the [HashiCorp Vault](https://www.hashicorp.com/products/vault) API.

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithVault()`.

## Dependencies

This plugin requires the official Vault API client library:
- `github.com/hashicorp/vault/api`

## Usage

To use the Vault source, use the `vault://` scheme followed by the path to the secret. You can either specify a key to retrieve a single value, or end the path with a `/` to retrieve the entire secret data map as JSON.

### Syntax

```text
vault://<PATH_TO_SECRET>/<KEY>
vault://<PATH_TO_SECRET>/
```

Vault has multiple secrets engines. The most common is the Key/Value (KV) secrets engine, which has two versions: [KV v1](https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v1) and [KV v2](https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2).

- For **KV v1**, the path is simply `<MOUNT_POINT>/<SECRET_NAME>`. So to get a `<KEY>`, the coordinate is:
  `vault://<MOUNT_POINT>/<SECRET_NAME>/<KEY>`
  Or to get the entire secret as JSON:
  `vault://<MOUNT_POINT>/<SECRET_NAME>/`
- For **KV v2**, the API requires inserting `data/` after the mount point. So the path is `<MOUNT_POINT>/data/<SECRET_NAME>`, and the coordinate is:
  `vault://<MOUNT_POINT>/data/<SECRET_NAME>/<KEY>`
  Or to get the entire secret as JSON:
  `vault://<MOUNT_POINT>/data/<SECRET_NAME>/`

Spelunk transparently handles the response formats from both KV v1 and KV v2.

### Examples

Retrieve key `password` from a **KV v2** secret located at `my-app/db` on mount point `secret`:

```text
vault://secret/data/my-app/db/password
```

Retrieve key `token` from a **KV v1** secret located at `config` on mount point `kv`:

```text
vault://kv/config/token
```

Retrieve the entire **KV v2** secret as JSON located at `my-app/db` on mount point `secret`:

```text
vault://secret/data/my-app/db/
```

## Configuration

To use this source, you must initialize `spelunk` with a Vault client:

```go
import (
    "github.com/detro/spelunk"
    "github.com/detro/spelunk/plugin/source/vault"
    "github.com/hashicorp/vault/api"
)

func main() {
    // 1. Create Vault client
    config := api.DefaultConfig()
    vaultClient, _ := api.NewClient(config)

    // Ensure the client has a valid token
    vaultClient.SetToken("your-vault-token")

    // 2. Initialize Spelunker with the Vault plugin
    s := spelunk.NewSpelunker(
        vault.WithVault(vaultClient),
    )

    // 3. Dig up secrets
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing**: Splits the location into `Path` and `Key` (the last segment of the path is treated as the Key).
2. **Retrieval**: Uses `vaultClient.Logical().ReadWithContext(ctx, path)` to fetch the secret at the specified path.
3. **Extraction**: If a `Key` was provided, it looks up the specific `Key` in the resulting data map. If the path ends with `/` (no key), it marshals the entire data map into a JSON string and returns it. It automatically supports both KV v1 (data at the root) and KV v2 (data inside the `data` envelope) by checking if `secret.Data["data"]` exists as a map.
4. **Errors**:
    - Returns `ErrCouldNotFetchSecret` if the API call fails.
    - Returns `ErrSecretNotFound` if the path doesn't exist.
    - Returns `ErrSecretKeyNotFound` if the path exists but the specific key is missing.

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates from a centralized Vault server.
