package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccClerkEnvironment_basic(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create app + configure its dev environment.
			{
				Config: testAccClerkEnvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "environment", "development"),
					resource.TestCheckResourceAttr(resourceName, "hibp", "true"),
					resource.TestCheckResourceAttr(resourceName, "support_email", "support@test.com"),
				),
			},
		},
	})
}

func TestAccClerkEnvironment_restrictions(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkEnvironmentConfig_restrictions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "restrictions.block_disposable_email_domains", "true"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.block_email_subaddresses", "true"),
				),
			},
		},
	})
}

func TestAccClerkEnvironment_organizationSettings(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkEnvironmentConfig_orgSettings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "organization_settings.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "organization_settings.max_allowed_memberships", "10"),
				),
			},
		},
	})
}

func TestAccClerkEnvironment_update(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with initial settings.
			{
				Config: testAccClerkEnvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_email", "support@test.com"),
				),
			},
			// Update support email.
			{
				Config: testAccClerkEnvironmentConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_email", "updated@test.com"),
				),
			},
		},
	})
}

func TestAccClerkEnvironment_import(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkEnvironmentConfig_basic(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // No read API means state won't match after import.
			},
		},
	})
}

// --- Config helpers ---

func testAccClerkEnvironmentConfig_basic(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}

resource "clerk_environment" "test" {
  application_id = clerk_application.test.id
  environment    = "development"

  hibp          = true
  support_email = "support@test.com"
}
`, name)
}

func testAccClerkEnvironmentConfig_updated(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}

resource "clerk_environment" "test" {
  application_id = clerk_application.test.id
  environment    = "development"

  hibp          = true
  support_email = "updated@test.com"
}
`, name)
}

func testAccClerkEnvironmentConfig_restrictions(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}

resource "clerk_environment" "test" {
  application_id = clerk_application.test.id
  environment    = "development"

  restrictions = {
    block_disposable_email_domains = true
    block_email_subaddresses       = true
  }
}
`, name)
}

func testAccClerkEnvironmentConfig_orgSettings(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}

resource "clerk_environment" "test" {
  application_id = clerk_application.test.id
  environment    = "development"

  organization_settings = {
    enabled                = true
    max_allowed_memberships = 10
  }
}
`, name)
}
