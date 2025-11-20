package entitytype

import (
	"context"
	"fmt"

	"terraform-provider-dx/dx"
	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EntityTypeResource{}
	_ resource.ResourceWithImportState = &EntityTypeResource{}
)

func NewEntityTypeResource() resource.Resource {
	return &EntityTypeResource{}
}

// EntityTypeResource defines the resource implementation.
type EntityTypeResource struct {
	client *dxapi.Client
}

func (r *EntityTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity_type"
}

func (r *EntityTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dxapi.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *dxapi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The API client was not configured. This is a bug in the provider.")
		return
	}
}

func (r *EntityTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating entity type resource")

	// Retrieve values from plan
	var plan EntityTypeModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Debug(ctx, "Plan has errors, returning early")
		return
	}

	tflog.Debug(ctx, "Got plan, validating...")
	ValidateModel(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := modelToRequestBody(ctx, plan, false)
	if err != nil {
		resp.Diagnostics.AddError("Error converting plan to request body", err.Error())
		return
	}

	// Create EntityType (apiResp is a struct of type APIEntityTypeResponse)
	apiResp, err := r.client.CreateEntityType(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating entity type", err.Error())
		return
	}

	// Keep a copy of the original plan to preserve null values
	oldPlan := plan
	responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EntityTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading entity type resource")

	var state EntityTypeModel

	// Load existing state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Prior state, before reading from API: %v", state))

	// Extract identifier
	identifier := state.Identifier.ValueString()
	if identifier == "" {
		resp.Diagnostics.AddError("Missing identifier", "The resource identifier is missing from the state")
		return
	}

	// Call the API to get the latest entity type data
	apiResp, err := r.client.GetEntityType(ctx, identifier)
	if err != nil {
		// TODO - implement resource not found error handling
		// if isNotFoundError(err) {
		// 	// Resource no longer exists remotely — remove from state
		// 	resp.State.RemoveResource(ctx)
		// 	return
		// }
		resp.Diagnostics.AddError(
			"Error reading entity type",
			fmt.Sprintf("Could not read entity type with identifier %s: %s", identifier, err.Error()),
		)
		return
	}

	// Map API response to Terraform state model
	// Keep a copy of the original state to preserve null values
	oldState := state
	responseBodyToModel(ctx, apiResp, &state, &oldState)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *EntityTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EntityTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) // Get the desired state
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Got plan, validating...")
	ValidateModel(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := modelToRequestBody(ctx, plan, true)
	if err != nil {
		resp.Diagnostics.AddError("Error converting plan to request body", err.Error())
		return
	}

	apiResp, err := r.client.UpdateEntityType(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating entity type", err.Error())
		return
	}

	// Map API response to Terraform state model
	// Keep a copy of the original plan to preserve null values
	oldPlan := plan
	responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EntityTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EntityTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get the current state
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()
	if identifier == "" {
		resp.Diagnostics.AddError("Missing identifier", "The resource identifier is missing from the state")
		return
	}

	success, err := r.client.DeleteEntityType(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting entity type", err.Error())
		return
	}
	if !success {
		resp.Diagnostics.AddError("Error deleting entity type", "API did not confirm deletion.")
		return
	}
	// No need to set state, resource will be removed by Terraform if this method returns successfully
}

func (r *EntityTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Importing entity type state")

	// Use the identifier as the import ID
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func ValidateModel(plan EntityTypeModel, diags *diag.Diagnostics) {
	// Validate required fields
	if plan.Identifier.IsNull() || plan.Identifier.IsUnknown() {
		diags.AddError("Missing required field", "The 'identifier' field must be specified.")
		return
	}
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		diags.AddError("Missing required field", "The 'name' field must be specified.")
		return
	}

	// Validate properties
	if len(plan.Properties) > 0 {
		propertyIdentifiers := make(map[string]bool)
		for i, prop := range plan.Properties {
			// Check required property fields
			if prop.Identifier.IsNull() || prop.Identifier.IsUnknown() {
				diags.AddError("Missing required field", fmt.Sprintf("Property at index %d is missing required field 'identifier'.", i))
			}
			if prop.Name.IsNull() || prop.Name.IsUnknown() {
				diags.AddError("Missing required field", fmt.Sprintf("Property at index %d is missing required field 'name'.", i))
			}
			if prop.Type.IsNull() || prop.Type.IsUnknown() {
				diags.AddError("Missing required field", fmt.Sprintf("Property at index %d is missing required field 'type'.", i))
			}

			// Check for duplicate property identifiers
			if !prop.Identifier.IsNull() {
				propIdentifier := prop.Identifier.ValueString()
				if propertyIdentifiers[propIdentifier] {
					diags.AddError("Duplicate property identifier", fmt.Sprintf("Property identifier '%s' is used more than once. Each property must have a unique identifier.", propIdentifier))
				}
				propertyIdentifiers[propIdentifier] = true
			}
		}
	}
}

func modelToRequestBody(ctx context.Context, plan EntityTypeModel, isUpdate bool) (map[string]interface{}, error) {
	tflog.Debug(ctx, "Converting plan to request body")

	// Construct API request payload
	payload := map[string]interface{}{
		"identifier": plan.Identifier.ValueString(),
		"name":       plan.Name.ValueString(),
	}

	// Add optional fields if they're present
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	// Add properties array
	if len(plan.Properties) > 0 {
		properties := []map[string]interface{}{}
		for idx, planProp := range plan.Properties {
			property := map[string]interface{}{
				"identifier": planProp.Identifier.ValueString(),
				"name":       planProp.Name.ValueString(),
				"type":       planProp.Type.ValueString(),
			}

			// Add optional property fields
			if !planProp.Description.IsNull() && !planProp.Description.IsUnknown() {
				property["description"] = planProp.Description.ValueString()
			}

			if !planProp.Visibility.IsNull() && !planProp.Visibility.IsUnknown() {
				property["visibility"] = planProp.Visibility.ValueString()
			} else {
				// Default to "visible" if not specified
				property["visibility"] = "visible"
			}

			if !planProp.Ordering.IsNull() && !planProp.Ordering.IsUnknown() {
				property["ordering"] = planProp.Ordering.ValueInt64()
			} else {
				// Default to index if not specified
				property["ordering"] = int64(idx)
			}

			// Build definition object based on property type
			propType := planProp.Type.ValueString()
			if propType == "multi_select" {
				// For multi_select, create definition with options
				definition := map[string]interface{}{}

				if len(planProp.Options) > 0 {
					options := make([]map[string]interface{}, 0, len(planProp.Options))
					for _, opt := range planProp.Options {
						color := "#3b82f6" // Default blue color
						if !opt.Color.IsNull() && !opt.Color.IsUnknown() {
							color = opt.Color.ValueString()
						}
						options = append(options, map[string]interface{}{
							"value": opt.Value.ValueString(),
							"color": color,
						})
					}
					definition["options"] = options
				} else {
					definition["options"] = []map[string]interface{}{}
				}

				property["definition"] = definition
			} else {
				// For other types (like text), definition is an empty object
				property["definition"] = map[string]interface{}{}
			}

			properties = append(properties, property)
		}
		payload["properties"] = properties
	} else {
		// Even if no properties in plan, send empty array for updates
		// to ensure API replaces with empty list
		if isUpdate {
			payload["properties"] = []map[string]interface{}{}
		}
	}

	// Add aliases map
	if plan.Aliases != nil && len(plan.Aliases) > 0 {
		aliases := make(map[string]bool)
		for key, val := range plan.Aliases {
			if !val.IsNull() && !val.IsUnknown() {
				aliases[key] = val.ValueBool()
			}
		}
		payload["aliases"] = aliases
	}

	return payload, nil
}

func responseBodyToModel(ctx context.Context, apiResp *dxapi.APIEntityTypeResponse, state *EntityTypeModel, oldPlan *EntityTypeModel) {
	tflog.Debug(ctx, "Mapping API response to Terraform model")

	// Required fields
	state.Id = types.StringValue(apiResp.EntityType.Identifier)         // ID is same as identifier
	state.Identifier = types.StringValue(apiResp.EntityType.Identifier) // User-facing identifier
	state.Name = types.StringValue(apiResp.EntityType.Name)

	// Optional fields
	state.Description = dx.StringOrNull(apiResp.EntityType.Description)

	// Computed fields
	state.CreatedAt = types.StringValue(apiResp.EntityType.CreatedAt)
	state.UpdatedAt = types.StringValue(apiResp.EntityType.UpdatedAt)
	state.Ordering = types.Int64Value(apiResp.EntityType.Ordering)

	// Properties array
	// Only set properties if they were originally specified (not null) in the plan,
	// or if the API returned non-empty properties
	if len(apiResp.EntityType.Properties) > 0 {
		properties := make([]PropertyModel, 0, len(apiResp.EntityType.Properties))
		for _, apiProp := range apiResp.EntityType.Properties {
			property := PropertyModel{
				Identifier:  types.StringValue(apiProp.Identifier),
				Name:        types.StringValue(apiProp.Name),
				Type:        types.StringValue(apiProp.Type),
				Description: dx.StringOrNull(apiProp.Description),
				Visibility:  dx.StringOrNull(apiProp.Visibility),
				Ordering:    dx.Int64OrNull(apiProp.Ordering),
			}

			// Extract options from definition if present
			if apiProp.Definition != nil && len(apiProp.Definition.Options) > 0 {
				options := make([]PropertyOptionModel, 0, len(apiProp.Definition.Options))
				for _, opt := range apiProp.Definition.Options {
					options = append(options, PropertyOptionModel{
						Value: types.StringValue(opt.Value),
						Color: types.StringValue(opt.Color),
					})
				}
				property.Options = options
			}

			properties = append(properties, property)
		}
		state.Properties = properties
	} else {
		// If the API returned no properties and the original plan had no properties (was null),
		// keep it as nil to maintain consistency
		if oldPlan.Properties == nil {
			state.Properties = nil
		} else {
			state.Properties = []PropertyModel{}
		}
	}

	// Aliases map
	// Only set aliases if they were originally specified (not null) in the plan,
	// or if the API returned non-empty aliases
	if apiResp.EntityType.Aliases != nil && len(apiResp.EntityType.Aliases) > 0 {
		aliases := make(map[string]types.Bool)
		for key, val := range apiResp.EntityType.Aliases {
			aliases[key] = types.BoolValue(val)
		}
		state.Aliases = aliases
	} else {
		// If the API returned no aliases and the original plan had no aliases (was null),
		// keep it as nil to maintain consistency
		if oldPlan.Aliases == nil {
			state.Aliases = nil
		} else {
			state.Aliases = map[string]types.Bool{}
		}
	}
}
