package entitytype_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/dx/entitytype"
	"terraform-provider-dx/internal/acctest"
)

func TestValidateModel_MissingIdentifier(t *testing.T) {
	var model = entitytype.EntityTypeModel{
		Name: types.StringValue("Test Entity Type"),
	}

	diags := diag.Diagnostics{}
	entitytype.ValidateModel(model, &diags)

	if !diags.HasError() {
		t.Error("Expected validation to fail due to missing identifier, but it passed")
		return
	}

	if len(diags) < 1 {
		t.Errorf("Expected at least 1 validation error, got %d", len(diags))
	}
}

func TestValidateModel_MissingName(t *testing.T) {
	var model = entitytype.EntityTypeModel{
		Identifier: types.StringValue("test_entity"),
	}

	diags := diag.Diagnostics{}
	entitytype.ValidateModel(model, &diags)

	if !diags.HasError() {
		t.Error("Expected validation to fail due to missing name, but it passed")
		return
	}

	if len(diags) < 1 {
		t.Errorf("Expected at least 1 validation error, got %d", len(diags))
	}
}

func TestValidateModel_DuplicatePropertyIdentifiers(t *testing.T) {
	// With map structure, duplicate keys are not possible in Go maps,
	// so this test is no longer applicable. The map key itself is the identifier.
	// This test is kept for reference but will always pass with map structure.
	var model = entitytype.EntityTypeModel{
		Identifier: types.StringValue("test_entity"),
		Name:       types.StringValue("Test Entity Type"),
		Properties: map[string]entitytype.PropertyModel{
			"team": {
				Name: types.StringValue("Team"),
				Type: types.StringValue("text"),
			},
		},
	}

	diags := diag.Diagnostics{}
	entitytype.ValidateModel(model, &diags)

	// With map structure, duplicate identifiers are impossible (map keys are unique)
	// So this test should pass
	if diags.HasError() {
		t.Errorf("Expected validation to pass with map structure, but got errors: %v", diags)
	}
}

func TestValidateModel_MissingPropertyFields(t *testing.T) {
	var model = entitytype.EntityTypeModel{
		Identifier: types.StringValue("test_entity"),
		Name:       types.StringValue("Test Entity Type"),
		Properties: map[string]entitytype.PropertyModel{
			"team": {
				// Missing name and type
			},
		},
	}

	diags := diag.Diagnostics{}
	entitytype.ValidateModel(model, &diags)

	if !diags.HasError() {
		t.Error("Expected validation to fail due to missing property fields, but it passed")
		return
	}

	// Should have at least 2 errors (name, type) - identifier is now the map key
	if len(diags) < 2 {
		t.Errorf("Expected at least 2 validation errors, got %d", len(diags))
	}
}

func TestValidateModel_ValidModel(t *testing.T) {
	var model = entitytype.EntityTypeModel{
		Identifier:  types.StringValue("service"),
		Name:        types.StringValue("Service"),
		Description: types.StringValue("A deployable service"),
		Properties: map[string]entitytype.PropertyModel{
			"team": {
				Name:       types.StringValue("Owning Team"),
				Type:       types.StringValue("multi_select"),
				Visibility: types.StringValue("visible"),
				Options: []entitytype.PropertyOptionModel{
					{
						Value: types.StringValue("platform"),
						Color: types.StringValue("#3b82f6"),
					},
					{
						Value: types.StringValue("data"),
						Color: types.StringValue("#ef4444"),
					},
				},
			},
		},
	}

	diags := diag.Diagnostics{}
	entitytype.ValidateModel(model, &diags)

	if diags.HasError() {
		t.Errorf("Expected validation to pass, but got errors: %v", diags)
	}
}

func TestAccDxEntityTypeResourceCreate(t *testing.T) {
	entityTypeIdentifier := fmt.Sprintf("tf_test_%d", acctest.RandInt())
	entityTypeName := fmt.Sprintf("Terraform Test Entity Type %d", acctest.RandInt())

	var testAccDxEntityTypeResourceBasic = fmt.Sprintf(`
provider "dx" {}

resource "dx_entity_type" "test" {
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
    "github_repository" = true
    "pagerduty_service" = true
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
					resource.TestCheckResourceAttr("dx_entity_type.test", "identifier", entityTypeIdentifier),
					resource.TestCheckResourceAttr("dx_entity_type.test", "name", entityTypeName),
					resource.TestCheckResourceAttr("dx_entity_type.test", "description", "This is a test entity type created by Terraform"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.%", "2"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.team.name", "Owning Team"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.team.type", "multi_select"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "properties.tier.name", "Service Tier"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "aliases.github_repository", "true"),
					resource.TestCheckResourceAttr("dx_entity_type.test", "aliases.pagerduty_service", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dx_entity_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(`
provider "dx" {}

resource "dx_entity_type" "test" {
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
    "github_repository" = true
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
