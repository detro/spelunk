# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Spelunk CLI Documentation & Architecture**: Added dedicated `README.md` and `ARCHITECTURE.md` documentation for `cmd/spelunk` detailing CLI capabilities, subcommands, provider auto-configuration, shell completion, and architectural design, along with root documentation links.
- **Spelunk CLI Utility**: Added a command-line interface in `cmd/spelunk` to dig up secrets directly from the terminal.
  - Subcommands: `dig` (default) to retrieve secrets, `exist` (alias: `is`) to verify secret existence without outputting values, and `creds` (alias: `check`) to validate backend provider credentials.
  - Built-in Support: Integrated all secret sources (AWS, Azure, GCP, Vault, Kubernetes, 1Password, Bitwarden, Keeper, file, env, plain, base64) and modifiers (JSONPath, TOMLPath, XPath, YAMLPath, base64).
  - Configurable Logging: Structured logging with support for colored console output, text, and JSON formats.
  - End-to-End Testing: Comprehensive integration test suites using Testcontainers for AWS, Azure, GCP, Vault, and Kubernetes.
- **CLI Automated Release & Signing Pipeline**: Added GoReleaser configuration (`.goreleaser.yaml`) and GitHub Actions release workflow triggering on `cmd/spelunk/v*` tags with cryptographic artifact signing via Sigstore Cosign (keyless) and GPG checksum signing.
- **Coordinates Reconstitution**: Implemented `fmt.Stringer` (`String()` method) on `types.SecretCoord` to reconstruct the URI representation from parsed secret coordinates.
- **Secret Source Coordinate Test Coverage**: Expanded test suites across all built-in and plugin `SecretSource` implementations to systematically exercise supported coordinate format variations (including leading/trailing slashes, encoded paths, version specifications, and offline parsing validations).

### Fixed

- **Kubernetes Source**: Fixed URI parsing for `k8s://NAME/` to correctly return the entire secret data map as JSON.
- **Task Tagging Propagation**: Fixed root `task tag` command to properly delegate tagging to `cmd/spelunk` child task via dependencies instead of searching it in directory traversal.

### Changed

- **Task Runner Output**: Silenced command echoing (`silent: true`) for verbose tasks (`lint`, `vuln`, `fmt`, `mod.*`, `tag`, `tools.*`) across root and CLI taskfiles.
- **CI Workflows**: Decoupled CLI release pipeline into a dedicated GitHub Actions workflow (`.github/workflows/release-cli.yaml`), separating it from continuous integration (`.github/workflows/ci.yaml`).
- **Dependencies**: Bumped Go toolchain to `1.26.6` and updated Go module dependencies across the workspace.

## [2.0.0] - 2026-05-29

### Added

- **Multi-Module Workspace Support**: Restructured the entire repository into a multi-module architecture leveraging [Go Workspaces](https://go.dev/doc/tutorial/workspaces).
- **Submodule Isolation**: Converted all 12 plugins (`plugin/modifier/*` and `plugin/source/*`) and 4 example applications (`examples/*`) into fully decoupled, isolated Go modules.
- **Root-level Public Utilities**: Created the root public package `github.com/detro/spelunk/v2/util` containing shared utilities (`post_process_jsonpath.go`, `mock_source.go`, and `mock_modifier.go`) to prevent import cycles and make testing helpers cleanly importable across standalone submodules.
- **Unified Tagging Tool**: Added a robust `task tag` command to `Taskfile.yaml` that automates tagging either the entire workspace at once (root + all submodules using their relative directory prefixes) or target submodules individually.

### Changed

- **Dependencies Separation (Ultra-lean Core)**: The core root module `github.com/detro/spelunk/v2` has been stripped down to a absolute minimum dependency surface (carrying almost zero external production dependencies). Users now only pull down the specific heavyweight SDK dependencies (e.g. AWS, Azure, GCP, Vault, Kubernetes) for the exact plugins they choose to import.
- **Plugin Module Import Paths**: All 12 plugin imports have been updated to target their isolated `v2` module paths (e.g. `github.com/detro/spelunk/plugin/source/vault/v2`).
- **Task Runner Optimization**: Enhanced and parallelized `Taskfile.yaml` commands (`build`, `test`, `lint`, `fmt`, `vuln`) to recursively cycle through the root module, all 12 plugin modules, and all 4 examples, leveraging workspace-aware Go test targets and concurrent execution via `xargs` to significantly speed up feedback loops.
- **Azure SDK Upgrade**: Upgraded `github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets` to `v1.5.0` in the Azure Key Vault plugin.

### Fixed

- **Azure Emulator Testing**: Configured Azure Key Vault integration tests to target API version `7.4` to maintain compatibility with `lowkey-vault` emulator, following the `azsecrets` SDK upgrade to `v1.5.0`.
- **Examples Build Protection**: Updated task runner configuration and workspace settings to prevent compiled binaries of `/examples` from being accidentally checked into Git.
- **Workspace Tagging Scope**: Refined tagging automation to ensure examples are excluded from automated plugin submodule tagging tasks.

## [1.4.0] - 2026-05-23

### Added

- **Plugins**:
    - `op://`: 1Password source implementation (available in `plugin/source/1password`).
    - `bw://`: Bitwarden Secrets Manager source implementation (available in `plugin/source/bitwarden`).
      - WARNING: currently untested due to lack of test environment.
    - `kp://`: Keeper Secrets Manager source implementation (available in `plugin/source/keeper`).
      - WARNING: currently untested due to lack of test environment.
    - `?jp=`: JSONPath extractor modifier for JSON secrets (available in `plugin/modifier/jsonpath`).
- **Tooling**:
    - Test tasks in `Taskfile.yaml` (`test`, `test.full`, `test.short`, `test.ci`) now support passing a specific directory path using `-- <path>`.
    - Added modular `tools.plugins`, `tools.update`, and `tools.install` tasks to `Taskfile.yaml` for robust `asdf`-based toolchain management.
    - Integrated `govulncheck` (v1.3.0) into `.tool-versions` toolchain and added `task vuln` for local vulnerability scanning.
    - Integrated `task vuln` check directly into the CI pipeline.

### Changed

- **Refactoring**: Unified `InvalidLocation` errors across all plugins by introducing a global `types.ErrInvalidLocation`, replacing plugin-specific errors (e.g. `ErrSecretSourceAWSInvalidLocation`, `ErrSecretSourceVaultInvalidLocation`, etc.) to simplify error handling for consumers.
- **Dependencies**: Bumped `task`, `golang`, `golangci-lint` and various Go module dependencies.
- **Support**: Documented in [README](./README.md) that for now we are not going to support LastPass (`lp://`)
  nor Dashlane (`dl://`) as a source. They both lack a Golang SDK and/or a REST API.

### Removed

- **BREAKING CHANGE**: Removed `jp` (JSONPath) modifier from default built-in modifiers of `Spelunker` to completely free the core root module from any external production dependencies. It has been moved to a plugin under `plugin/modifier/jsonpath/` and must now be explicitly registered using `jsonpath.WithJSONPath()`.


## [1.3.2] - 2026-04-07

### Changed

- **Dependencies**: Bumped `github.com/go-jose/go-jose/v4` to `4.1.4` and other dependencies.

## [1.3.1] - 2026-03-19

### Fixed

- **Security**: Addressed `CVE-2026-33186` - see [advisory](https://github.com/advisories/GHSA-p77j-4mvh-x3m3).

### Changed

- **Dependencies**: Bumped toolchain dependencies.

## [1.3.0] - 2026-03-16

### Added

- **Plugins**:
    - `?xp=`: XPath extractor modifier for XML secrets (available in `plugin/modifier/xpath`).
    - `?yp=`: YAML JSONPath extractor modifier for YAML secrets (available in `plugin/modifier/yamlpath`).
    - `?tp=`: TOML JSONPath extractor modifier for TOML secrets (available in `plugin/modifier/tomlpath`).

### Changed

- **Refactoring**: Extracted JSONPath post-processing and test source mocking to internal utilities (`internal/jsonpathutil` and `internal/testutil`) to facilitate code reuse across extractors.
- **Errors Improvement**: All `jsonpath`-based modifiers now compile the JSONPath expression _before_ querying to separate syntax errors from matching errors.

## [1.2.0] - 2026-03-13

### Added

- **Plugins**:
    - `aws://`: AWS Secrets Manager source implementation (available in `plugin/source/aws`).
    - `gcp://`: Google Cloud Secret Manager source implementation (available in `plugin/source/gcp`).
    - `az://`: Azure Key Vault source implementation (available in `plugin/source/azure`).
- **Built-in Modifiers**:
    - `?b64d`: Decode base64 strings back to their original secret value.
        Useful to decode binary value returned by Sources like `aws://` and `gcp://`.
    - `?b64` and `?b64e`: Encode secret value to a base64 string.
- **Documentation**:
    - Added direct links to the documentation for each built-in Secret Source and Secret Modifier in the README.
    - Explicitly documented built-in vs plugin architecture.
    - Updated `AGENTS.md` with extra safety measures and AI instructions.

### Changed

- **CI**: Restricted permissions of the auto-generated GITHUB_TOKEN in GitHub Actions.
- **Testing**: Refactored Testcontainers spawning and secret creation utilities across tests.

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
