# Look up an organization by ID.
data "clerk_organization" "by_id" {
  application_id = clerk_application.my_app.id
  environment    = "development"
  id             = "org_abc123"
}

# Look up an organization by slug.
data "clerk_organization" "by_slug" {
  application_id = clerk_application.my_app.id
  environment    = "development"
  slug           = "acme-corp"
}

output "org_name" {
  value = data.clerk_organization.by_slug.name
}
