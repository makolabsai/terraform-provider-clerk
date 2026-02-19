# Look up an existing Clerk application by ID.
data "clerk_application" "existing" {
  id = "app_abc123"
}

output "dev_instance_id" {
  value = data.clerk_application.existing.dev_instance_id
}

output "dev_publishable_key" {
  value = data.clerk_application.existing.dev_publishable_key
}
