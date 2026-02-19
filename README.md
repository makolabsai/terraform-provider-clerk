# Terraform Provider for Clerk

A Terraform/OpenTofu provider for managing [Clerk](https://clerk.com) authentication platform resources as infrastructure-as-code.

## Overview

This provider uses the [Clerk Platform API](https://clerk.com/docs/reference/backend-api) (workspace-level) and [Clerk Backend API](https://clerk.com/docs/reference/backend-api) (per-instance) to manage Clerk resources. It takes a single Platform API key and internally routes Backend API calls to the correct instance using registered secret keys.

### Supported Resources

| Resource | Description |
|----------|-------------|
| `clerk_application` | Manages Clerk applications (create, update, delete) with dev/prod instances |
| `clerk_environment` | Configures instance settings, restrictions, and organization settings per environment |

### Supported Data Sources

| Data Source | Description |
|-------------|-------------|
| `clerk_application` | Looks up an existing Clerk application by ID |

> **Note:** Authentication strategies (email/password/OAuth/MFA) are only configurable via the [Clerk Dashboard](https://clerk.com/docs/guides/configure/auth-strategies/sign-up-sign-in-options), not through this provider.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 (or [OpenTofu](https://opentofu.org/) >= 1.6)
- [Go](https://golang.org/doc/install) >= 1.22 (for building from source)
- A [Clerk](https://clerk.com) account with a Platform API key

## Installation

### From Source

```bash
git clone https://github.com/makolabsai/terraform-provider-clerk.git
cd terraform-provider-clerk
make install
```

This installs the provider to `~/.terraform.d/plugins/` for local use.

## Authentication

The provider requires a Clerk Platform API key for workspace-level operations. You can provide it in two ways:

### Environment Variable (recommended)

```bash
export CLERK_PLATFORM_API_KEY="your-platform-api-key"
```

### Provider Configuration

```hcl
provider "clerk" {
  platform_api_key = var.clerk_platform_api_key
}
```

## Usage

### Create an Application

```hcl
resource "clerk_application" "payments" {
  name                = "Payments App"
  deletion_protection = true
}
```

The application resource exposes instance keys for both development and production environments:

- `dev_instance_id`, `dev_secret_key`, `dev_publishable_key`
- `prod_instance_id`, `prod_secret_key`, `prod_publishable_key`

### Configure an Environment

```hcl
resource "clerk_environment" "payments_dev" {
  application_id = clerk_application.payments.id
  environment    = "development"

  hibp          = true
  support_email = "support@example.com"

  restrictions = {
    block_disposable_email_domains = true
    block_email_subaddresses       = true
  }

  organization_settings = {
    enabled                 = true
    max_allowed_memberships = 10
    admin_delete_enabled    = true
  }
}
```

### Look Up an Existing Application

```hcl
data "clerk_application" "existing" {
  id = "app_abc123"
}

output "dev_publishable_key" {
  value = data.clerk_application.existing.dev_publishable_key
}
```

### Full Example

See [`examples/`](./examples/) for complete working configurations.

## Resource Reference

### clerk_application

Manages a Clerk application. Each application can have multiple instances (development, production) with distinct user pools.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | The name of the application |
| `deletion_protection` | bool | no | Prevents accidental deletion. Defaults to `true`. Set to `false` before destroying. |
| `domain` | string | no | Domain for the application (create-only) |
| `template` | string | no | Application template, e.g. `b2b-saas` (create-only) |
| `environment_types` | list(string) | no | Environment types to create (create-only) |

**Computed attributes:** `id`, `dev_instance_id`, `dev_secret_key`, `dev_publishable_key`, `prod_instance_id`, `prod_secret_key`, `prod_publishable_key`

### clerk_environment

Configures a Clerk instance's settings. The instance is auto-created by Clerk; this resource manages its configuration.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `application_id` | string | yes | The Clerk application ID |
| `environment` | string | yes | `"development"` or `"production"` |
| `test_mode` | bool | no | Whether test mode is enabled |
| `hibp` | bool | no | Have I Been Pwned password checking |
| `enhanced_email_deliverability` | bool | no | Clerk-managed email deliverability |
| `support_email` | string | no | Support contact email |
| `url_based_session_syncing` | bool | no | URL-based session syncing (dev) |
| `development_origin` | string | no | Dev origin URL |
| `restrictions` | object | no | Email restriction settings (see below) |
| `organization_settings` | object | no | Organization feature settings (see below) |

**Restrictions block:** `allowlist`, `blocklist`, `block_email_subaddresses`, `block_disposable_email_domains`, `ignore_dots_for_gmail_addresses`

**Organization settings block:** `enabled`, `max_allowed_memberships`, `creator_role_id`, `admin_delete_enabled`, `domains_enabled`, `domains_enrollment_modes`, `domains_default_role_id`

### data.clerk_application

Reads an existing Clerk application by ID.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | The application ID to look up |

**Computed attributes:** `dev_instance_id`, `dev_publishable_key`, `prod_instance_id`, `prod_publishable_key`

## Development

### Prerequisites

- Go >= 1.22
- [golangci-lint](https://golangci-lint.run/)

### Commands

```bash
make build     # Build the provider binary
make test      # Run unit tests
make testacc   # Run acceptance tests (requires CLERK_PLATFORM_API_KEY)
make lint      # Run golangci-lint
make fmt       # Format code
make check     # Run fmt + lint + test
make install   # Install provider locally
make clean     # Remove build artifacts
```

### Running Acceptance Tests

Acceptance tests make real API calls to Clerk and require:

```bash
export CLERK_PLATFORM_API_KEY="your-platform-api-key"
export TF_ACC=1
go test ./... -v -timeout 120m
```

To run a single test:

```bash
TF_ACC=1 go test ./internal/provider -run TestAccClerkEnvironment_basic -v
```

### Architecture

```
internal/
  client/        # Clerk API client wrappers (Platform + Backend API)
  provider/      # Provider configuration, test helpers
  resources/     # Terraform resources (CRUD)
  datasources/   # Terraform data sources (read-only)
```

The provider follows a two-tier API pattern:
- **Platform API** (workspace-level): Manages applications, accessed with the platform API key
- **Backend API** (per-instance): Manages instance settings, accessed with per-instance secret keys resolved via internal key routing

## License

See [LICENSE](./LICENSE) for details.
