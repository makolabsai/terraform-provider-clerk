---
page_title: "clerk_environment Resource"
description: |-
  Configures a Clerk instance's settings per environment.
---

# clerk_environment

Configures a Clerk instance's settings (development or production). The instance is auto-created by Clerk when the application is created; this resource manages its configuration.

This resource covers three areas of instance configuration:
- **Instance settings** - General settings like HIBP, email deliverability, and support email
- **Restrictions** - Email validation rules (allowlist, blocklist, disposable domains)
- **Organization settings** - Organization feature configuration

~> **Note:** Authentication strategies (email/password/OAuth/MFA/session settings) are only configurable via the [Clerk Dashboard](https://clerk.com/docs/guides/configure/auth-strategies/sign-up-sign-in-options) and are not available through this resource.

~> **Note:** This resource does not support drift detection. Changes made via the Clerk Dashboard will not be detected by Terraform.

## Example Usage

### Basic Instance Settings

```hcl
resource "clerk_environment" "dev" {
  application_id = clerk_application.example.id
  environment    = "development"

  hibp          = true
  support_email = "dev-support@example.com"
}
```

### With Restrictions

```hcl
resource "clerk_environment" "prod" {
  application_id = clerk_application.example.id
  environment    = "production"

  hibp                          = true
  enhanced_email_deliverability = true
  support_email                 = "support@example.com"

  restrictions = {
    block_disposable_email_domains = true
    block_email_subaddresses       = true
  }
}
```

### With Organization Settings

```hcl
resource "clerk_environment" "prod_orgs" {
  application_id = clerk_application.example.id
  environment    = "production"

  organization_settings = {
    enabled                 = true
    max_allowed_memberships = 25
    admin_delete_enabled    = true
    domains_enabled         = true
    domains_enrollment_modes = ["automatic_invitation"]
  }
}
```

## Argument Reference

### Required

- `application_id` (String) - The Clerk application ID this environment belongs to. Changing this forces a new resource.
- `environment` (String) - The environment type: `"development"` or `"production"`. Changing this forces a new resource.

### Instance Settings (Optional)

- `test_mode` (Boolean) - Whether test mode is enabled. Defaults to `true` for development instances.
- `hibp` (Boolean) - Whether Have I Been Pwned password checking is enabled.
- `enhanced_email_deliverability` (Boolean) - Whether Clerk sends OTP emails via shared domain (Postmark) in production.
- `support_email` (String) - Contact email displayed to users needing support.
- `clerk_js_version` (String) - Specific Clerk.js version for hosted account pages.
- `url_based_session_syncing` (Boolean) - Whether URL-based session syncing is enabled.
- `development_origin` (String) - Origin URL for development instances.

### Restrictions Block (Optional)

- `restrictions` (Block) - Instance restriction settings:
  - `allowlist` (Boolean) - Whether the allowlist is enabled.
  - `blocklist` (Boolean) - Whether the blocklist is enabled.
  - `block_email_subaddresses` (Boolean) - Whether email subaddresses (user+tag@domain.com) are blocked.
  - `block_disposable_email_domains` (Boolean) - Whether disposable email domains are blocked.
  - `ignore_dots_for_gmail_addresses` (Boolean) - Whether dots are ignored in Gmail addresses for uniqueness.

### Organization Settings Block (Optional)

- `organization_settings` (Block) - Organization feature settings:
  - `enabled` (Boolean) - Whether organizations are enabled.
  - `max_allowed_memberships` (Number) - Maximum memberships per organization.
  - `creator_role_id` (String) - Role ID assigned to organization creators.
  - `admin_delete_enabled` (Boolean) - Whether admins can delete the organization.
  - `domains_enabled` (Boolean) - Whether organization domains are enabled.
  - `domains_enrollment_modes` (List of String) - Enrollment modes for organization domains.
  - `domains_default_role_id` (String) - Default role ID for domain-enrolled members.

## Attribute Reference

- `id` - Composite identifier in the format `{application_id}/{environment}`.

## Import

Environments can be imported using the composite ID:

```bash
terraform import clerk_environment.dev app_abc123/development
terraform import clerk_environment.prod app_abc123/production
```

~> **Note:** After import, all settings will show as unknown in state since there is no read API. Run `terraform apply` to push your desired configuration.

## Destroy Behavior

Destroying this resource does **not** delete the Clerk instance (instances are permanent). Instead, it resets instance settings and restrictions to their defaults.
