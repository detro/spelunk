# Environment Variable Secret Source (`env://`)

The **Environment Variable** secret source allows you to retrieve secrets directly from the process's environment variables.

## Status

**Built-in**: This source is included and enabled by default in `spelunk`.

## Usage

To use the Environment Variable source, use the `env://` scheme followed by the name of the environment variable.

### Syntax

```
env://<VARIABLE_NAME>
```

- **Type**: `env`
- **Location**: The name of the environment variable to look up.

### Example

Suppose you have an environment variable named `API_KEY` with the value `12345-abcde`.

```bash
export API_KEY="12345-abcde"
```

To retrieve this secret:

```
env://API_KEY
```

## Behavior

1. **Lookup**: The source uses `os.LookupEnv` to find the variable.
2. **Validation**: If the environment variable is not set, `DigUp` will return an error (`ErrSecretNotFound`).
3. **Result**: Returns the string value of the environment variable.

## Use Cases

- **Twelve-Factor Apps**: Adhering to the [12-factor methodology](https://12factor.net/config) by storing config in the environment.
- **Containerization**: Easily injecting secrets into containers via environment variables.
