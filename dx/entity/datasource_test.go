package entity_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/internal/acctest"
)

func TestAccDxEntityDataSource(t *testing.T) {
	entityIdentifier := fmt.Sprintf("tf_test_ds_entity_%d", acctest.RandInt())
	entityName := fmt.Sprintf("Terraform Test DS Entity %d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create an entity and then read it with a data source
			{
				Config: testAccEntityDataSourceConfig(entityIdentifier, entityName),
				Check: resource.ComposeTestCheckFunc(
					// Check the resource was created
					resource.TestCheckResourceAttr("dx_entity.test", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("dx_entity.test", "type", "service"),
					resource.TestCheckResourceAttr("dx_entity.test", "name", entityName),

					// Check the data source reads the same entity
					resource.TestCheckResourceAttr("data.dx_entity.test", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("data.dx_entity.test", "type", "service"),
					resource.TestCheckResourceAttr("data.dx_entity.test", "name", entityName),
					resource.TestCheckResourceAttr("data.dx_entity.test", "description", "Test entity for data source"),
					resource.TestCheckResourceAttrSet("data.dx_entity.test", "id"),
					resource.TestCheckResourceAttrSet("data.dx_entity.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.dx_entity.test", "updated_at"),

					// Verify computed fields match
					resource.TestCheckResourceAttrPair("dx_entity.test", "id", "data.dx_entity.test", "id"),
					resource.TestCheckResourceAttrPair("dx_entity.test", "created_at", "data.dx_entity.test", "created_at"),
				),
			},
		},
	})
}

func TestAccDxEntityDataSourceWithProperties(t *testing.T) {
	entityIdentifier := fmt.Sprintf("tf_test_ds_props_%d", acctest.RandInt())
	entityName := fmt.Sprintf("Test Props Entity %d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityDataSourceConfigWithProperties(entityIdentifier, entityName),
				Check: resource.ComposeTestCheckFunc(
					// Verify data source reads properties
					resource.TestCheckResourceAttr("data.dx_entity.test_props", "identifier", entityIdentifier),
					resource.TestCheckResourceAttrSet("data.dx_entity.test_props", "properties"),

					// Verify aliases are read
					resource.TestCheckResourceAttrSet("data.dx_entity.test_props", "aliases"),
				),
			},
		},
	})
}

func testAccEntityDataSourceConfig(identifier, name string) string {
	return fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "test" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Test entity for data source"
  
  properties = {
    tier = "Tier-1"
  }
}

data "dx_entity" "test" {
  identifier = dx_entity.test.identifier
}
`, identifier, name)
}

func testAccEntityDataSourceConfigWithProperties(identifier, name string) string {
	return fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "test_props" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Entity with properties and aliases"
  
  properties = {
    tier        = "Tier-1"
    language    = ["Go", "Typescript"]
    environment = "production"
  }
  
  aliases = {
    github_repo = [
      {
        identifier = "520637360"
      }
    ]
  }
}

data "dx_entity" "test_props" {
  identifier = dx_entity.test_props.identifier
}

# Test that we can use the data source outputs
output "entity_type" {
  value = data.dx_entity.test_props.type
}

output "entity_properties" {
  value = data.dx_entity.test_props.properties
}
`, identifier, name)
}
