# Spelunk CLI

Standalone command-line interface for `spelunk`. Resolves, extracts, and inspects secrets from multiple backends using unified Secret Coordinates (`scheme://location?modifiers`).

## Overview

`spelunk` is a CLI tool designed to extract secrets from various storage backends (Kubernetes, HashiCorp Vault, AWS Secrets Manager, Azure Key Vault, Google Cloud Secret Manager, 1Password, Keeper, environment variables, local files, and Base64 payloads) without writing application code.

### Why It Exists

* **Shell Scripting & Pipelines**: Extract secrets directly inside terminal sessions, shell scripts, and CI/CD pipelines.
* **Unified Interface**: Replace fragmented provider CLI tools (`aws`, `az`, `gcloud`, `vault`, `kubectl`, `op`, `ksm`) with single syntax.
* **Selective Provider Detection**: Automatically detects configured providers from the environment; you only need to supply credentials for the secret backends you actually use.
* **Pipeline Safe**: Writes raw secret values directly to `stdout` without trailing newlines, banners, or formatting artifacts. Logs and diagnostics route to `stderr`.
* **Pre-flight Checks**: Verify credential access (`creds`) or test secret coordinate existence (`exists`) before running workflows.

### Building on the Spelunk Library

The CLI binary bundles:

* **Core Engine**: `github.com/detro/spelunk/v2` coordinate parser and pipeline orchestrator.
* **Built-in Sources & Modifiers**: `plain://`, `file://`, `env://`, `base64://`, and `b64`/`b64e`/`b64d` modifiers.
* **Plugin Sources**: Decoupled plugins compiled together into single binary (`aws://`, `az://`, `gcp://`, `vault://`, `k8s://`, `op://`, `kp://`). The Bitwarden source (`bw://`) is not bundled: its SDK requires CGO and a platform-specific native library, which prevents shipping the CLI as a single cross-compiled static binary. It stays available to Go programs using the library.
* **Plugin Modifiers**: Full path extraction suite (`?jp=`, `?yp=`, `?tp=`, `?xp=`).
* **Auto-Configurators**: Automatic credential discovery from standard environment variables, configuration files (`~/.aws/credentials`, `~/.kube/config`, `~/.config/gcloud/...`), and CLI flags.

## Installation

### 1. Homebrew (macOS and Linux)

```shell
brew install detro/tap/spelunk
```

The cask covers macOS and Linux, on both `amd64` and `arm64`. It clears the macOS quarantine
attribute for you, and installs `bash`, `zsh` and `fish`
completions in the location each shell expects, so no extra setup is needed.

### 2. Pre-compiled Binaries (GitHub Releases)

Download pre-compiled binaries for Linux, macOS, and Windows directly from the [GitHub Releases](https://github.com/detro/spelunk/releases) page.

#### macOS Quarantine Notes

If downloaded via a web browser or GUI application, macOS Gatekeeper may block execution due to the quarantine attribute. You can clear this attribute using either method:

* **Command Line (`xattr`)**:

  ```shell
  xattr -d com.apple.quarantine ./spelunk
  chmod +x ./spelunk
  ```

* **System Settings**:

  Open **System Settings** -> **Privacy & Security**, scroll down to the **Security** section, and click **Allow Anyway** next to `spelunk`.

### 3. From Source (`go install`)

```shell
go install github.com/detro/spelunk/cmd/spelunk@latest
```

### 4. Build with Task

```shell
# From repository root
task --dir cmd/spelunk build

# Executable output location:
# cmd/spelunk/bin/spelunk
```

## Commands

### `dig` (Default)

Extracts secret value at given coordinates and writes raw value to `stdout`.

```shell
# Explicit subcommand
spelunk dig "k8s://production/app-secret/db-password"

# Default shorthand (subcommand can be omitted)
spelunk "k8s://production/app-secret/db-password"
```

### `exists`

Checks if secret exists and is accessible. Returns exit code `0` on success, non-zero on failure. Useful for health checks and conditional branching in scripts.

```shell
if spelunk exists "vault://secret/data/production/api-key"; then
  echo "Secret exists and is accessible"
fi
```

### `creds`

Scans environment and CLI flags, detects configured provider credentials, and verifies connectivity against each backend using non-mutating validation calls.

```shell
spelunk creds -v
```

### `completion`

Generates shell auto-completion scripts for `bash`, `zsh`, or `fish`.

```shell
# Bash
source <(spelunk completion bash)

# Zsh
source <(spelunk completion zsh)

# Fish
spelunk completion fish | source
```

## Coordinate Examples

### Built-in Backends

```shell
# Environment variable
spelunk "env://GITHUB_TOKEN"

# Local file with JSONPath extraction
spelunk "file:///etc/config.json?jp=$.database.password"

# Base64 string decoded
spelunk "base64://c3BlbHVuay1zZWNyZXQ=?b64d"

# Plain text with modifier chaining
spelunk "plain://payload?b64"
```

### Cloud & Secret Managers

```shell
# Kubernetes Secret (namespace / secret-name / key)
spelunk "k8s://prod/app-config/api-token"

# Kubernetes Secret (entire secret as JSON)
spelunk "k8s://prod/app-config/"

# HashiCorp Vault KV v2 (mount / secret-path / key)
spelunk "vault://secret/data/production/database/password"

# AWS Secrets Manager with JSONPath modifier
spelunk "aws://production/app/credentials?jp=$.password"

# Google Cloud Secret Manager
spelunk "gcp://projects/my-project/secrets/api-key/versions/latest"

# Azure Key Vault
spelunk "az://production-database-password"

# 1Password (Vault / Item / Field)
spelunk "op://Engineering/Database/password"

# Keeper Secrets Manager (Record UID / Field)
spelunk "kp://abcdef1234567890abcdef/password"
```

### Modifier Chaining

Modifiers apply sequentially from left to right:

```shell
# Extract JSON field, decode Base64, then extract YAML field
spelunk "aws://app/config?jp=$.encoded_payload&b64d&yp=$.auth.token"
```

## Practical Usage

### Command Substitution

```shell
# Inject secret into command argument
curl -H "Authorization: Bearer $(spelunk env://API_KEY)" https://api.example.com

# Pass secret to Docker login
spelunk "vault://secret/data/ci/docker/password" | docker login --username user --password-stdin
```

### Writing Secret to File

```shell
spelunk "k8s://prod/tls-certs/tls.crt" > /tmp/tls.crt
```

## Configuration Reference

The CLI selectively initializes providers by inspecting the environment. You only need to configure credentials for the secret backends you actually query; unconfigured providers are safely ignored.

Credentials are auto-detected from standard environment variables, configuration files, and CLI flags:

| Backend | CLI Flags | Environment Variables | Auto-discovery Files |
|---|---|---|---|
| **AWS** | `--aws-region`<br>`--aws-profile`<br>`--aws-endpoint-url` | `AWS_REGION`<br>`AWS_PROFILE`<br>`AWS_ACCESS_KEY_ID`<br>`AWS_SECRET_ACCESS_KEY`<br>`AWS_SESSION_TOKEN`<br>`AWS_ENDPOINT_URL_SECRETSMANAGER` | `~/.aws/credentials`<br>`~/.aws/config` |
| **Azure** | `--azure-vault-url`<br>`--azure-tenant-id`<br>`--azure-client-id`<br>`--azure-client-secret`<br>`--azure-insecure-skip-tls-verify` | `AZURE_KEYVAULT_URL`<br>`AZURE_TENANT_ID`<br>`AZURE_CLIENT_ID`<br>`AZURE_CLIENT_SECRET` | Default Azure CLI / Managed Identity credentials |
| **GCP** | `--gcp-credentials-file` | `GOOGLE_APPLICATION_CREDENTIALS`<br>`GOOGLE_APPLICATION_CREDENTIALS_JSON`<br>`SECRET_MANAGER_EMULATOR_HOST` | `~/.config/gcloud/application_default_credentials.json` |
| **Vault** | `--vault-addr`<br>`--vault-token`<br>`--vault-namespace` | `VAULT_ADDR`<br>`VAULT_TOKEN`<br>`VAULT_NAMESPACE` | `~/.vault-token` |
| **Kubernetes** | `--kubeconfig` | `KUBECONFIG` | In-cluster service account<br>`~/.kube/config` |
| **1Password** | `--op-service-account-token`<br>`--op-integration-name`<br>`--op-integration-version` | `OP_SERVICE_ACCOUNT_TOKEN` | - |
| **Keeper** | `--ksm-config` | `KSM_CONFIG` | Local file path or base64 config string |

### Logging Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--log.verbose` | `-v` | `0` | Increase log verbosity (repeatable: `-v`, `-vv`) |
| `--log.quiet` | `-q` | `0` | Decrease log verbosity (repeatable: `-q`, `-qq`) |
| `--log.format` | `-l` | `tinted` | Log format output (`tinted`, `json`, `boring`) |
