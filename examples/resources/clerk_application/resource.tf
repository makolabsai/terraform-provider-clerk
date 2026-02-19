# Create a Clerk application with default settings.
resource "clerk_application" "basic" {
  name = "My App"
}

# Create an application with deletion protection disabled (for testing).
resource "clerk_application" "dev_only" {
  name                = "Dev Testing App"
  deletion_protection = false
}

# Create an application from a template.
resource "clerk_application" "saas" {
  name     = "SaaS Platform"
  template = "b2b-saas"
}

# Access instance keys for downstream use.
output "dev_publishable_key" {
  value = clerk_application.basic.dev_publishable_key
}

output "dev_secret_key" {
  value     = clerk_application.basic.dev_secret_key
  sensitive = true
}
