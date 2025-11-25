package entitytype_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/internal/acctest"
)

func TestAccDxEntityTypeResourceCreate(t *testing.T) {
	entityTypeIdentifier := fmt.Sprintf("tf_test_%d", acctest.RandInt())
	entityTypeName := fmt.Sprintf("Terraform Test Entity Type %d", acctest.RandInt())

	var testAccDxEntityTypeResourceBasic = fmt.Sprintf(`
provider "dx" {}

resource "dx_entity_type" "tf-integration-test" {
  identifier  = "%s"
  name        = "%s"
  description = "This is a test entity type created by Terraform"

  properties = {
    team = {
      name       = "Owning Team"
      type       = "multi_select"
      visibility = "visible"
      options = [
        { value = "platform", color = "#3b82f6" },
        { value = "data", color = "#ef4444" },
        { value = "product", color = "#10b981" }
      ]
    }
    tier = {
      name       = "Service Tier"
      type       = "text"
      visibility = "visible"
    }
  }

  aliases = {
    "github_repo" = true
  }
}
`, entityTypeIdentifier, entityTypeName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDxEntityTypeResourceBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "identifier", entityTypeIdentifier),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "name", entityTypeName),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "description", "This is a test entity type created by Terraform"),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "properties.%", "2"),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "properties.team.name", "Owning Team"),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "properties.team.type", "multi_select"),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "properties.tier.name", "Service Tier"),
					resource.TestCheckResourceAttr("dx_entity_type.tf-integration-test", "aliases.github_repo", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dx_entity_type.tf-integration-test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_entity_type" "tf-integration-test" {
  identifier  = "%s"
  name        = "%s Updated"
  description = "Updated description"

  properties = {
    team = {
      name       = "Team Owner"
      type       = "multi_select"
      visibility = "visible"
      options = [
        { value = "platform", color = "#3b82f6" },
        { value = "data", color = "#ef4444" },
        { value = "product", color = "#10b981" },
        { value = "infrastructure", color = "#f59e0b" }
      ]
    }
    tier = {
      name       = "Service Tier"
      type       = "text"
      visibility = "visible"
    }
    language = {
      name = "Programming Language"
      type = "text"
    }
  }

  aliases = {
    "github_repo" = true
  }
}
`, entityTypeIdentifier, entityTypeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity_type.test", "name", entityTypeName+" Updated"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.%", "3"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.team.name", "Team Owner"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.team.options.#", "4"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "aliases.github_repository", "true"),
				),
			},
			// Delete testing automatically occurs at the end
		},
	})
}

func TestAccDxEntityTypeResourceMinimal(t *testing.T) {
	entityTypeIdentifier := fmt.Sprintf("tf_minimal_%d", acctest.RandInt())
	entityTypeName := fmt.Sprintf("Minimal Entity Type %d", acctest.RandInt())

	var testAccDxEntityTypeResourceMinimal = fmt.Sprintf(`
provider "dx" {}

resource "dx_entity_type" "minimal" {
  identifier = "%s"
  name       = "%s"
}
`, entityTypeIdentifier, entityTypeName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDxEntityTypeResourceMinimal,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_entity_type.minimal", "identifier", entityTypeIdentifier),
					resource.TestCheckResourceAttr("dx_entity_type.minimal", "name", entityTypeName),
					resource.TestCheckResourceAttrSet("dx_entity_type.minimal", "created_at"),
					resource.TestCheckResourceAttrSet("dx_entity_type.minimal", "updated_at"),
				),
			},
		},
	})
}
