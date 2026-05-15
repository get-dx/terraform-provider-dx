package relation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *RelationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DX Catalog Relation definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the relation (same as 'identifier').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier for the relation definition. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Relation type.",
				Validators: []validator.String{
					stringvalidator.OneOf("consumes", "depends on", "parent of", "part of", "provides", "manages"),
				},
			},
			"inverse_type": schema.StringAttribute{
				Computed:    true,
				Description: "The inverse relation type, derived automatically by the API.",
			},
			"cardinality": schema.StringAttribute{
				Required:    true,
				Description: "Cardinality constraint. Cannot be changed after creation.",
				Validators: []validator.String{
					stringvalidator.OneOf("one_to_one", "one_to_many", "many_to_one", "many_to_many"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Human-readable description of the relation.",
			},
			"source_entity_type_identifier": schema.StringAttribute{
				Required:    true,
				Description: "Entity type identifier for the source side of the relation. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_entity_type_identifier": schema.StringAttribute{
				Required:    true,
				Description: "Entity type identifier for the target side of the relation. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the relation was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the relation was last updated.",
			},
		},
	}
}
