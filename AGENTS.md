# AGENTS.md

This file provides guidance to AI Agents like Claude Code (claude.ai/code)
when working with code in this repository.

## Project Overview

`tfc` is a CLI tool for interacting with Terraform Cloud/Enterprise and HCP
Terraform. It provides commands for managing workspaces, organizations, runs,
and variables.

## Development Commands

### Building

```bash
make build              # Build the binary to ./bin/tfc
```

### Testing

```bash
make test               # Run all tests with coverage
go test -v ./internal/tfc -run TestWorkspacesList  # Run specific test
go test -v ./cmd/tfc/workspace/list -run TestList_default  # Run specific subtest
```

### Linting and Formatting

```bash
make lint               # Run golangci-lint
make lint-go-fix        # Auto-fix linting issues
make tidy               # Format code and tidy modules
make format-go          # Format Go code only
```

## Architecture

### Command Structure

The CLI is built using cobra and follows this hierarchy:

- `cmd/tfc/main.go` - Entry point that initializes the factory and root
  command
- `cmd/tfc/cmd.go` - Root command (`NewCmdRoot`) that registers subcommands:
  - `workspace` - Workspace management commands
  - `organization` - Organization management commands
  - `run` - Run management commands
  - `version` - Version information

Each subcommand follows the pattern: `cmd/tfc/<resource>/<resource>.go` with
nested commands in subdirectories (e.g., `cmd/tfc/workspace/list/list.go`).

### Factory Pattern

The `Factory` (pkg/cmdutil/factory.go) provides dependency injection for:

- `TFEClient` - Lazy initialization of the Terraform Cloud API client
- `IOStreams` - I/O streams for terminal interaction
- `Clock` - Time utilities for testing
- `ExecutableName` and `AppVersion` - Build metadata

### Client Wrapper

The `internal/tfc` package wraps the official `hashicorp/go-tfe` client:

- `tfc.Client` embeds `tfe.Client` and adds custom service methods
- Services like `WorkspacesService` and `OrganizationsService` extend the base
  TFE client with additional functionality
- Custom pagination logic in `internal/tfc/tfepaging/pager.go` uses Go 1.23
  iterators to handle paginated API responses

### Testing Utilities

The `internal/tfc/tfetest` package provides test helpers:
- `Setup()` creates a test HTTP server with mux and client
- Supports middleware for request logging and custom behavior
- Returns a cleanup function to close the test server

### Configuration

The application uses environment variables for configuration:

- `TFE_TOKEN` - Terraform Enterprise API token (required)
- `TFE_ADDRESS` - Custom TFE address (default: https://\$TFE_HOSTNAME)
- `TFE_HOSTNAME` - TFE hostname (default: app.terraform.io)

### Key Packages

- `pkg/cmdutil` - Command utilities and factory
- `pkg/iolib` - I/O streams wrapper
- `pkg/table` - Table formatting for CLI output
- `pkg/term` - Terminal color and formatting
- `pkg/text` - Text utilities (heredoc support)
- `pkg/ptr` - Pointer utilities
- `pkg/credentials` - TFE token management
- `internal/build` - Build version information injected via ldflags

## Implementation Patterns

### Adding a New Command

1. Create command directory under `cmd/tfc/<resource>/<action>/`
2. Implement `NewCmd<Action>()` function that takes `*cmdutil.Factory`
3. Use factory to get TFE client: `client, err := f.TFEClient()`
4. Register command in parent command's `NewCmd<Resource>()` function

### Adding Service Methods

1. Add method to appropriate service in `internal/tfc/` (e.g., `workspace.go`)
2. If wrapping TFE client directly, delegate to `s.tfe.<Service>.<Method>()`
3. For custom logic, use `internal/tfc/tfepaging.New()` for paginated
   endpoints
4. Return custom types or wrapped TFE types as needed

### Writing Tests

1. Use `internal/tfc/tfetest.Setup()` to create test client and mux
2. Register handlers on the mux for endpoints being tested
3. For command tests, use `internal/test` helpers for streams and assertions
4. Test files are co-located with implementation files (e.g.,
   `workspace_test.go`)
