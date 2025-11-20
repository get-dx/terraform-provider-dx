package entitytype

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PropertySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"identifier": schema.StringAttribute{
			Required:    true,
			Description: "Unique identifier for the property.",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Display name for the property.",
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "Property type (e.g., 'multi_select', 'text').",
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Description: "Description of the property.",
		},
		"visibility": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Property visibility setting. Options: 'hidden', 'visible'. Defaults to 'visible' if not specified.",
			Validators: []validator.String{
				stringvalidator.OneOf("hidden", "visible"),
			},
		},
		"ordering": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Sort order for the property. If not specified, properties will be ordered by their position in the list.",
		},
		"options": schema.ListNestedAttribute{
			Optional:    true,
			Description: "Available options for multi_select properties.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"value": schema.StringAttribute{
						Required:    true,
						Description: "The option value.",
					},
					"color": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Hex color code for the option (e.g., '#ef4444'). Defaults to '#3b82f6' (blue) if not specified.",
					},
				},
			},
		},
	}
}

func EntityTypeSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier of the entity type (same as 'identifier').",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"identifier": schema.StringAttribute{
			Required:    true,
			Description: "User-defined unique identifier for the entity type. This cannot be changed after creation.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Display name for the entity type.",
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Description: "Detailed explanation of the entity type.",
		},
		"properties": schema.ListNestedAttribute{
			Optional:    true,
			Description: "Custom properties to attach to the entity type. Note: When updating, you must include ALL existing properties in your configuration, as the API replaces the entire properties list.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: PropertySchema(),
			},
		},
		"aliases": schema.MapAttribute{
			Optional:    true,
			ElementType: types.BoolType,
			Description: "Key-value pairs enabling specific aliases for the entity type (e.g., 'github_repository': true).",
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: "Timestamp when the entity type was created.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: "Timestamp when the entity type was last updated.",
		},
		"ordering": schema.Int64Attribute{
			Computed:    true,
			Description: "Sort order for the entity type.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
	}
}

func (r *EntityTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DX Entity Type. Entity types are used to define the data model for entities in a software catalog.",
		Attributes:  EntityTypeSchema(),
	}
}
