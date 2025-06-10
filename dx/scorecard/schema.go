package scorecard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func LevelSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"key": schema.StringAttribute{Required: true},
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			}},
		"name":  schema.StringAttribute{Required: true},
		"color": schema.StringAttribute{Required: true},
		"rank":  schema.Int32Attribute{Required: true},
	}
}

func CheckGroupSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"key": schema.StringAttribute{Required: true},
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			}},
		"name":     schema.StringAttribute{Required: true},
		"ordering": schema.Int32Attribute{Required: true},
	}
}

func CheckSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name":               schema.StringAttribute{Required: true},
		"description":        schema.StringAttribute{Required: true},
		"ordering":           schema.Int32Attribute{Required: true},
		"sql":                schema.StringAttribute{Required: true},
		"filter_sql":         schema.StringAttribute{Required: true},
		"filter_message":     schema.StringAttribute{Required: true},
		"output_enabled":     schema.BoolAttribute{Required: true},
		"output_type":        schema.StringAttribute{Optional: true},
		"output_aggregation": schema.StringAttribute{Optional: true},
		"output_custom_options": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"unit":     schema.StringAttribute{Required: true, Description: "The unit of the output, e.g. `widget`"},
				"decimals": schema.NumberAttribute{Required: true, Description: "The number of decimals to display, or `auto` for default behavior."},
			},
		},
		"estimated_dev_days": schema.Float32Attribute{Optional: true},
		"external_url":       schema.StringAttribute{Required: true},
		"published":          schema.BoolAttribute{Required: true},

		// Fields for level-based scorecards
		"scorecard_level_key": schema.StringAttribute{Optional: true},

		// Fields for points-based scorecards
		"scorecard_check_group_key": schema.StringAttribute{Optional: true},
		"points":                    schema.Int32Attribute{Optional: true},
	}
}

func ScorecardSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The unique ID of the scorecard.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the scorecard.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "The type of scorecard. Options: 'LEVEL', 'POINTS'.",
			// Validators: []validator.String{
			// 	stringvalidator.OneOf("LEVEL", "POINTS"),
			// },
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"entity_filter_type": schema.StringAttribute{
			Required:    true,
			Description: "The filtering strategy when deciding what entities this scorecard should assess. Options: 'entity_types', 'sql'",
			// Validators: []validator.String{
			// 	stringvalidator.OneOf("entity_types", "sql"),
			// },
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"evaluation_frequency_hours": schema.Int32Attribute{
			Required:    true,
			Description: "How often the scorecard is evaluated (in hours). [2|4|8|24]",
			// Validators: []validator.Number{
			// 	numbervalidator.OneOf(2, 4, 8, 24),
			// },
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
			},
		},

		// Conditionally required for levels-based scorecards
		"empty_level_label": schema.StringAttribute{
			Optional:    true,
			Description: "The label to display when an entity has not achieved any levels in the scorecard (levels scorecards only).",
		},
		"empty_level_color": schema.StringAttribute{
			Optional:    true,
			Description: "The color hex code to display when an entity has not achieved any levels in the scorecard (levels scorecards only).",
		},
		"levels": schema.ListNestedAttribute{
			Optional:    true,
			Description: "The levels that can be achieved in this scorecard (levels scorecards only).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: LevelSchema(),
			},
		},

		// Conditionally required for points-based scorecards
		"check_groups": schema.ListNestedAttribute{
			Optional:    true,
			Description: "Groups of checks, to help organize the scorecard for entity owners (points scorecards only).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: CheckGroupSchema(),
			},
		},

		// Optional metadata
		"description": schema.StringAttribute{
			Optional:    true,
			Description: "Description of the scorecard.",
		},
		"published": schema.BoolAttribute{
			Optional:    true,
			Description: "Whether the scorecard is published.",
		},
		"entity_filter_type_identifiers": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "List of entity type identifiers that the scorecard should run against.",
		},
		"entity_filter_sql": schema.StringAttribute{
			Optional:    true,
			Description: "Custom SQL used to filter entities that the scorecard should run against.",
		},

		// For now, all check field are required. This may change in the future.
		"checks": schema.ListNestedAttribute{
			Optional:    true,
			Description: "List of checks that are applied to entities in the scorecard.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: CheckSchema(),
			},
		},
	}
}

func (r *ScorecardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DX Scorecard.",
		Attributes:  ScorecardSchema(),
	}
}
