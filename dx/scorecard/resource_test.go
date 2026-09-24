package scorecard_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-dx/dx/scorecard"
	"terraform-provider-dx/internal/acctest"
)

func TestDuplicateOrderingWithinLevel(t *testing.T) {
	var model = scorecard.ScorecardModel{
		Name:                        types.StringValue("Terraform Provider Scorecard"),
		Type:                        types.StringValue("LEVEL"),
		EntityFilterType:            types.StringValue("entity_types"),
		EntityFilterTypeIdentifiers: []types.String{types.StringValue("service")},
		EmptyLevelLabel:             types.StringValue("Incomplete"),
		EmptyLevelColor:             types.StringValue("#cccccc"),
		Published:                   types.BoolValue(true),
		Tags:                        []scorecard.TagModel{{Value: types.StringValue("test")}},
		Levels: map[string]scorecard.LevelModel{
			"bronze": {
				Name:  types.StringValue("Bronze"),
				Color: types.StringValue("#FB923C"),
				Rank:  types.Int32Value(1),
			},
		},
		Checks: map[string]scorecard.CheckModel{
			"check_a": {
				Name:              types.StringValue("Check A"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(0), // Duplicate ordering!
			},
			"check_b": {
				Name:              types.StringValue("Check B"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(0), // Duplicate ordering!
			},
		},
	}

	diags := diag.Diagnostics{}

	scorecard.ValidateModel(model, &diags)

	if !diags.HasError() {
		t.Error("Expected validation to fail due to duplicate ordering, but it passed")
		return
	}

	if len(diags) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(diags))
		return
	}

	expectedMsg := "Level `bronze`: the following checks have a duplicate ordering of 0: `check_a`, `check_b`"
	actualMsg := diags[0].Detail()

	if actualMsg != expectedMsg {
		t.Errorf("Expected error message:\n%s\n\nGot:\n%s", expectedMsg, actualMsg)
	}
}

func TestDuplicateOrderingWithinCheckGroup(t *testing.T) {
	var model = scorecard.ScorecardModel{
		Name:                        types.StringValue("Terraform Provider Points Scorecard"),
		Type:                        types.StringValue("POINTS"),
		EntityFilterType:            types.StringValue("entity_types"),
		EntityFilterTypeIdentifiers: []types.String{types.StringValue("service")},
		Published:                   types.BoolValue(true),
		Tags:                        []scorecard.TagModel{{Value: types.StringValue("test")}},
		CheckGroups: map[string]scorecard.CheckGroupModel{
			"security": {
				Name:     types.StringValue("Security"),
				Ordering: types.Int32Value(1),
			},
		},
		Checks: map[string]scorecard.CheckModel{
			"check_x": {
				Name:                   types.StringValue("Check X"),
				ScorecardCheckGroupKey: types.StringValue("security"),
				Ordering:               types.Int32Value(5), // Duplicate ordering!
				Points:                 types.Int32Value(10),
			},
			"check_y": {
				Name:                   types.StringValue("Check Y"),
				ScorecardCheckGroupKey: types.StringValue("security"),
				Ordering:               types.Int32Value(5), // Duplicate ordering!
				Points:                 types.Int32Value(15),
			},
		},
	}

	diags := diag.Diagnostics{}

	scorecard.ValidateModel(model, &diags)

	if !diags.HasError() {
		t.Error("Expected validation to fail due to duplicate ordering, but it passed")
		return
	}

	if len(diags) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(diags))
		return
	}

	expectedMsg := "Check group `security`: the following checks have a duplicate ordering of 5: `check_x`, `check_y`"
	actualMsg := diags[0].Detail()

	if actualMsg != expectedMsg {
		t.Errorf("Expected error message:\n%s\n\nGot:\n%s", expectedMsg, actualMsg)
	}
}

func TestAccDxScorecardResourceCreateScorecard(t *testing.T) {
	scorecardName := fmt.Sprintf("Terraform Provider Scorecard %d", acctest.RandInt())
	// The %[2]s slot carries the deprecated evaluation_frequency_hours attribute, so the
	// same scorecard can be applied both with and without it.
	var testAccDxScorecardResourceBasic = `
provider "dx" {}

resource "dx_scorecard" "level_based_example" {
  name                           = "%[1]s"
  description                    = "This is a test scorecard"
  type                           = "LEVEL"
  entity_filter_type             = "entity_types"
  entity_filter_type_identifiers = ["service"]
%[2]s
  empty_level_label              = "Incomplete"
  empty_level_color              = "#cccccc"
  published                      = true

  tags = [
    { value = "test" },
    { value = "production" },
    { value = "Terraform Acceptance Test" },
  ]

  levels = {
    bronze = {
      name  = "Bronze"
      color = "#FB923C"
      rank  = 1
    },
    silver = {
      name  = "Silver"
      color = "#9CA3AF"
      rank  = 2
    },
    gold = {
      name  = "Gold"
      color = "#FBBF24"
      rank  = 3
    },
  }

  checks = {
    test_check = {
      name                = "Test Check"
      scorecard_level_key = "bronze"
      ordering            = 0

      description    = "This is a test check"
      sql            = <<-EOT
        select 'PASS' as status, 123 as output
      EOT
      output_enabled = true
      output_type    = "custom"
      output_custom_options = {
        unit     = "widget"
        decimals = 0
      }
      output_aggregation = "median"
      external_url       = "http://example.com"
      published          = true
      estimated_dev_days = 1.5
    },

    another_check = {
      name                = "Another Check"
      scorecard_level_key = "bronze"
      ordering            = 1

      sql                = <<-EOT
        with random_number as (
          select ROUND(RANDOM() * 10) as value
        )
        select case
            when value >= 7 then 'PASS'
            when value >= 4 then 'WARN'
            else 'FAIL'
          end as status,
          value as output
        from random_number
      EOT
      output_enabled     = true
      output_type        = "duration_seconds"
      output_aggregation = "median"
      published          = false
      estimated_dev_days = null
    },

    neat_silver_check = {
      name                = "Neat silver check"
      scorecard_level_key = "silver"
      ordering            = 0

      description        = "This is a neat silver check"
      sql                = <<-EOT
        select 'PASS' as status
      EOT
      output_enabled     = false
      published          = false
      estimated_dev_days = 1.5
    },
  }
}

`

	// 24 is deliberate: the API ignores evaluation_frequency_hours and always reports a
	// fixed 2 back, so 24 is a value the API would never return on its own.
	withDeprecatedAttr := fmt.Sprintf(testAccDxScorecardResourceBasic, scorecardName, "  evaluation_frequency_hours     = 24")
	withoutDeprecatedAttr := fmt.Sprintf(testAccDxScorecardResourceBasic, scorecardName, "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// A configuration still carrying the deprecated evaluation_frequency_hours must
			// apply cleanly.
			{
				Config: withDeprecatedAttr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_scorecard.level_based_example", "name", scorecardName),
				),
			},
			// Dropping the attribute is the migration the deprecation message asks users to
			// make, so verify it applies cleanly and leaves nothing behind in state.
			{
				Config: withoutDeprecatedAttr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dx_scorecard.level_based_example", "name", scorecardName),
					resource.TestCheckNoResourceAttr("dx_scorecard.level_based_example", "evaluation_frequency_hours"),
				),
			},
		},
	})
}
