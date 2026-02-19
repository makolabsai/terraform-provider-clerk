package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccClerkOrganization_basic(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgName := "Test Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resourceName := "clerk_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkOrganizationConfig_basic(rName, orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
					resource.TestCheckResourceAttrSet(resourceName, "slug"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccClerkOrganization_update(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgName := "Test Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	orgNameUpdated := orgName + " Updated"
	resourceName := "clerk_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkOrganizationConfig_basic(rName, orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
				),
			},
			{
				Config: testAccClerkOrganizationConfig_basic(rName, orgNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", orgNameUpdated),
				),
			},
		},
	})
}

func TestAccClerkOrganization_maxMembers(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgName := "Test Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resourceName := "clerk_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkOrganizationConfig_maxMembers(rName, orgName, 25),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
					resource.TestCheckResourceAttr(resourceName, "max_allowed_memberships", "25"),
				),
			},
			// Update max memberships.
			{
				Config: testAccClerkOrganizationConfig_maxMembers(rName, orgName, 50),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "max_allowed_memberships", "50"),
				),
			},
		},
	})
}

func TestAccClerkOrganization_import(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgName := "Test Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resourceName := "clerk_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkOrganizationConfig_basic(rName, orgName),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf("%s/%s/%s",
						rs.Primary.Attributes["application_id"],
						rs.Primary.Attributes["environment"],
						rs.Primary.ID,
					), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccClerkOrganizationDataSource_byId(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgName := "Test Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resourceName := "clerk_organization.test"
	dataSourceName := "data.clerk_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkOrganizationDataSourceConfig_byId(rName, orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slug", resourceName, "slug"),
				),
			},
		},
	})
}

func TestAccClerkOrganizationDataSource_bySlug(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgName := "Test Org " + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resourceName := "clerk_organization.test"
	dataSourceName := "data.clerk_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkOrganizationDataSourceConfig_bySlug(rName, orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slug", resourceName, "slug"),
				),
			},
		},
	})
}

// --- Config helpers ---

// testAccClerkOrganizationBase returns the shared app + environment config
// required for organization tests. Organization settings must be enabled.
func testAccClerkOrganizationBase(appName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}

resource "clerk_environment" "test" {
  application_id = clerk_application.test.id
  environment    = "development"

  organization_settings = {
    enabled = true
  }
}
`, appName)
}

func testAccClerkOrganizationConfig_basic(appName, orgName string) string {
	return testAccClerkOrganizationBase(appName) + fmt.Sprintf(`
resource "clerk_organization" "test" {
  application_id = clerk_application.test.id
  environment    = "development"
  name           = %[1]q

  depends_on = [clerk_environment.test]
}
`, orgName)
}

func testAccClerkOrganizationConfig_maxMembers(appName, orgName string, maxMembers int) string {
	return testAccClerkOrganizationBase(appName) + fmt.Sprintf(`
resource "clerk_organization" "test" {
  application_id        = clerk_application.test.id
  environment           = "development"
  name                  = %[1]q
  max_allowed_memberships = %[2]d

  depends_on = [clerk_environment.test]
}
`, orgName, maxMembers)
}

func testAccClerkOrganizationDataSourceConfig_byId(appName, orgName string) string {
	return testAccClerkOrganizationConfig_basic(appName, orgName) + `
data "clerk_organization" "test" {
  application_id = clerk_application.test.id
  environment    = "development"
  id             = clerk_organization.test.id
}
`
}

func testAccClerkOrganizationDataSourceConfig_bySlug(appName, orgName string) string {
	return testAccClerkOrganizationConfig_basic(appName, orgName) + `
data "clerk_organization" "test" {
  application_id = clerk_application.test.id
  environment    = "development"
  slug           = clerk_organization.test.slug
}
`
}
