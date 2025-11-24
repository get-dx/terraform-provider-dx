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
	Description types.String             `tfsdk:"description"` // Entity type description
	Properties  map[string]PropertyModel `tfsdk:"properties"`  // Custom properties, keyed by identifier
	Aliases     map[string]types.Bool    `tfsdk:"aliases"`     // Alias type mappings

	// Computed fields (from API)
	CreatedAt types.String `tfsdk:"created_at"` // Creation timestamp
	UpdatedAt types.String `tfsdk:"updated_at"` // Last update timestamp
	Ordering  types.Int64  `tfsdk:"ordering"`   // Sort order
}

// PropertyModel describes a custom property on an entity type.
// Note: The identifier is the map key, not a field in this struct.
type PropertyModel struct {
	Name             types.String          `tfsdk:"name"`                // Required: display name
	Type             types.String          `tfsdk:"type"`                // Required: property type (e.g., "multi_select", "text", "computed", "url")
	Description      types.String          `tfsdk:"description"`         // Optional: property description
	Visibility       types.String          `tfsdk:"visibility"`          // Optional: property visibility
	Ordering         types.Int64           `tfsdk:"ordering"`            // Optional: sort order for the property
	Options          []PropertyOptionModel `tfsdk:"options"`             // Optional: for multi_select type
	SQL              types.String          `tfsdk:"sql"`                 // Optional: SQL query for computed type
	CallToAction     types.String          `tfsdk:"call_to_action"`      // Optional: call-to-action text for url type
	CallToActionType types.String          `tfsdk:"call_to_action_type"` // Optional: call-to-action type for url type
}

// PropertyOptionModel describes an option for a multi_select property.
type PropertyOptionModel struct {
	Value types.String `tfsdk:"value"` // Required: the option value
	Color types.String `tfsdk:"color"` // Required: hex color code for the option
}
