package entity

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-dx/dx"
	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EntityResource{}
	_ resource.ResourceWithImportState = &EntityResource{}
)

func NewEntityResource() resource.Resource {
	return &EntityResource{}
}

// EntityResource defines the resource implementation.
type EntityResource struct {
	client *dxapi.Client
}

func (r *EntityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (r *EntityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EntityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating entity resource")

	// Check null states before decoding (Go maps/slices lose null vs empty distinction)
	nullStates := checkNullStates(ctx, req.Plan)

	// Retrieve values from plan
	var plan EntityResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Debug(ctx, "Plan has errors, returning early")
		return
	}

	payload := modelToRequestBody(ctx, plan)

	// Create Entity (apiResp is a struct of type APIEntityResponse)
	apiResp, err := r.client.CreateEntity(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating entity", err.Error())
		return
	}

	// Keep a copy of the original plan to preserve null values
	oldPlan := plan
	responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

	// Restore null states for fields that were null in the plan
	restoreNullStates(&plan, nullStates)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EntityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading entity resource")

	// Check null states before decoding (Go maps/slices lose null vs empty distinction)
	nullStates := checkNullStates(ctx, req.State)

	var state EntityResourceModel

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

	// Call the API to get the latest entity data
	apiResp, err := r.client.GetEntity(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading entity",
			fmt.Sprintf("Could not read entity with identifier %s: %s", identifier, err.Error()),
		)
		return
	}

	// Map API response to Terraform state model
	// Keep a copy of the original state to preserve null values
	oldState := state
	responseBodyToModel(ctx, apiResp, &state, &oldState)

	// Restore null states for fields that were null in the prior state
	restoreNullStates(&state, nullStates)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *EntityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Check null states before decoding (Go maps/slices lose null vs empty distinction)
	nullStates := checkNullStates(ctx, req.Plan)

	var plan EntityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) // Get the desired state
	if resp.Diagnostics.HasError() {
		return
	}

	// Get prior state to detect removed aliases
	var priorState EntityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := modelToRequestBody(ctx, plan)

	// Add empty arrays for removed relation types so the API removes them
	addRemovedMapKeys(ctx, payload, priorState, plan)

	apiResp, err := r.client.UpdateEntity(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating entity", err.Error())
		return
	}

	// Map API response to Terraform state model
	// Keep a copy of the original plan to preserve null values
	oldPlan := plan
	responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

	// Restore null states for fields that were null in the plan
	restoreNullStates(&plan, nullStates)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EntityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EntityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get the current state
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()
	if identifier == "" {
		resp.Diagnostics.AddError("Missing identifier", "The resource identifier is missing from the state")
		return
	}

	success, err := r.client.DeleteEntity(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting entity", err.Error())
		return
	}
	if !success {
		resp.Diagnostics.AddError("Error deleting entity", "API did not confirm deletion.")
		return
	}
	// No need to set state, resource will be removed by Terraform if this method returns successfully
}

func (r *EntityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Importing entity state")

	// Use the identifier as the import ID
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// nullFieldStates tracks which optional map/list fields were null in the plan.
// This is needed because Go maps/slices lose the null vs empty distinction when decoded.
type nullFieldStates struct {
	AliasesNull      bool
	OwnerTeamIdsNull bool
	OwnerUserIdsNull bool
}

// planGetter is an interface for getting attributes from a plan or state.
type planGetter interface {
	GetAttribute(ctx context.Context, p path.Path, target interface{}) diag.Diagnostics
}

// checkNullStates checks which optional map/list fields are null in the plan before decoding.
func checkNullStates(ctx context.Context, plan planGetter) nullFieldStates {
	var states nullFieldStates

	var aliasesAttr types.Map
	plan.GetAttribute(ctx, path.Root("aliases"), &aliasesAttr)
	states.AliasesNull = aliasesAttr.IsNull()

	var ownerTeamIdsAttr types.List
	plan.GetAttribute(ctx, path.Root("owner_team_ids"), &ownerTeamIdsAttr)
	states.OwnerTeamIdsNull = ownerTeamIdsAttr.IsNull()

	var ownerUserIdsAttr types.List
	plan.GetAttribute(ctx, path.Root("owner_user_ids"), &ownerUserIdsAttr)
	states.OwnerUserIdsNull = ownerUserIdsAttr.IsNull()

	return states
}

// restoreNullStates sets fields back to nil if they were null in the original plan/state
// AND the API returned empty data. If the API returned actual data, we keep it
// (this is important for import where prior state is empty but API has data).
func restoreNullStates(model *EntityResourceModel, states nullFieldStates) {
	if states.AliasesNull && len(model.Aliases) == 0 {
		model.Aliases = nil
	}
	if states.OwnerTeamIdsNull && len(model.OwnerTeamIds) == 0 {
		model.OwnerTeamIds = nil
	}
	if states.OwnerUserIdsNull && len(model.OwnerUserIds) == 0 {
		model.OwnerUserIds = nil
	}
}

// addRemovedMapKeys adds empty arrays for alias types that existed in
// the prior state but are not in the new plan. This tells the API to remove them.
// It also adds null values for removed property keys.
func addRemovedMapKeys(_ context.Context, payload map[string]interface{}, priorState EntityResourceModel, plan EntityResourceModel) {
	// Handle removed alias types
	if len(priorState.Aliases) > 0 {
		// Get or create the aliases map in payload
		aliases, ok := payload["aliases"].(map[string][]dxapi.APIAlias)
		if !ok {
			aliases = make(map[string][]dxapi.APIAlias)
		}

		// Check each alias type from prior state
		for aliasType := range priorState.Aliases {
			// If this alias type is not in the new plan, add an empty array
			if _, existsInPlan := plan.Aliases[aliasType]; !existsInPlan {
				aliases[aliasType] = []dxapi.APIAlias{}
			}
		}

		payload["aliases"] = aliases
	}

	// Handle removed property keys
	// Convert prior state and plan properties to Go maps to compare keys
	var priorProps map[string]interface{}
	var planProps map[string]interface{}

	if !priorState.Properties.IsNull() && !priorState.Properties.IsUnknown() {
		if underlyingValue := priorState.Properties.UnderlyingValue(); underlyingValue != nil {
			if goValue, err := attrValueToGoValue(underlyingValue); err == nil {
				if m, ok := goValue.(map[string]interface{}); ok {
					priorProps = m
				}
			}
		}
	}

	if !plan.Properties.IsNull() && !plan.Properties.IsUnknown() {
		if underlyingValue := plan.Properties.UnderlyingValue(); underlyingValue != nil {
			if goValue, err := attrValueToGoValue(underlyingValue); err == nil {
				if m, ok := goValue.(map[string]interface{}); ok {
					planProps = m
				}
			}
		}
	}

	if len(priorProps) > 0 {
		// Get or create the properties map in payload
		properties, ok := payload["properties"].(map[string]interface{})
		if !ok {
			properties = make(map[string]interface{})
		}

		// Check each property key from prior state
		for propKey := range priorProps {
			// If this property key is not in the new plan, add null to remove it
			if _, existsInPlan := planProps[propKey]; !existsInPlan {
				properties[propKey] = nil
			}
		}

		payload["properties"] = properties
	}
}

func modelToRequestBody(ctx context.Context, plan EntityResourceModel) map[string]interface{} {
	tflog.Info(ctx, "Converting plan to request body")

	// Ensure identifier is set (should be set by plan modifier, but check just in case)
	identifier := plan.Identifier.ValueString()
	if identifier == "" {
		tflog.Warn(ctx, "Identifier is empty, this should have been set by the plan modifier")
		// This shouldn't happen, but if it does, we'll let the API return an error
	}

	// Construct API request payload with defaults
	payload := map[string]interface{}{
		"identifier":     identifier,
		"type":           plan.Type.ValueString(),
		"name":           nil,
		"description":    nil,
		"owner_team_ids": []string{},
		"owner_user_ids": []string{},
		"domain":         nil,
		"properties":     map[string]interface{}{},
		"aliases":        map[string][]dxapi.APIAlias{},
	}

	// Add optional fields
	addNameToPayload(payload, plan)
	addDescriptionToPayload(payload, plan)
	addOwnerTeamIdsToPayload(payload, plan)
	addOwnerUserIdsToPayload(payload, plan)
	addDomainToPayload(payload, plan)
	addPropertiesToPayload(ctx, payload, plan)
	addAliasesToPayload(payload, plan)

	return payload
}

// addNameToPayload adds the name field to the payload if set.
func addNameToPayload(payload map[string]interface{}, plan EntityResourceModel) {
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		payload["name"] = plan.Name.ValueString()
	}
}

// addDescriptionToPayload adds the description field to the payload if set.
func addDescriptionToPayload(payload map[string]interface{}, plan EntityResourceModel) {
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
}

// addOwnerTeamIdsToPayload adds the owner_team_ids array to the payload if set.
func addOwnerTeamIdsToPayload(payload map[string]interface{}, plan EntityResourceModel) {
	if len(plan.OwnerTeamIds) > 0 {
		teamIds := make([]string, 0, len(plan.OwnerTeamIds))
		for _, teamId := range plan.OwnerTeamIds {
			if !teamId.IsNull() && !teamId.IsUnknown() {
				teamIds = append(teamIds, teamId.ValueString())
			}
		}
		if len(teamIds) > 0 {
			payload["owner_team_ids"] = teamIds
		}
	}
}

// addOwnerUserIdsToPayload adds the owner_user_ids array to the payload if set.
func addOwnerUserIdsToPayload(payload map[string]interface{}, plan EntityResourceModel) {
	if len(plan.OwnerUserIds) > 0 {
		userIds := make([]string, 0, len(plan.OwnerUserIds))
		for _, userId := range plan.OwnerUserIds {
			if !userId.IsNull() && !userId.IsUnknown() {
				userIds = append(userIds, userId.ValueString())
			}
		}
		if len(userIds) > 0 {
			payload["owner_user_ids"] = userIds
		}
	}
}

// addDomainToPayload adds the domain field to the payload if set.
func addDomainToPayload(payload map[string]interface{}, plan EntityResourceModel) {
	if !plan.Domain.IsNull() && !plan.Domain.IsUnknown() {
		payload["domain"] = plan.Domain.ValueString()
	}
}

// addPropertiesToPayload converts and adds properties from types.Dynamic to the payload.
func addPropertiesToPayload(ctx context.Context, payload map[string]interface{}, plan EntityResourceModel) {
	if !plan.Properties.IsNull() && !plan.Properties.IsUnknown() {
		underlyingValue := plan.Properties.UnderlyingValue()
		if underlyingValue != nil {
			goValue, err := attrValueToGoValue(underlyingValue)

			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Failed to convert properties to Go value: %v", err))
			}

			payload["properties"] = goValue
		}
	}
}

// addAliasesToPayload converts and adds aliases from map[string][]AliasModel to the payload.
func addAliasesToPayload(payload map[string]interface{}, plan EntityResourceModel) {
	if len(plan.Aliases) > 0 {
		aliases := make(map[string][]dxapi.APIAlias)
		for aliasType, aliasArray := range plan.Aliases {
			apiAliasArray := make([]dxapi.APIAlias, 0, len(aliasArray))
			for _, aliasModel := range aliasArray {
				if !aliasModel.Identifier.IsNull() && !aliasModel.Identifier.IsUnknown() {
					alias := dxapi.APIAlias{
						Identifier: aliasModel.Identifier.ValueString(),
					}
					apiAliasArray = append(apiAliasArray, alias)
				}
			}
			if len(apiAliasArray) > 0 {
				aliases[aliasType] = apiAliasArray
			}
		}
		if len(aliases) > 0 {
			payload["aliases"] = aliases
		}
	}
}

func responseBodyToModel(ctx context.Context, apiResp *dxapi.APIEntityResponse, state *EntityResourceModel, oldPlan *EntityResourceModel) {
	tflog.Debug(ctx, "Mapping API response to Terraform model")

	// Required fields
	state.Id = types.StringValue(apiResp.Entity.Identifier)         // ID is same as identifier
	state.Identifier = types.StringValue(apiResp.Entity.Identifier) // User-facing identifier
	state.Type = types.StringValue(apiResp.Entity.Type)

	// Optional fields
	state.Name = dx.StringOrNull(apiResp.Entity.Name)
	state.Description = dx.StringOrNull(apiResp.Entity.Description)

	// Owner teams
	if len(apiResp.Entity.OwnerTeams) > 0 {
		teamIds := make([]types.String, 0, len(apiResp.Entity.OwnerTeams))
		for _, team := range apiResp.Entity.OwnerTeams {
			teamIds = append(teamIds, types.StringValue(team.Id))
		}
		state.OwnerTeamIds = teamIds
	} else {
		// API returned no owner teams - preserve exactly what was in the plan
		state.OwnerTeamIds = oldPlan.OwnerTeamIds
	}

	// Owner users
	if len(apiResp.Entity.OwnerUsers) > 0 {
		userIds := make([]types.String, 0, len(apiResp.Entity.OwnerUsers))
		for _, user := range apiResp.Entity.OwnerUsers {
			userIds = append(userIds, types.StringValue(user.Id))
		}
		state.OwnerUserIds = userIds
	} else {
		// API returned no owner users - preserve exactly what was in the plan
		state.OwnerUserIds = oldPlan.OwnerUserIds
	}

	// Domain
	if apiResp.Entity.Domain != nil {
		state.Domain = types.StringValue(apiResp.Entity.Domain.Identifier)
	} else {
		state.Domain = types.StringNull()
	}

	// Properties - convert from API response to types.Dynamic without validation
	// Use JSON round-trip to convert to Dynamic value
	if len(apiResp.Entity.Properties) > 0 {
		jsonBytes, err := json.Marshal(apiResp.Entity.Properties)
		if err == nil {
			var normalized interface{}
			if err := json.Unmarshal(jsonBytes, &normalized); err == nil {
				dynamicVal, err := interfaceToDynamic(ctx, normalized)
				if err == nil {
					state.Properties = dynamicVal
				} else {
					// If conversion fails, preserve the plan value
					state.Properties = oldPlan.Properties
				}
			} else {
				state.Properties = oldPlan.Properties
			}
		} else {
			state.Properties = oldPlan.Properties
		}
	} else {
		// API returned empty properties - preserve plan if it had properties, otherwise set to empty map
		if !oldPlan.Properties.IsNull() && !oldPlan.Properties.IsUnknown() {
			// Plan had properties, so create an empty map to match
			emptyMap := map[string]interface{}{}
			dynamicVal, err := interfaceToDynamic(ctx, emptyMap)
			if err == nil {
				state.Properties = dynamicVal
			} else {
				// If conversion fails, preserve the plan value
				state.Properties = oldPlan.Properties
			}
		} else {
			// Plan was null, so set to null
			state.Properties = types.DynamicNull()
		}
	}

	// Aliases map
	// Convert from map[string][]APIAlias to map[string][]AliasModel
	if len(apiResp.Entity.Aliases) > 0 {
		aliases := make(map[string][]AliasModel)
		for aliasType, aliasArray := range apiResp.Entity.Aliases {
			aliasModels := make([]AliasModel, 0, len(aliasArray))
			for _, alias := range aliasArray {
				aliasModel := AliasModel{
					Identifier: types.StringValue(alias.Identifier),
				}
				aliasModels = append(aliasModels, aliasModel)
			}
			if len(aliasModels) > 0 {
				aliases[aliasType] = aliasModels
			}
		}
		state.Aliases = aliases
	} else {
		// API returned no aliases - preserve exactly what was in the plan
		// This maintains the null vs empty map distinction
		state.Aliases = oldPlan.Aliases
	}

	// Computed fields
	state.CreatedAt = types.StringValue(apiResp.Entity.CreatedAt)
	state.UpdatedAt = types.StringValue(apiResp.Entity.UpdatedAt)
}

// interfaceToDynamic converts an interface{} value to a types.Dynamic value.
func interfaceToDynamic(ctx context.Context, val interface{}) (types.Dynamic, error) {
	if val == nil {
		return types.DynamicNull(), nil
	}

	// Use JSON marshaling/unmarshaling to normalize the structure
	// This ensures nested maps and lists are properly handled
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return types.DynamicNull(), fmt.Errorf("error marshaling value to JSON: %w", err)
	}

	var normalizedVal interface{}
	if err := json.Unmarshal(jsonBytes, &normalizedVal); err != nil {
		return types.DynamicNull(), fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	// Convert the normalized value to tftypes.Value
	// For maps, we need to create them as map types, not wrap in DynamicPseudoType
	tfValue, err := goValueToTerraformValueForDynamic(normalizedVal)
	if err != nil {
		return types.DynamicNull(), fmt.Errorf("error converting Go value to tftypes.Value: %w", err)
	}

	// Create an attr.Value from tftypes.Value
	attrVal, err := valueFromTerraform(ctx, tfValue)
	if err != nil {
		return types.DynamicNull(), fmt.Errorf("error creating attr.Value: %w", err)
	}

	// Create Dynamic value from attr.Value
	dynamicVal := basetypes.NewDynamicValue(attrVal)

	return dynamicVal, nil
}

// goValueToTerraformValueForDynamic converts a Go value to tftypes.Value for use in Dynamic types.
// This version creates proper map types for the top-level map (not wrapped in DynamicPseudoType).
func goValueToTerraformValueForDynamic(val interface{}) (tftypes.Value, error) {
	switch v := val.(type) {
	case nil:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil), nil
	case string:
		return tftypes.NewValue(tftypes.DynamicPseudoType, v), nil
	case bool:
		return tftypes.NewValue(tftypes.DynamicPseudoType, v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return tftypes.NewValue(tftypes.DynamicPseudoType, v), nil
	case []interface{}:
		// Convert slice to list of tftypes.Value
		values := make([]tftypes.Value, 0, len(v))
		for _, item := range v {
			itemValue, err := goValueToTerraformValueForDynamic(item)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("error converting list item: %w", err)
			}
			values = append(values, itemValue)
		}
		return tftypes.NewValue(tftypes.List{ElementType: tftypes.DynamicPseudoType}, values), nil
	case map[string]interface{}:
		// For the top-level map (properties), create it as a map type with DynamicPseudoType elements
		// This ensures Terraform sees it as an object type
		values := make(map[string]tftypes.Value, len(v))
		for key, item := range v {
			// Convert each value, wrapping in DynamicPseudoType
			var itemValue tftypes.Value
			switch itemVal := item.(type) {
			case []interface{}:
				// Convert slice to []tftypes.Value first
				tfValues := make([]tftypes.Value, 0, len(itemVal))
				for _, elem := range itemVal {
					elemValue, err := goValueToTerraformValueForDynamic(elem)
					if err != nil {
						return tftypes.Value{}, fmt.Errorf("error converting slice element for key %s: %w", key, err)
					}
					tfValues = append(tfValues, elemValue)
				}
				// Pass []tftypes.Value directly to DynamicPseudoType
				itemValue = tftypes.NewValue(tftypes.DynamicPseudoType, tfValues)
			case map[string]interface{}:
				// Convert nested map to map[string]tftypes.Value first
				tfMapValues := make(map[string]tftypes.Value, len(itemVal))
				for k, v := range itemVal {
					elemValue, err := goValueToTerraformValueForDynamic(v)
					if err != nil {
						return tftypes.Value{}, fmt.Errorf("error converting nested map value for key %s.%s: %w", key, k, err)
					}
					tfMapValues[k] = elemValue
				}
				// Pass map[string]tftypes.Value directly to DynamicPseudoType
				itemValue = tftypes.NewValue(tftypes.DynamicPseudoType, tfMapValues)
			default:
				// For primitive types, wrap directly in DynamicPseudoType
				itemValue = tftypes.NewValue(tftypes.DynamicPseudoType, item)
			}
			values[key] = itemValue
		}
		// Create map with DynamicPseudoType elements (this creates an object type)
		return tftypes.NewValue(tftypes.Map{ElementType: tftypes.DynamicPseudoType}, values), nil
	default:
		return tftypes.NewValue(tftypes.DynamicPseudoType, val), nil
	}
}

// valueFromTerraform converts a tftypes.Value to an attr.Value.
// This is a helper to work around the type system.
func valueFromTerraform(ctx context.Context, tfValue tftypes.Value) (attr.Value, error) {
	// Use DynamicValue's ValueFromTerraform method
	// We'll create a temporary DynamicValue to use its conversion
	dynamicType := basetypes.DynamicType{}
	return dynamicType.ValueFromTerraform(ctx, tfValue)
}

// attrValueToGoValue converts an attr.Value to a Go value (map, slice, or primitive).
// This is used to extract values from types.Dynamic for sending to the API.
func attrValueToGoValue(val attr.Value) (interface{}, error) {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil, nil
	}

	// Handle different attr.Value types
	switch v := val.(type) {
	case types.String:
		return v.ValueString(), nil
	case types.Bool:
		return v.ValueBool(), nil
	case types.Number:
		// Convert big.Float to float64
		f, _ := v.ValueBigFloat().Float64()
		return f, nil
	case types.Int64:
		return v.ValueInt64(), nil
	case types.Float64:
		return v.ValueFloat64(), nil
	case types.List:
		elements := v.Elements()
		result := make([]interface{}, 0, len(elements))
		for _, elem := range elements {
			goVal, err := attrValueToGoValue(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil
	case types.Tuple:
		elements := v.Elements()
		result := make([]interface{}, 0, len(elements))
		for _, elem := range elements {
			goVal, err := attrValueToGoValue(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil
	case types.Map:
		elements := v.Elements()
		result := make(map[string]interface{}, len(elements))
		for key, elem := range elements {
			goVal, err := attrValueToGoValue(elem)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil
	case types.Object:
		attrs := v.Attributes()
		result := make(map[string]interface{}, len(attrs))
		for key, elem := range attrs {
			goVal, err := attrValueToGoValue(elem)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil
	case types.Dynamic:
		// Recursively handle nested Dynamic values
		return attrValueToGoValue(v.UnderlyingValue())
	default:
		// For unknown types, try to use reflection or return an error
		return nil, fmt.Errorf("unsupported attr.Value type: %T", val)
	}
}
