package relation_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/internal/acctest"
)

func TestAccDxCatalogRelationResource(t *testing.T) {
	relationIdentifier := fmt.Sprintf("tf_test_rel_%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_catalog_relation" "test" {
  identifier                     = "%s"
  type                           = "depends on"
  cardinality                    = "many_to_many"
  source_entity_type_identifier  = "service"
  target_entity_type_identifier  = "service"
  description                    = "Service depends on another service"
}
`, relationIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "identifier", relationIdentifier),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "type", "depends on"),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "cardinality", "many_to_many"),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "source_entity_type_identifier", "service"),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "target_entity_type_identifier", "service"),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "description", "Service depends on another service"),
					resource.TestCheckResourceAttrSet("dx_catalog_relation.test", "inverse_type"),
					resource.TestCheckResourceAttrSet("dx_catalog_relation.test", "created_at"),
					resource.TestCheckResourceAttrSet("dx_catalog_relation.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dx_catalog_relation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update type and description
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_catalog_relation" "test" {
  identifier                     = "%s"
  type                           = "consumes"
  cardinality                    = "many_to_many"
  source_entity_type_identifier  = "service"
  target_entity_type_identifier  = "service"
  description                    = "Updated description"
}
`, relationIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "type", "consumes"),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "cardinality", "many_to_many"),
				),
			},
			// Remove description
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_catalog_relation" "test" {
  identifier                     = "%s"
  type                           = "consumes"
  cardinality                    = "many_to_many"
  source_entity_type_identifier  = "service"
  target_entity_type_identifier  = "service"
}
`, relationIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_catalog_relation.test", "type", "consumes"),
					resource.TestCheckNoResourceAttr("dx_catalog_relation.test", "description"),
				),
			},
			// Delete testing automatically occurs at the end
		},
	})
}

func TestAccDxCatalogRelationResourceMinimal(t *testing.T) {
	relationIdentifier := fmt.Sprintf("tf_test_rel_min_%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_catalog_relation" "minimal" {
  identifier                     = "%s"
  type                           = "parent of"
  cardinality                    = "one_to_many"
  source_entity_type_identifier  = "service"
  target_entity_type_identifier  = "service"
}
`, relationIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_catalog_relation.minimal", "identifier", relationIdentifier),
					resource.TestCheckResourceAttr("dx_catalog_relation.minimal", "type", "parent of"),
					resource.TestCheckResourceAttr("dx_catalog_relation.minimal", "cardinality", "one_to_many"),
					resource.TestCheckResourceAttrSet("dx_catalog_relation.minimal", "created_at"),
					resource.TestCheckResourceAttrSet("dx_catalog_relation.minimal", "updated_at"),
				),
			},
		},
	})
}
