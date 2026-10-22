package relationedge

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RelationEdgeModel struct {
	Id                     types.String `tfsdk:"id"`
	RelationIdentifier     types.String `tfsdk:"relation_identifier"`
	SourceEntityIdentifier types.String `tfsdk:"source_entity_identifier"`
	TargetEntityIdentifier types.String `tfsdk:"target_entity_identifier"`
}
