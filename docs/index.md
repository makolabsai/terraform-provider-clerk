---
page_title: "Clerk Provider"
description: |-
  The Clerk provider enables managing Clerk authentication platform resources via Terraform.
---

# Clerk Provider

The Clerk provider enables managing [Clerk](https://clerk.com) authentication platform resources via Terraform/OpenTofu. It supports creating applications, configuring instance settings, and managing environment-level configuration.

## Authentication

The provider requires a Clerk Platform API key for workspace-level operations.

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

## Example Usage

```hcl
terraform {
  required_providers {
    clerk = {
      source = "registry.terraform.io/makolabsai/clerk"
    }
  }
}

provider "clerk" {}

resource "clerk_application" "example" {
  name = "My Application"
}

resource "clerk_environment" "example_dev" {
  application_id = clerk_application.example.id
  environment    = "development"

  hibp          = true
  support_email = "support@example.com"
}
```

## Schema

### Optional

- `platform_api_key` (String, Sensitive) - The Clerk Platform API key. Can also be set via `CLERK_PLATFORM_API_KEY` environment variable.
