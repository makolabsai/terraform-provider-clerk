terraform {
  required_providers {
    clerk = {
      source = "registry.terraform.io/makolabsai/clerk"
    }
  }
}

# Configure the Clerk provider with a Platform API key.
# The key can also be set via the CLERK_PLATFORM_API_KEY environment variable.
provider "clerk" {
  platform_api_key = var.clerk_platform_api_key
}

variable "clerk_platform_api_key" {
  description = "Clerk Platform API key for workspace-level operations"
  type        = string
  sensitive   = true
}
