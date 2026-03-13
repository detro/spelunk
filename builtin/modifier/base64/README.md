# Base64 Encoder Secret Modifier (`b64`)

This modifier allows you to encode a dug-up secret value into a Base64 string.

_Note: This modifier is also aliased as `?b64e`._

## Status

**Built-in**: This modifier is included and enabled by default in `spelunk`.

## Usage

To use the Base64 modifier, append `?b64` to your secret coordinates URI.

### Syntax

```
<scheme>://<location>?b64
```

- **Modifier Key**: `b64` (or `b64e`)
- **Value**: None required. Any value provided is ignored.

### Example

Suppose you have a plain secret `my-secret-password`.

To base64 encode it:
```
plain://my-secret-password?b64
```

**Result**: `bXktc2VjcmV0LXBhc3N3b3Jk`

## Behavior

1. **Encoding**: The modifier takes the secret value retrieved by the source and encodes it using standard Base64 encoding.
