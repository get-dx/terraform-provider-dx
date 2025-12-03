package entity_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/internal/acctest"
)

func TestAccDxEntityResourceCreate(t *testing.T) {
	entityIdentifier := fmt.Sprintf("tf_test_entity_%d", acctest.RandInt())
	entityName := fmt.Sprintf("Terraform Test Entity %d", acctest.RandInt())

	var testAccDxEntityResourceBasic = fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "tf-integration-test" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "This is a test entity created by Terraform"

  properties = {
    tier            = "Tier-1"
    "slack-team" = "https://slack.com/channels/test-channel"
    language       = ["Go", "TypeScript"]
  }

  aliases = {
    github_repo = [
      {
        identifier = "520637360"
      }
    ]
  }

  relations = {
    "service-consumes-api" = ["defaultapi"]
  }
}
`, entityIdentifier, entityName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDxEntityResourceBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity.tf-integration-test", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("dx_entity.tf-integration-test", "type", "service"),
					resource.TestCheckResourceAttr("dx_entity.tf-integration-test", "name", entityName),
					resource.TestCheckResourceAttr("dx_entity.tf-integration-test", "description", "This is a test entity created by Terraform"),
					resource.TestCheckResourceAttrSet("dx_entity.tf-integration-test", "created_at"),
					resource.TestCheckResourceAttrSet("dx_entity.tf-integration-test", "updated_at"),
				),
			},
			// ImportState testing
			// Note: relations are not returned by entities.info API, so we ignore them during import
			{
				ResourceName:            "dx_entity.tf-integration-test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"relations", "properties"},
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "tf-integration-test" {
  identifier  = "%s"
  type        = "service"
  name        = "%s Updated"
  description = "Updated description for test entity"

  properties = {
    tier            = "Tier-2"
    "slack-team" = "https://slack.com/channels/updated-channel"
    language       = ["Go", "TypeScript", "Python"]
  }

  aliases = {
    github_repo = [
      {
        identifier = "520637360"
      },
      {
        identifier = "962275774"
      }
    ]
  }
}
`, entityIdentifier, entityName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity.tf-integration-test", "name", entityName+" Updated"),
					resource.TestCheckResourceAttr("dx_entity.tf-integration-test", "description", "Updated description for test entity"),
				),
			},
			// Delete testing automatically occurs at the end
		},
	})
}

func TestAccDxEntityResourceMinimal(t *testing.T) {
	entityIdentifier := fmt.Sprintf("tf_minimal_entity_%d", acctest.RandInt())

	var testAccDxEntityResourceMinimal = fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "minimal" {
  identifier = "%s"
  type       = "service"
}
`, entityIdentifier)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDxEntityResourceMinimal,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity.minimal", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("dx_entity.minimal", "type", "service"),
					resource.TestCheckResourceAttrSet("dx_entity.minimal", "created_at"),
					resource.TestCheckResourceAttrSet("dx_entity.minimal", "updated_at"),
				),
			},
		},
	})
}

func TestAccDxEntityResourceWithOptionalFields(t *testing.T) {
	entityIdentifier := fmt.Sprintf("tf_optional_entity_%d", acctest.RandInt())
	entityName := fmt.Sprintf("Optional Fields Entity %d", acctest.RandInt())

	// Test with some optional fields but not all (no aliases, no relations)
	var testAccDxEntityResourceOptional = fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "optional" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Entity with some optional fields"

  properties = {
    tier = "Tier-3"
  }
}
`, entityIdentifier, entityName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDxEntityResourceOptional,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity.optional", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("dx_entity.optional", "type", "service"),
					resource.TestCheckResourceAttr("dx_entity.optional", "name", entityName),
					resource.TestCheckResourceAttr("dx_entity.optional", "description", "Entity with some optional fields"),
					resource.TestCheckResourceAttrSet("dx_entity.optional", "created_at"),
					resource.TestCheckResourceAttrSet("dx_entity.optional", "updated_at"),
				),
			},
			// Update to add aliases and relations
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "optional" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Entity with some optional fields"

  properties = {
    tier = "Tier-3"
  }

  aliases = {
    github_repo = [
      {
        identifier = "962275774"
      }
    ]
  }

  relations = {
    "service-consumes-api" = ["defaultapi"]
  }
}
`, entityIdentifier, entityName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity.optional", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("dx_entity.optional", "type", "service"),
				),
			},
			// Update to remove aliases and relations (back to null)
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "optional" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Entity with some optional fields"

  properties = {
    tier = "Tier-3"
  }
}
`, entityIdentifier, entityName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity.optional", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("dx_entity.optional", "type", "service"),
				),
			},
		},
	})
}
