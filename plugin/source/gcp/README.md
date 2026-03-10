# Google Cloud Secret Manager Secret Source (`gcp://`)

The **GCP Secret Manager** secret source retrieves secrets directly from [Google Cloud Secret Manager](https://cloud.google.com/security/products/secret-manager).

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithGCP()`.

## Dependencies

This plugin requires the official Google Cloud Go SDK:
- `cloud.google.com/go/secretmanager/apiv1`

## Usage

To use the GCP Secret Manager source, use the `gcp://` scheme followed by the full resource path to the secret. You can optionally specify a version; if no version is provided, `latest` is assumed.

### Syntax

```text
gcp://projects/<PROJECT_ID_OR_NUM>/secrets/<SECRET_NAME>
gcp://projects/<PROJECT_ID_OR_NUM>/secrets/<SECRET_NAME>/versions/<VERSION>
```

- Expected format of `<PROJECT_ID_OR_NUM>` is documented at: [AIP-2510](https://google.aip.dev/cloud/2510).
- Expected format of `<SECRET_NAME>` is documented at: [GCP Secret Manager](https://cloud.google.com/security/products/secret-manager).

### Examples

Retrieve the latest version of a secret:

```text
gcp://projects/my-project-123/secrets/my-database-password
```

Retrieve a specific version of a secret:

```text
gcp://projects/1234567890/secrets/my-api-key/versions/2
```

Using modifiers (e.g., extracting JSON path) safely ignores trailing slashes in the path:

```text
gcp://projects/my-project-123/secrets/my-json-secret/?jp=$.password
```

## Configuration

To use this source, you must initialize `spelunk` with a GCP Secret Manager client:

```go
import (
    "context"
    secretmanager "cloud.google.com/go/secretmanager/apiv1"
    "github.com/detro/spelunk"
    spelunkgcp "github.com/detro/spelunk/plugin/source/gcp"
)

func main() {
    ctx := context.Background()

    // 1. Create the Secret Manager client (uses Application Default Credentials automatically)
    gcpClient, _ := secretmanager.NewClient(ctx)
    defer gcpClient.Close()

    // 2. Initialize Spelunker with the GCP plugin
    s := spelunk.NewSpelunker(
        spelunkgcp.WithGCP(gcpClient),
    )

    // 3. Dig up secrets
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing**: The provided location is expected to start with `projects/` and contain `/secrets/`.
   - Any trailing slash (e.g., when the URI contains query parameters like `/?jp=$.password`) is stripped automatically.
   - If no `/versions/` suffix is present, `/versions/latest` is automatically appended to the request.
2. **Retrieval**: Uses `gcpClient.AccessSecretVersion` to fetch the payload of the secret.
3. **Extraction**: Returns the decoded payload data as a string.
4. **Errors**:
    - Returns `ErrSecretSourceGCPInvalidLocation` if the location format is invalid.
    - Returns `ErrCouldNotFetchSecret` if the API call fails for other reasons.
    - Returns `ErrSecretNotFound` if the secret or version does not exist, or if the payload is empty.

## Testing

Integration tests for this plugin are powered by [Testcontainers](https://golang.testcontainers.org/) using the [blackwell-systems/gcp-secret-manager-emulator](https://github.com/blackwell-systems/gcp-secret-manager-emulator) image. They are automatically skipped in short test mode (`go test -short` or `task test.short`).

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates managed by Google Cloud Secret Manager across GCP environments (GKE, Cloud Run, Compute Engine).
