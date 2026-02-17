# File Secret Source (`file://`)

The **File** secret source retrieves secret values from files on the local filesystem.

## Status

**Built-in**: This source is included and enabled by default in `spelunk`.

## Usage

To use the File source, use the `file://` scheme followed by the path to the file.

### Syntax

```
file://<PATH_TO_FILE>
```

- **Type**: `file`
- **Location**: The file path (absolute or relative).

### Examples

**Absolute Path**:
```
file:///etc/secrets/db_password.txt
```
*(Note: Three slashes. Two for the scheme `file://`, one for the root `/`)*

**Relative Path**:
```
file://secrets/api_key.txt
```

**Explicit Relative Path**:
```
file://./.env.local
```

## Behavior

1. **Check Existence**: Verifies the file exists using `os.Stat`. Returns `ErrSecretNotFound` if missing.
2. **Read**: Opens and reads the entire file content using `io.ReadAll`.
3. **Result**: Returns the file content as a string.

## Use Cases

- **Docker/Kubernetes**: Reading secrets mounted as files (e.g., Kubernetes Secrets mounted to `/var/run/secrets`).
- **Local Development**: Reading configuration from local secret files not checked into version control.
