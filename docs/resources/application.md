---
page_title: "clerk_application Resource"
description: |-
  Manages a Clerk application.
---

# clerk_application

Manages a Clerk application. Each application can have multiple instances (development, production) with distinct user pools.

When created, Clerk automatically provisions development and production instances. The resource exposes instance IDs, secret keys, and publishable keys for both environments.

## Example Usage

### Basic Application

```hcl
resource "clerk_application" "example" {
  name = "My Application"
}
```

### Application with Deletion Protection Disabled

```hcl
resource "clerk_application" "testing" {
  name                = "Testing App"
  deletion_protection = false
}
```

### Application from Template

```hcl
resource "clerk_application" "saas" {
  name     = "SaaS Platform"
  template = "b2b-saas"
}
```

## Argument Reference

- `name` (String, Required) - The name of the application.
- `deletion_protection` (Boolean, Optional) - Whether deletion protection is enabled. When `true`, the application cannot be destroyed. Set to `false` before destroying. Defaults to `true`.
- `domain` (String, Optional) - The domain for the application. Only set at creation time; changing this forces a new resource.
- `template` (String, Optional) - Application template (e.g., `b2b-saas`, `b2c-saas`, `waitlist`). Only set at creation time; changing this forces a new resource.
- `environment_types` (List of String, Optional) - List of environment types to create instances for. Only set at creation time; changing this forces a new resource.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier of the Clerk application.
- `dev_instance_id` - The instance ID for the development environment.
- `dev_secret_key` (Sensitive) - The secret key for the development instance.
- `dev_publishable_key` - The publishable key for the development instance.
- `prod_instance_id` - The instance ID for the production environment.
- `prod_secret_key` (Sensitive) - The secret key for the production instance.
- `prod_publishable_key` - The publishable key for the production instance.

## Import

Applications can be imported using the application ID:

```bash
terraform import clerk_application.example app_abc123
```

~> **Note:** The `name`, `deletion_protection`, and secret key fields cannot be recovered on import since the Clerk API does not return them in the GET response.
