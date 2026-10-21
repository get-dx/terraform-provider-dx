package relationedge_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/internal/acctest"
)

func testAccRelationEdgeConfig(suffix int, target string) string {
	return fmt.Sprintf(`
provider "dx" {}

resource "dx_catalog_relation" "test" {
  identifier                    = "tf_test_edge_rel_%[1]d"
  type                          = "depends on"
  cardinality                   = "many_to_many"
  source_entity_type_identifier = "service"
  target_entity_type_identifier = "service"
}

resource "dx_entity" "source" {
  identifier = "tf_test_edge_src_%[1]d"
  type       = "service"
  name       = "Terraform Edge Source %[1]d"
}

resource "dx_entity" "target_a" {
  identifier = "tf_test_edge_tgt_a_%[1]d"
  type       = "service"
  name       = "Terraform Edge Target A %[1]d"
}

resource "dx_entity" "target_b" {
  identifier = "tf_test_edge_tgt_b_%[1]d"
  type       = "service"
  name       = "Terraform Edge Target B %[1]d"
}

resource "dx_catalog_relation_edge" "test" {
  relation_identifier      = dx_catalog_relation.test.identifier
  source_entity_identifier = dx_entity.source.identifier
  target_entity_identifier = dx_entity.%[2]s.identifier
}
`, suffix, target)
}

func TestAccDxCatalogRelationEdgeResource(t *testing.T) {
	suffix := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRelationEdgeConfig(suffix, "target_a"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_catalog_relation_edge.test", "relation_identifier", fmt.Sprintf("tf_test_edge_rel_%d", suffix)),
					resource.TestCheckResourceAttr("dx_catalog_relation_edge.test", "source_entity_identifier", fmt.Sprintf("tf_test_edge_src_%d", suffix)),
					resource.TestCheckResourceAttr("dx_catalog_relation_edge.test", "target_entity_identifier", fmt.Sprintf("tf_test_edge_tgt_a_%d", suffix)),
					resource.TestCheckResourceAttr("dx_catalog_relation_edge.test", "id", fmt.Sprintf("tf_test_edge_rel_%[1]d/tf_test_edge_src_%[1]d/tf_test_edge_tgt_a_%[1]d", suffix)),
				),
			},
			// Changing the target replaces the edge
			{
				Config: testAccRelationEdgeConfig(suffix, "target_b"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_catalog_relation_edge.test", "target_entity_identifier", fmt.Sprintf("tf_test_edge_tgt_b_%d", suffix)),
				),
			},
			// Delete testing automatically occurs at the end
		},
	})
}
