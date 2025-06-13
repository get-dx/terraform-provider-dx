package scorecard

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScorecardModel describes the resource data model.
type ScorecardModel struct {
	// Required fields
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	EntityFilterType    types.String `tfsdk:"entity_filter_type"`
	EvaluationFrequency types.Int32  `tfsdk:"evaluation_frequency_hours"`

	// Conditionally required fields for levels based scorecards
	EmptyLevelLabel types.String          `tfsdk:"empty_level_label"`
	EmptyLevelColor types.String          `tfsdk:"empty_level_color"`
	Levels          map[string]LevelModel `tfsdk:"levels"`

	// Conditionally required fields for points based scorecards
	CheckGroups map[string]CheckGroupModel `tfsdk:"check_groups"`

	// Optional fields
	Description                 types.String          `tfsdk:"description"`
	Published                   types.Bool            `tfsdk:"published"`
	EntityFilterTypeIdentifiers []types.String        `tfsdk:"entity_filter_type_identifiers"`
	EntityFilterSql             types.String          `tfsdk:"entity_filter_sql"`
	Checks                      map[string]CheckModel `tfsdk:"checks"`
}

type LevelModel struct {
	Id    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
	Rank  types.Int32  `tfsdk:"rank"`
}

type CheckGroupModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Ordering types.Int32  `tfsdk:"ordering"`
}

type CheckModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Ordering      types.Int32  `tfsdk:"ordering"`
	Sql           types.String `tfsdk:"sql"`
	FilterSql     types.String `tfsdk:"filter_sql"`
	FilterMessage types.String `tfsdk:"filter_message"`
	OutputEnabled types.Bool   `tfsdk:"output_enabled"`

	OutputType          types.String              `tfsdk:"output_type"`
	OutputAggregation   types.String              `tfsdk:"output_aggregation"`
	OutputCustomOptions *OutputCustomOptionsModel `tfsdk:"output_custom_options"`

	EstimatedDevDays types.Float32 `tfsdk:"estimated_dev_days"`
	ExternalUrl      types.String  `tfsdk:"external_url"`
	Published        types.Bool    `tfsdk:"published"`

	// Additional fields for level based scorecards
	ScorecardLevelKey types.String `tfsdk:"scorecard_level_key"`

	// Additional fields for points based scorecards
	ScorecardCheckGroupKey types.String `tfsdk:"scorecard_check_group_key"`
	Points                 types.Int32  `tfsdk:"points"`
}

type OutputCustomOptionsModel struct {
	Unit     types.String `tfsdk:"unit"`
	Decimals types.Int32  `tfsdk:"decimals"`
}
