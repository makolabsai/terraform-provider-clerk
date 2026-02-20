---
page_title: "clerk_organization Data Source"
description: |-
  Reads an existing Clerk organization by ID or slug.
---

# clerk_organization (Data Source)

Reads an existing Clerk organization by ID or slug. Use this to reference organizations that were created outside of Terraform or in a different Terraform configuration.

Exactly one of `id` or `slug` must be specified.

## Example Usage

### Look Up by ID

```hcl
data "clerk_organization" "by_id" {
  application_id = clerk_application.my_app.id
  environment    = "development"
  id             = "org_abc123"
}
```

### Look Up by Slug

```hcl
data "clerk_organization" "by_slug" {
  application_id = clerk_application.my_app.id
  environment    = "development"
  slug           = "acme-corp"
}

output "org_name" {
  value = data.clerk_organization.by_slug.name
}
```

## Argument Reference

### Required

- `application_id` (String) - The Clerk application ID the organization belongs to.
- `environment` (String) - The environment type: `"development"` or `"production"`.

### Optional (exactly one required)

- `id` (String) - The organization ID to look up.
- `slug` (String) - The organization slug to look up.

## Attribute Reference

- `id` - The unique identifier of the organization.
- `name` - The name of the organization.
- `slug` - URL-friendly identifier for the organization.
- `max_allowed_memberships` - Maximum number of memberships allowed.
- `admin_delete_enabled` - Whether organization admins can delete the organization.
- `created_at` - Unix timestamp of when the organization was created.
