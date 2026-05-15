package relation

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RelationModel struct {
	Id                         types.String `tfsdk:"id"`
	Identifier                 types.String `tfsdk:"identifier"`
	Type                       types.String `tfsdk:"type"`
	InverseType                types.String `tfsdk:"inverse_type"`
	Cardinality                types.String `tfsdk:"cardinality"`
	Description                types.String `tfsdk:"description"`
	SourceEntityTypeIdentifier types.String `tfsdk:"source_entity_type_identifier"`
	TargetEntityTypeIdentifier types.String `tfsdk:"target_entity_type_identifier"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	UpdatedAt                  types.String `tfsdk:"updated_at"`
}
