---
page_title: "clerk_application Data Source"
description: |-
  Reads an existing Clerk application by ID.
---

# clerk_application (Data Source)

Reads an existing Clerk application by ID. Use this to reference applications that were created outside of Terraform or in a different Terraform configuration.

## Example Usage

```hcl
data "clerk_application" "existing" {
  id = "app_abc123"
}

output "dev_publishable_key" {
  value = data.clerk_application.existing.dev_publishable_key
}
```

### Reference in Other Resources

```hcl
data "clerk_application" "main" {
  id = "app_abc123"
}

resource "clerk_environment" "dev" {
  application_id = data.clerk_application.main.id
  environment    = "development"

  hibp = true
}
```

## Argument Reference

- `id` (String, Required) - The unique identifier of the Clerk application to look up.

## Attribute Reference

- `dev_instance_id` - The instance ID for the development environment.
- `dev_publishable_key` - The publishable key for the development instance.
- `dev_secret_key` (Sensitive) - The secret key for the development instance.
- `prod_instance_id` - The instance ID for the production environment.
- `prod_publishable_key` - The publishable key for the production instance.
- `prod_secret_key` (Sensitive) - The secret key for the production instance.
