package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccClerkProvider_endToEnd exercises the full lifecycle:
// application → environment configuration → organization creation/update/deletion.
// This validates that key routing (platform key → per-instance secret key) works
// correctly across all three resource types in a single test.
func TestAccClerkProvider_endToEnd(t *testing.T) {
	rName := "tf-acc-e2e-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	orgName := "E2E Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	orgNameUpdated := orgName + " Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create app + environment + organization.
			{
				Config: testAccE2EConfig(rName, orgName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Application created.
					resource.TestCheckResourceAttrSet("clerk_application.e2e", "id"),
					resource.TestCheckResourceAttr("clerk_application.e2e", "name", rName),

					// Environment configured.
					resource.TestCheckResourceAttrSet("clerk_environment.e2e", "id"),
					resource.TestCheckResourceAttr("clerk_environment.e2e", "organization_settings.enabled", "true"),

					// Organization created.
					resource.TestCheckResourceAttrSet("clerk_organization.e2e", "id"),
					resource.TestCheckResourceAttr("clerk_organization.e2e", "name", orgName),
					resource.TestCheckResourceAttr("clerk_organization.e2e", "max_allowed_memberships", "10"),

					// Data source reads the organization back.
					resource.TestCheckResourceAttrPair(
						"data.clerk_organization.e2e", "id",
						"clerk_organization.e2e", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.clerk_organization.e2e", "name",
						"clerk_organization.e2e", "name",
					),
				),
			},
			// Step 2: Update organization name and max memberships.
			{
				Config: testAccE2EConfig(rName, orgNameUpdated, 50),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("clerk_organization.e2e", "name", orgNameUpdated),
					resource.TestCheckResourceAttr("clerk_organization.e2e", "max_allowed_memberships", "50"),
				),
			},
		},
	})
}

func testAccE2EConfig(appName, orgName string, maxMembers int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "e2e" {
  name                = %[1]q
  deletion_protection = false
}

resource "clerk_environment" "e2e" {
  application_id = clerk_application.e2e.id
  environment    = "development"

  organization_settings = {
    enabled = true
  }
}

resource "clerk_organization" "e2e" {
  application_id          = clerk_application.e2e.id
  environment             = "development"
  name                    = %[2]q
  max_allowed_memberships = %[3]d

  depends_on = [clerk_environment.e2e]
}

data "clerk_organization" "e2e" {
  application_id = clerk_application.e2e.id
  environment    = "development"
  id             = clerk_organization.e2e.id
}
`, appName, orgName, maxMembers)
}
