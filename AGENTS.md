# Spelunk Agent Guide

This document provides essential information for AI agents working on the `spelunk` codebase.

## Project Overview

`spelunk` is a Go library designed to extract secrets from various storage backends (Kubernetes, local files, environment variables, etc.).

## 🛠 Toolchain

The project uses the following tools for development and maintenance:

- **Go**: Language (v1.26+)
- **Task**: Task runner (replaces Makefiles)
- **GolangCI-Lint**: Linter
- **Glow**: Markdown renderer (for README)
- **ASDF**: Version manager (see `.tool-versions`)
- **Pkgsite**: For local documentation preview

## ⚡️ Common Commands

All development tasks are defined in `Taskfile.yaml`. **Always use `task` instead of direct `go` commands when possible.**

| Command | Description |
|---------|-------------|
| `task build` | Build the project |
| `task test` | Run all tests with race detection and coverage |
| `task lint` | Run linter |
| `task lint-fix` | Run linter and automatically fix issues |
| `task fmt` | Format code |
| `task run -- <args>` | Run the project (passes args to `go run .`) |
| `task update-dependencies` | Update and tidy Go modules |
| `task readme` | Render README.md in terminal |
| `task docs` | Run local godoc server (`pkgsite`) and open in browser |

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
    - **`kubernetes/`**: `k8s://` source implementation (integration tested with Testcontainers).
- **`options.go`**: Functional options for configuring `Spelunker`.
- **`doc.go`**: Package-level documentation.
- **`pkg/`**: Public library code.
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

## ⚠️ Gotchas

- **Whitespace Trimming**: By default, `Spelunker` trims whitespace from retrieved secrets. This can be disabled via `WithoutTrimValue()` option.
- **SecretCoord Parsing**: 
    - `SecretCoord.Location` includes both Authority (userinfo/host) and Path.
    - If a URI contains userinfo (e.g., `plain://user:pass@host`), it is correctly preserved in `Location`.
- **Error Types**: `SecretCoord` parsing returns specific errors (e.g., `ErrSecretCoordHaveNoType`). Tests should check for these using `errors.Is`.
- **Modifiers**: 
    - Support for URI modifiers (e.g., `?jp=...`) allows processing secrets after retrieval.
    - Currently supports JSONPath via `jp` modifier (e.g., `k8s://NAMESPACE/NAME/KEY?jp=$.kafka.brokers`).
- **Dependencies**: Ensure you have the correct Go version (1.26) as specified in `go.mod`.
