---
page_title: "clerk_organization Resource"
description: |-
  Manages a Clerk organization within a specific application environment.
---

# clerk_organization

Manages a Clerk organization within a specific application environment. Organizations represent tenants or teams that group users together.

~> **Note:** Organizations must be enabled on the Clerk instance before creating organization resources. Use the `organization_settings` block in `clerk_environment` to enable them.

## Example Usage

### Basic Organization

```hcl
resource "clerk_organization" "basic" {
  application_id = clerk_application.my_app.id
  environment    = "development"
  name           = "Acme Corp"
}
```

### Organization with Custom Slug and Membership Limit

```hcl
resource "clerk_organization" "enterprise" {
  application_id          = clerk_application.my_app.id
  environment             = "development"
  name                    = "Enterprise Client"
  slug                    = "enterprise-client"
  max_allowed_memberships = 100
}
```

## Argument Reference

### Required

- `application_id` (String) - The Clerk application ID this organization belongs to. Changing this forces a new resource.
- `environment` (String) - The environment type: `"development"` or `"production"`. Changing this forces a new resource.
- `name` (String) - The name of the organization.

### Optional

- `slug` (String) - URL-friendly identifier for the organization. Auto-generated from name if not provided.
- `max_allowed_memberships` (Number) - Maximum number of memberships allowed in the organization. 0 means unlimited.
- `admin_delete_enabled` (Boolean) - Whether organization admins can delete the organization. Only settable on update, not on creation.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier of the Clerk organization.
- `created_at` - Unix timestamp of when the organization was created.
- `updated_at` - Unix timestamp of when the organization was last updated.

## Import

Organizations can be imported using the composite ID format `{application_id}/{environment}/{organization_id}`:

```bash
terraform import clerk_organization.example app_abc123/development/org_xyz789
```
