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

## 📂 Project Structure

- **`spelunker.go`**: Core logic for the `Spelunker` struct and `DigUp` method.
- **`coordinates.go`**: `SecretCoord` type and parsing logic.
- **`source.go`**: `SecretSource` interface definition.
- **`options.go`**: Functional options for configuring `Spelunker`.
- **`doc.go`**: Package-level documentation.
- **`internal/`**: Internal implementation details.
- **`pkg/`**: Public library code.
- **`Taskfile.yaml`**: Task definitions.
- **`.tool-versions`**: ASDF tool versions.

## 🧪 Testing & Quality

- **Tests**: Use `spelunk_test` package for black-box testing.
- Run tests with `task test`.
- Ensure code passes `task lint` before finishing.

## 📝 Conventions & Style

- **Naming**: Follow standard Go conventions (CamelCase, short variable names where appropriate).
- **Error Handling**: Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **Imports**: Group standard library imports separately from third-party imports.
- **Interfaces**: Define interfaces where they are used (consumer-side), but `SecretSource` is defined in `source.go` for clarity.

## ⚠️ Gotchas

- **Whitespace Trimming**: By default, `Spelunker` trims whitespace from retrieved secrets. This can be disabled via `WithoutTrimValue()` option.
- **Error Types**: `SecretCoord` parsing returns specific errors (e.g., `ErrSecretCoordHaveNoType`). Tests should check for these using `errors.Is`.
- **Modifiers**: Support for URI modifiers (e.g., `?jsonpath=...`) is planned but implementation is currently partial (TODO in `DigUp`).
- **Dependencies**: Ensure you have the correct Go version (1.26) as specified in `go.mod`.
