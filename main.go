package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/makolabsai/terraform-provider-clerk/internal/provider"
)

// version is set at build time via ldflags by GoReleaser.
var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.NewWithVersion(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/makolabsai/clerk",
	})
	if err != nil {
		log.Fatal(err)
	}
}
