# Base64 Decoder Secret Modifier (`b64d`)

The `b64d` modifier decodes a base64-encoded secret value.

## Usage

To use it, append the modifier `b64d` to the given secret coordinates string:

```
plain://bXktc2VjcmV0?b64d
```

## Behavior

- It takes the retrieved secret value and decodes it using Go's standard `encoding/base64` package.
- If the secret value is not a valid base64 string, the modifier will return an `ErrSecretModifierBase64DecoderFailedDecoding` error.
- Any modifier arguments (e.g. `?b64d=foo`) are currently ignored.
