package entity

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AliasSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"identifier": schema.StringAttribute{
			Required:    true,
			Description: "The unique identifier for the alias (e.g., GitHub repository ID).",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "The name of the alias (computed from the Data Cloud database).",
		},
		"url": schema.StringAttribute{
			Computed:    true,
			Description: "The URL of the alias (computed from the Data Cloud database).",
		},
	}
}

func EntitySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier of the entity (same as 'identifier').",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"identifier": schema.StringAttribute{
			Required:    true,
			Description: "User-defined unique identifier for the entity. This cannot be changed after creation.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "The identifier of the entity type (e.g., 'service', 'api', 'domain').",
		},
		"name": schema.StringAttribute{
			Optional:    true,
			Description: "Display name for the entity.",
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Description: "Description of the entity.",
		},
		"owner_team_ids": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Array of owner team IDs assigned to the entity.",
		},
		"owner_user_ids": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Array of owner user IDs assigned to the entity.",
		},
		"domain": schema.StringAttribute{
			Optional:    true,
			Description: "The identifier of the domain entity parent assigned to the entity.",
		},
		"properties": schema.DynamicAttribute{
			Optional:    true,
			Description: "Key-value pairs of entity properties and their values. Values can be strings, numbers, null, objects, or lists of any of those types.",
		},
		"aliases": schema.MapAttribute{
			ElementType: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"identifier": types.StringType,
					},
				},
			},
			Optional:    true,
			Description: "Key-value pairs of aliases assigned to the entity. Keys are alias types (e.g., 'github_repo'), values are arrays of alias objects with 'identifier' (required) field.",
		},
		"relations": schema.MapAttribute{
			ElementType: types.ListType{ElemType: types.StringType},
			Optional:    true,
			Description: "Key-value pairs of relations and their associated entity identifiers. Keys are relation types (e.g., 'service-consumes-api'), values are arrays of entity identifiers.",
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: "Timestamp when the entity was created.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: "Timestamp when the entity was last updated.",
		},
	}
}

func (r *EntityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DX Entity. Entities represent items in your software catalog (e.g., services, APIs, domains).",
		Attributes:  EntitySchema(),
	}
}
