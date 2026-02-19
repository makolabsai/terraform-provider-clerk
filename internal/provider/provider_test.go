package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/makolabsai/terraform-provider-clerk/internal/provider"
)

// testAccProtoV6ProviderFactories is used in acceptance tests to instantiate
// the provider. It maps the provider name "clerk" to the factory function.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"clerk": providerserver.NewProtocol6WithError(provider.New()()),
}

// testAccPreCheck validates that the required environment variables are set
// before running acceptance tests. Tests will be skipped if prerequisites
// are not met.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}

	if os.Getenv("CLERK_PLATFORM_API_KEY") == "" {
		t.Skip("CLERK_PLATFORM_API_KEY must be set for acceptance tests")
	}
}

// testAccProviderConfig returns a Terraform configuration string that
// configures the Clerk provider using environment variables. This is
// prepended to resource configurations in acceptance tests.
func testAccProviderConfig() string {
	return `
provider "clerk" {}
`
}

func TestProvider_Schema(t *testing.T) {
	// Verify the provider can be instantiated without error.
	// This is a unit test (no TF_ACC required).
	p := provider.New()()
	if p == nil {
		t.Fatal("expected provider to be non-nil")
	}
}
