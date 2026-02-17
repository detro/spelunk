# Plain Text Secret Source (`plain://`)

The **Plain Text** secret source treats the URI location itself as the secret value. It returns exactly what you put in the authority and path of the URI.

## Status

**Built-in**: This source is included and enabled by default in `spelunk`.

## Usage

To use the Plain Text source, use the `plain://` scheme followed by your secret string.

### Syntax

```
plain://<SECRET_VALUE>
```

- **Type**: `plain`
- **Location**: The string value to return.

### Example

To provide the string "my-secret-token":

```
plain://my-secret-token
```

**Note**: Since the "Location" in `spelunk` coordinates includes both the URI authority (host/userinfo) and path, you can include slashes:

```
plain://path/to/my/secret
```

This will return `path/to/my/secret`.

## Behavior

1. **Passthrough**: This source simply returns the `Location` part of the `SecretCoord` struct.
2. **No Decoding**: Unlike `base64://`, this source performs no decoding. The value is used as-is.

## Use Cases

- **Development**: Hardcoding non-sensitive values during local development without changing the `spelunk` interface.
- **Testing**: verifying that your application's secret wiring works correctly with a known, static string.
