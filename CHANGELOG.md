# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **Windows CLI Startup**: The Windows executable was linked against the Bitwarden SDK's DLL import stub rather than its static library, so it aborted immediately on launch with `STATUS_DLL_NOT_FOUND`, looking for a `bitwarden_c.dll` that is not distributed anywhere. The Bitwarden code is now linked statically, and a release is aborted before publishing if a Windows binary ever imports that DLL again.

### Added

- **Release Archives**: CLI binaries are now shipped as archives (`.tar.gz`, `.zip` on Windows) that bundle the `README`, `CHANGELOG`, and `LICENSE`.
- **SBOMs**: Software Bill of Materials is generated and published for each release archive (`syft` is now installed in the release workflow).
- **Release Metadata**: Added a `metadata` section to the release configuration, pinning artifact timestamps to the commit date for reproducible builds.
- **Homebrew (Disabled)**: Added a commented-out Homebrew tap release section, ready to be enabled.

### Changed

- **Artifact Signing**: Signing now covers the checksums file only (archives are covered transitively), with signing output surfaced in release logs.
- **Releases**: Dropped pre-release handling; all releases are treated as regular releases.

## [2.1.0] - 2026-08-18

### Added

- **Spelunk CLI Utility**: New CLI (`cmd/spelunk`) to retrieve secrets directly from the terminal.
  - Subcommands: `dig` (retrieve secret), `exists` (alias `is`, verify existence), and `creds` (alias `check`, validate provider credentials).
  - Provider Support: Pre-configured with all built-in sources, plugin sources (AWS, Azure, GCP, Vault, Kubernetes, 1Password, Bitwarden, Keeper), and modifiers (`jp`, `yp`, `tp`, `xp`, `b64`).
  - Logging: Structured logging with colored console, text, and JSON formats.
  - Documentation: Dedicated `README.md` and `ARCHITECTURE.md` for CLI configuration, shell completions, and design.
- **CLI Release Pipeline**: Automated multi-platform releases (`.github/workflows/release-cli.yaml`) via GoReleaser across Linux, macOS, and Windows with Cosign bundle signing and GPG signatures.
- **Coordinate Formatting**: Added `fmt.Stringer` (`String()`) to `types.SecretCoord` to reconstruct URI strings from parsed coordinates.
- **Test Coverage**: Added test suites across all sources covering URI syntax variations (slashes, URL encoding, version tags).

### Fixed

- **Kubernetes Source**: Fixed `k8s://NAME/` URI parsing to return full secret data map as JSON.
- **Task Tagging**: Fixed root `task tag` to propagate cleanly to `cmd/spelunk`.

### Changed

- **Task Runner Output**: Silenced command echoing (`silent: true`) on verbose tasks (`lint`, `vuln`, `fmt`, `tag`, `tools.*`).
- **CI Workflows**: Separated CLI releases (`release-cli.yaml`) from continuous integration (`ci.yaml`).
- **Dependencies**: Updated Go toolchain to `1.26.6` and bumped module dependencies across workspace.

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
