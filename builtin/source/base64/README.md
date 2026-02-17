# Base64 Secret Source (`base64://`)

The **Base64** secret source allows you to embed secrets directly into the URI by encoding them in Base64 format. This is primarily useful for testing, configuration injection (where secrets are passed as env vars or flags), or when you want to use `spelunk` uniformly even for static values.

## Status

**Built-in**: This source is included and enabled by default in `spelunk`.

## Usage

To use the Base64 source, prefix your base64-encoded secret string with `base64://`.

### Syntax

```
base64://<base64_encoded_string>
```

- **Type**: `base64`
- **Location**: The Base64-encoded string itself.

### Example

Suppose you want to provide the secret "mysecretpassword".

1. **Encode the secret**:
    ```bash
    echo -n "mysecretpassword" | base64
    # Output: bXlzZWNyZXRwYXNzd29yZA==
    ```

2. **Construct the URI**:
    ```
    base64://bXlzZWNyZXRwYXNzd29yZA==
    ```

When `spelunk` processes this URI, it will return the original string: `mysecretpassword`.

## Behavior

1. **Decoding**: The source uses standard Base64 decoding ([RFC 4648](https://datatracker.ietf.org/doc/html/rfc4648)).
2. **Validation**: If the provided string is not valid Base64, `DigUp` will return an error (`ErrSecretSourceBase64FailedDecoding`).
3. **Result**: Returns the decoded byte slice converted to a string.

## Use Cases

- **Testing**: Quickly verify that your application correctly handles secret retrieval without needing external dependencies like files or Kubernetes.
- **Fallback**: Provide a default value directly in configuration if a more secure source is unavailable.
- **Static Configuration**: Embed non-sensitive "secrets" directly in code or config files while maintaining a consistent `spelunk` interface.
