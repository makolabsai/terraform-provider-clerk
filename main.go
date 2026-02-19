package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/makolabsai/terraform-provider-clerk/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New(), providerserver.ServeOpts{
		Address: "registry.terraform.io/makolabsai/clerk",
	})
	if err != nil {
		log.Fatal(err)
	}
}
