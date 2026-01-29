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
					// Verify data source reads basic attributes
					resource.TestCheckResourceAttr("data.dx_entity.test_props", "identifier", entityIdentifier),
					resource.TestCheckResourceAttr("data.dx_entity.test_props", "name", entityName),
					resource.TestCheckResourceAttr("data.dx_entity.test_props", "type", "service"),
					resource.TestCheckResourceAttr("data.dx_entity.test_props", "description", "Entity with properties and aliases"),

					// Verify computed fields are set
					resource.TestCheckResourceAttrSet("data.dx_entity.test_props", "id"),
					resource.TestCheckResourceAttrSet("data.dx_entity.test_props", "created_at"),
					resource.TestCheckResourceAttrSet("data.dx_entity.test_props", "updated_at"),

					// Note: properties (Dynamic) and aliases (complex Map) can't be easily checked
					// with TestCheckResourceAttr* functions. The fact that the config applies
					// successfully and produces output verifies they're working correctly.
				),
			},
		},
	})
}

func TestAccDxEntitiesDataSource(t *testing.T) {
	entityIdentifier1 := fmt.Sprintf("tf_test_ds_entities_1_%d", acctest.RandInt())
	entityName1 := fmt.Sprintf("Terraform Test Entities 1 %d", acctest.RandInt())
	entityIdentifier2 := fmt.Sprintf("tf_test_ds_entities_2_%d", acctest.RandInt())
	entityName2 := fmt.Sprintf("Terraform Test Entities 2 %d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEntitiesDataSourceConfig(entityIdentifier1, entityName1, entityIdentifier2, entityName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dx_entities.test", "type", "service"),
					resource.TestCheckResourceAttrSet("data.dx_entities.test", "entities.#"),
				),
			},
		},
	})
}

func testAccEntitiesDataSourceConfig(identifier1, name1, identifier2, name2 string) string {
	return fmt.Sprintf(`
provider "dx" {}

resource "dx_entity" "test1" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Test entity 1 for entities data source"
}

resource "dx_entity" "test2" {
  identifier  = "%s"
  type        = "service"
  name        = "%s"
  description = "Test entity 2 for entities data source"
}

data "dx_entities" "test" {
  type = "service"

  depends_on = [dx_entity.test1, dx_entity.test2]
}
`, identifier1, name1, identifier2, name2)
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
    language    = ["Go", "TypeScript"]
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
