# Create a basic organization within an application's development environment.
resource "clerk_organization" "basic" {
  application_id = clerk_application.my_app.id
  environment    = "development"
  name           = "Acme Corp"
}

# Create an organization with a custom slug and membership limit.
resource "clerk_organization" "with_options" {
  application_id          = clerk_application.my_app.id
  environment             = "development"
  name                    = "Enterprise Client"
  slug                    = "enterprise-client"
  max_allowed_memberships = 100
}

# Import an existing organization using the composite ID format:
#   terraform import clerk_organization.existing {application_id}/{environment}/{organization_id}
