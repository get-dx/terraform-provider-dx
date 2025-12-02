package entity

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EntityModel describes the resource data model.
type EntityModel struct {
	// Required fields
	Id         types.String `tfsdk:"id"`         // Same as identifier for Terraform conventions
	Identifier types.String `tfsdk:"identifier"` // User-defined unique identifier
	Type       types.String `tfsdk:"type"`       // Entity type identifier

	// Optional fields
	Name         types.String              `tfsdk:"name"`           // Display name
	Description  types.String              `tfsdk:"description"`    // Entity description
	OwnerTeamIds []types.String            `tfsdk:"owner_team_ids"` // Array of owner team IDs
	OwnerUserIds []types.String            `tfsdk:"owner_user_ids"` // Array of owner user IDs
	Domain       types.String              `tfsdk:"domain"`         // Domain entity identifier
	Properties   types.Dynamic             `tfsdk:"properties"`     // Entity properties (key-value pairs, values can be strings, numbers, null, objects, or lists)
	Aliases      map[string][]AliasModel   `tfsdk:"aliases"`        // Aliases map (map of alias type to array of alias objects)
	Relations    map[string][]types.String `tfsdk:"relations"`      // Relations map

	// Computed fields (from API)
	CreatedAt types.String `tfsdk:"created_at"` // Creation timestamp
	UpdatedAt types.String `tfsdk:"updated_at"` // Last update timestamp
}

// AliasModel describes an alias entry for an entity.
type AliasModel struct {
	Identifier types.String `tfsdk:"identifier"` // Required: the alias identifier
}
