# Keeper Secret Source (`kp://`)

The **Keeper** secret source retrieves secrets directly from [Keeper Secrets Manager](https://docs.keeper.io/secrets-manager/secrets-manager) using the official Go SDK.

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithKeeper()`.

**Testing & Contributions**: Due to the lack of a free Keeper Secrets Manager tier and the requirement of a live vault for testing (as Testcontainers or local emulators do not exist), **this plugin is currently not covered by integration tests**.

We are actively looking for contributors willing to provide access to a test environment or who can test this thoroughly. If you can help, please reach out or open a PR!

## Dependencies

This plugin requires the official Keeper Secrets Manager Go SDK:
- `github.com/keeper-security/secrets-manager-go/core`

## Usage

To use the Keeper source, use the `kp://` scheme followed by the Record UID. You can specify a specific field, or end the path with a `/` to retrieve the entire record data as JSON.

### Syntax

```text
kp://<RECORD_UID>/<FIELD>
kp://<RECORD_UID>/
```

Supported standard fields are `title`, `password`, and `notes`. The plugin will also search fields dynamically by Label.

### Examples

Retrieve the `password` field from a specific Record UID:

```text
kp://Oq9bA_k.../password
```

Retrieve a custom field labeled `api_key`:

```text
kp://Oq9bA_k.../api_key
```

Retrieve the entire record as JSON:

```text
kp://Oq9bA_k.../
```

## Configuration

To use this source, you must initialize `spelunk` with a Keeper Secrets Manager client:

```go
import (
    "os"
    ksm "github.com/keeper-security/secrets-manager-go/core"
    "github.com/detro/spelunk"
    "github.com/detro/spelunk/plugin/source/keeper"
)

func main() {
    // 1. Create Keeper client
    token := os.Getenv("KSM_TOKEN") // or use a config file
    clientOptions := &ksm.ClientOptions{
        Token:  token,
        Config: ksm.NewMemoryKeyValueStorage(),
    }
    client := ksm.NewSecretsManager(clientOptions)

    // 2. Initialize Spelunker with the Keeper plugin
    s := spelunk.NewSpelunker(
        keeper.WithKeeper(client),
    )

    // 3. Dig up secrets
    coord, _ := types.NewSecretCoord("kp://Oq9bA_k.../password")
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing**: Extracts the `RecordUID` and optional `Field` using a regular expression, and validates that the `RecordUID` is a valid 22-character base64url string.
2. **Retrieval**: Uses `client.GetSecrets()` to fetch the record. Keeper handles the zero-knowledge client-side decryption automatically.
3. **Extraction**: If a `Field` was provided, it first checks standard fields, then searches fields by Label. If no field is provided (or the path ends with `/`), it returns the raw JSON of the record.
4. **Errors**:
    - Returns `types.ErrInvalidLocation` if the format is missing the Record UID, or if it is not a valid 22-character base64url string.
    - Returns `ErrSecretNotFound` if the Record UID doesn't exist or is not shared with the application.
    - Returns `ErrSecretKeyNotFound` if the requested field does not exist on the record.

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates from Keeper Secrets Manager using Zero-Knowledge encryption.
