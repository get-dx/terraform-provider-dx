package entitytype

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EntityTypeModel describes the resource data model.
type EntityTypeModel struct {
	// Required fields
	Id         types.String `tfsdk:"id"`         // Same as identifier for Terraform conventions
	Identifier types.String `tfsdk:"identifier"` // User-defined unique identifier
	Name       types.String `tfsdk:"name"`       // Display name

	// Optional fields
	Description types.String          `tfsdk:"description"` // Entity type description
	Properties  []PropertyModel       `tfsdk:"properties"`  // Custom properties
	Aliases     map[string]types.Bool `tfsdk:"aliases"`     // Alias type mappings

	// Computed fields (from API)
	CreatedAt types.String `tfsdk:"created_at"` // Creation timestamp
	UpdatedAt types.String `tfsdk:"updated_at"` // Last update timestamp
	Ordering  types.Int64  `tfsdk:"ordering"`   // Sort order
}

// PropertyModel describes a custom property on an entity type.
type PropertyModel struct {
	Identifier  types.String   `tfsdk:"identifier"`  // Required: unique property identifier
	Name        types.String   `tfsdk:"name"`        // Required: display name
	Type        types.String   `tfsdk:"type"`        // Required: property type (e.g., "multi_select", "text")
	Description types.String   `tfsdk:"description"` // Optional: property description
	Visibility  types.String   `tfsdk:"visibility"`  // Optional: property visibility
	Ordering    types.Int64    `tfsdk:"ordering"`    // Optional: sort order for the property
	Options     []types.String `tfsdk:"options"`     // Optional: for multi_select type (shorthand for definition.options)
}
