# Configure the development environment for an application.
resource "clerk_application" "payments" {
  name                = "Payments App"
  deletion_protection = false
}

# Basic instance settings.
resource "clerk_environment" "dev" {
  application_id = clerk_application.payments.id
  environment    = "development"

  hibp          = true
  support_email = "dev-support@example.com"
}

# Configure restrictions to block disposable emails.
resource "clerk_environment" "prod" {
  application_id = clerk_application.payments.id
  environment    = "production"

  hibp                          = true
  enhanced_email_deliverability = true
  support_email                 = "support@example.com"

  restrictions = {
    block_disposable_email_domains = true
    block_email_subaddresses       = true
  }

  organization_settings = {
    enabled                 = true
    max_allowed_memberships = 25
    admin_delete_enabled    = true
  }
}
