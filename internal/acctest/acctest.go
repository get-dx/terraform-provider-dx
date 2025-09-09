package acctest

import (
	"os"
	"terraform-provider-dx/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dx": providerserver.NewProtocol6WithError(provider.New("acceptance-tests")()),
	}
)

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("DX_WEB_API_TOKEN"); v == "" {
		t.Fatal("DX_WEB_API_TOKEN must be set for acceptance tests")
	}
}
