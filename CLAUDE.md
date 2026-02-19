# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider for the [Clerk](https://clerk.com) authentication platform. Written in Go using the Terraform Plugin Framework. Enables managing Clerk applications and instance settings via Terraform/OpenTofu.

## Repository Status

Active development. Phase 1 (scaffolding) and Phase 2 (application lifecycle) are complete. Phase 3 (environment configuration) is in progress.

### Implemented Resources
- `clerk_application` — Full CRUD with deletion protection, dev/prod instance key exposure
- `clerk_environment` — Instance settings, restrictions, organization settings per environment

### Implemented Data Sources
- `clerk_application` — Read-only lookup by ID

## Build & Development Commands

```bash
# Build the provider
go build -o terraform-provider-clerk

# Run all tests
go test ./...

# Run acceptance tests (requires CLERK_PLATFORM_API_KEY)
TF_ACC=1 go test ./... -v -timeout 120m

# Run a single test
TF_ACC=1 go test ./internal/provider -run TestAccClerkEnvironment_basic -v

# Lint
golangci-lint run ./...

# Install provider locally for manual testing
make install
```

## Architecture

This provider follows the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) pattern (not the older SDKv2).

### Two-Tier API Pattern

- **Platform API** (workspace-level): Manages applications. Uses `CLERK_PLATFORM_API_KEY`.
- **Backend API** (per-instance): Manages instance settings. Secret keys are resolved via internal key routing — `clerk_application` registers keys on create/read, and `clerk_environment` looks them up by `{app_id}/{environment}`.

### Structure

```text
internal/
  client/
    client.go          # ClerkClient with backend client registry (thread-safe)
    platform_api.go    # Platform API HTTP client (application CRUD)
    backend_api.go     # Backend API wrapper using Clerk SDK (instancesettings)
  provider/
    provider.go        # Provider schema, Configure(), resource/datasource registration
  resources/
    application.go     # clerk_application (CRUD + deletion protection + key routing)
    environment.go     # clerk_environment (settings-only, no read API, composite ID)
  datasources/
    application.go     # data.clerk_application (read-only by ID)
```

### Key Patterns

- **Provider configuration**: `platform_api_key` attribute or `CLERK_PLATFORM_API_KEY` env var
- **Key routing**: `ClerkClient.RegisterBackendClient()` / `GetBackendConfig()` maps `{app_id}/{env}` to Clerk SDK client configs
- **Deletion protection**: Provider-side boolean (defaults to true) that blocks `Delete`. Must handle `types.Bool` null/unknown states.
- **Settings-only resources**: `clerk_environment` doesn't create/delete the instance — it configures it. Delete resets to defaults. Read preserves state (no drift detection).
- **Composite IDs**: `clerk_environment` uses `{application_id}/{environment}` as its ID

### Naming Conventions

- Resource type names: `clerk_<resource>` (e.g., `clerk_application`, `clerk_environment`)
- Go files: snake_case matching the resource name
- Test files: `<resource>_test.go` with functions named `TestAcc<Resource>_<scenario>`
- Acceptance tests in `internal/provider/`, unit tests alongside the code they test

### Clerk API Limitations

- Authentication strategies (email/password/OAuth/MFA) are dashboard-only — no API
- Backend API `instancesettings` has no GET endpoints — only PATCH (update)
- Platform API doesn't return application `name` in GET responses
