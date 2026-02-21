# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-02-21

### Added

- **Plugins**:
    - `vault://`: HashiCorp Vault KV Secret source implementation (available in `plugin/source/vault`). Supports both KV v1 and v2 engines.
- **Features**:
    - Both `k8s://` and `vault://` plugins now support returning the entire secret data map as a JSON object when the URI path ends with a `/` instead of a specific key.
- **Documentation**:
    - Expanded `AGENTS.md` and `ARCHITECTURE.md` to cover new plugins, examples, and testing instructions.
    - Updated README with additional vanity badges and future features list.

### Changed

- **CI**: Inverted order of testing and linting, and excluded markdown changes from CI triggers.

## [1.0.0] - 2026-02-16

### Added

- **Core**: Initial release of `spelunk`, a Go library for unified secret retrieval.
- **Coordinates**: Support for URI-based secret coordinates (`scheme://location?modifier=arg`).
- **Spelunker**: Main client implementation with configurable options.
- **Built-in Sources**:
    - `env://`: Retrieve secrets from environment variables.
    - `file://`: Retrieve secrets from local files.
    - `plain://`: Use plain text strings as secrets (useful for testing).
    - `base64://`: Decode base64 strings as secrets.
- **Built-in Modifiers**:
    - `?jp=`: Extract values from JSON content using JSONPath syntax.
- **Plugins**:
    - `k8s://`: Kubernetes Secret source implementation (available in `plugin/source/kubernetes`).
- **Extensibility**: Public interfaces `SecretSource` and `SecretModifier` for custom implementations.
- **Tooling**: Comprehensive toolchain managed via [asdf](https://asdf-vm.com/) and [Task](https://taskfile.dev/).
  Includes `Taskfile.yaml` for build, test, lint, and documentation tasks.
- **Examples**: Integration examples with popular libraries:
    - [Kong](https://github.com/alecthomas/kong)
    - [Viper](https://github.com/spf13/viper)
    - [Urfave CLI](https://github.com/urfave/cli)
    - Standard library `flag` package
- **Automation**:
    - **CI**: GitHub Actions workflow (`.github/workflows/ci.yaml`) for automated build,
      test (with coverage), lint, and format checks using `task`.
    - **Dependabot**: Automated dependency updates for Go modules (weekly) and GitHub Actions (monthly).
- **Documentation**: Added `README.md`, `ARCHITECTURE.md`, `AGENTS.md`, and `CONTRIBUTING.md`.
