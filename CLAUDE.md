# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider for the [Clerk](https://clerk.com) authentication platform. Written in Go using the Terraform Plugin Framework. Enables managing Clerk resources (organizations, users, roles, permissions, etc.) via Terraform/OpenTofu.

## Repository Status

This project is in early development. The initial scaffolding is being built out.

## Build & Development Commands

```bash
# Build the provider
go build -o terraform-provider-clerk

# Run all tests
go test ./...

# Run acceptance tests (requires CLERK_PLATFORM_API_KEY)
TF_ACC=1 go test ./... -v

# Run a single test
go test ./internal/provider -run TestAccClerkOrganization_basic -v

# Run tests with acceptance flag for a specific package
TF_ACC=1 go test ./internal/provider -run TestAcc -v -timeout 120m

# Lint
golangci-lint run ./...

# Generate documentation
go generate ./...

# Install provider locally for manual testing
go install .
```

## Architecture

This provider follows the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) pattern (not the older SDKv2).

### Expected Structure

```text
├── main.go                     # Entry point, serves the provider
├── internal/
│   ├── provider/               # Provider configuration and registration
│   │   ├── provider.go         # Provider schema, Configure(), Resources(), DataSources()
│   │   └── provider_test.go
│   ├── resources/              # Terraform resources (CRUD)
│   │   ├── organization.go
│   │   └── ...
│   ├── datasources/            # Terraform data sources (read-only)
│   └── client/                 # Clerk API client wrapper
├── examples/                   # Example Terraform configurations for docs
│   ├── provider/
│   ├── resources/
│   └── data-sources/
└── docs/                       # Generated documentation (via tfplugindocs)
```

### Key Patterns

- **Provider configuration**: API key passed via `platform_api_key` attribute or `CLERK_PLATFORM_API_KEY` environment variable
- **Resources**: Each resource implements `resource.Resource` interface with CRUD methods (Create, Read, Update, Delete)
- **Data sources**: Each data source implements `datasource.DataSource` interface with Read method
- **Client**: Thin wrapper around the Clerk Go SDK, instantiated in `provider.Configure()`
- **Testing**: Acceptance tests use `resource.Test()` with real API calls against Clerk (gated by `TF_ACC=1`)

### Naming Conventions

- Resource type names: `clerk_<resource>` (e.g., `clerk_organization`, `clerk_user`)
- Go files: snake_case matching the resource name
- Test files: `<resource>_test.go` with functions named `TestAcc<Resource>_<scenario>`

## Infrastructure Notes

- Spacelift manages all OpenTofu/Terraform deployments — never apply changes directly
- Push changes via PR, not directly to main
- CI runs via GitHub Actions
