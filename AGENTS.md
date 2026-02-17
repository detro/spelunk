# Spelunk Agent Guide

This document provides essential information for AI agents working on the `spelunk` codebase.

## Project Overview

`spelunk` is a Go library designed to extract secrets from various storage backends (Kubernetes, local files, environment variables, etc.).

## 🛠 Toolchain

The project uses the following tools for development and maintenance:

- **Go**: Language (v1.26+)
- **Task**: Task runner (replaces Makefiles)
- **GolangCI-Lint**: Linter (configured with `gci`, `gofmt`, `gofumpt`, `goimports`, `golines`, `swaggo`)
- **Glow**: Markdown renderer (for README)
- **ASDF**: Version manager (see `.tool-versions`)
- **Pkgsite**: For local documentation preview

## ⚡️ Common Commands

All development tasks are defined in `Taskfile.yaml`. **Always use `task` instead of direct `go` commands when possible.**

| Command | Description |
|---------|-------------|
| `task build` | Build the project |
| `task test` | Run all tests (alias for `test.full`) |
| `task test.full` | Run all tests with race detection and coverage |
| `task test.short` | Run short tests (skips integration tests) |
| `task test.ci` | Run tests with coverage profile for CI |
| `task lint` | Run linter |
| `task lint-fix` | Run linter and automatically fix issues |
| `task fmt` | Format code |
| `task run -- <args>` | Run the project (passes args to `go run .`) |
| `task update-dependencies` | Update and tidy Go modules |
| `task readme` | Render README.md in terminal |
| `task docs` | Run local godoc server (`pkgsite`) and open in browser |

## ⚙️ CI/CD

The project's CI pipeline (`.github/workflows/ci.yaml`) is driven entirely by `task` commands. This ensures that what you run locally matches what runs in CI.

- **Build**: `task build`
- **Test**: `task test.ci`
- **Lint & Format**: `task lint-fix` && `task fmt`

## 📂 Project Structure

- **`spelunker.go`**: Core logic for the `Spelunker` struct and `DigUp` method.
- **`types/`**: Core types and interfaces.
    - **`coordinates.go`**: `SecretCoord` type and parsing logic.
    - **`source.go`**: `SecretSource` interface definition.
    - **`modifier.go`**: `SecretModifier` interface definition.
    - **`errors.go`**: Common error definitions (`ErrSecretNotFound`, etc.).
- **`builtin/`**: Built-in implementations.
    - **`source/plain/`**: `plain://` source implementation.
    - **`source/file/`**: `file://` source implementation.
    - **`source/env/`**: `env://` source implementation.
    - **`source/base64/`**: `base64://` source implementation.
    - **`modifier/jsonpath/`**: `jp` modifier implementation (JSONPath extraction).
- **`plugin/`**: External plugins (opt-in).
    - **`source/kubernetes/`**: `k8s://` source implementation (integration tested with Testcontainers).
- **`examples/`**: Example implementations (e.g., `kong/`, `viper/`, `basic/`).
- **`docs/`**: Documentation assets (images, logos).
- **`options.go`**: Functional options for configuring `Spelunker`.
- **`doc.go`**: Package-level documentation.
- **`Taskfile.yaml`**: Task definitions.
- **`.tool-versions`**: ASDF tool versions.

## 🧪 Testing & Quality

- **Framework**: Use [testify](https://github.com/stretchr/testify) (`require`, `assert`) for all tests.
- **Scope**: Use `spelunk_test` package for black-box testing (e.g., `spelunker_test.go`, `types/coordinates_test.go`).
- **Integration Tests**: Use [Testcontainers for Go](https://golang.testcontainers.org/) for integration tests (e.g., Kubernetes). Use `testcontainers.CleanupContainer` for resource management.
- **Execution**: Run tests with `task test`.
- **Linting**: Ensure code passes `task lint` before finishing.
- **Test Data**: Use `testdata/` directories for file-based tests.

## 📝 Conventions & Style

- **Naming**: Follow standard Go conventions (CamelCase, short variable names where appropriate).
- **Error Handling**: Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **Imports**: Group standard library imports separately from third-party imports.
- **Interfaces**: Define interfaces where they are used (consumer-side), but `SecretSource` and `SecretModifier` are defined in `types/` for clarity and reuse.
- **Type Safety**: Ensure types implement expected interfaces with compile-time checks (e.g., `var _ types.SecretSource = (*SecretSourceFile)(nil)`).
- **JSON Marshaling**: `SecretCoord` implements `encoding.TextUnmarshaler`, allowing direct use in `json.Unmarshal`.
- **Markdown**:
  - **Heading**: always have an empty line before and after each heading
  - **Lists**: Always use only one space between the number/dot and the text
  - **Code and quote blocks**: Always place an empty line before and after each code or quote block

## 🤖 AI Contribution Guidelines

Based on `CONTRIBUTING.md`:
1. **Small and Steady**: Keep changes small and focused.
2. **Sole Responsibility**: You are responsible for every line of code.
3. **Know Your Code**: Understand every design decision.

## ⚠️ Gotchas

- **Modifiers Application Order**: `SecretCoord.Modifiers` is a slice of key-value pairs (`[][2]string`). Modifiers are applied in the exact order they appear in the connection string URI. Duplicate keys are allowed and preserved.
- **JSONPath Behavior**:
    - The `jp` modifier returns the **first element** if the JSONPath matches a list.
    - **Floats**: Converted to string without scientific notation and with minimal necessary precision (e.g. `1.50000` -> `1.5`).
    - **Nulls**: Returns an error if the result is explicitly `null`.
- **Whitespace Trimming**: By default, `Spelunker` trims whitespace from retrieved secrets *after* applying modifiers. This can be disabled via `WithoutTrimValue()` option.
- **SecretCoord Parsing**: 
    - `SecretCoord.Location` includes both Authority (userinfo/host) and Path.
    - If a URI contains userinfo (e.g., `plain://user:pass@host`), it is correctly preserved in `Location`.
- **Error Types**: `SecretCoord` parsing returns specific errors (e.g., `ErrSecretCoordHaveNoType`). Tests should check for these using `errors.Is`.
- **Dependencies**: Ensure you have the correct Go version (1.26) as specified in `go.mod`.
