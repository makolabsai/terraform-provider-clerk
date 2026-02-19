package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccClerkApplication_basic(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and verify.
			{
				Config: testAccClerkApplicationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "dev_instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dev_publishable_key"),
					resource.TestCheckResourceAttrSet(resourceName, "dev_secret_key"),
				),
			},
			// Import state.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// name is not returned by the API, so it can't be verified on import.
				// deletion_protection is provider-side only, not in the API.
				// secret keys also require include_secret_keys=true which import may not trigger identically.
				ImportStateVerifyIgnore: []string{"name", "template", "deletion_protection", "dev_secret_key", "prod_secret_key"},
			},
		},
	})
}

func TestAccClerkApplication_update(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	rNameUpdated := rName + "-updated"
	resourceName := "clerk_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccClerkApplicationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			// Update name (in-place, no replacement).
			{
				Config: testAccClerkApplicationConfig(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccClerkApplication_deletionProtection(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with deletion_protection = true (explicit).
			{
				Config: testAccClerkApplicationConfigWithDeletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			// Disable deletion_protection so the test can clean up.
			{
				Config: testAccClerkApplicationConfigWithDeletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccClerkApplicationDataSource_basic(t *testing.T) {
	rName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	resourceName := "clerk_application.test"
	dataSourceName := "data.clerk_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClerkApplicationDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dev_instance_id", resourceName, "dev_instance_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dev_publishable_key", resourceName, "dev_publishable_key"),
				),
			},
		},
	})
}

func testAccClerkApplicationConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}
`, name)
}

func testAccClerkApplicationConfigWithDeletionProtection(name string, protected bool) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = %[2]t
}
`, name, protected)
}

func testAccClerkApplicationDataSourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "clerk_application" "test" {
  name                = %[1]q
  deletion_protection = false
}

data "clerk_application" "test" {
  id = clerk_application.test.id
}
`, name)
}
