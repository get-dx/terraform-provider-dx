package scorecard

import (
	"context"
	"fmt"
	"sort"

	"terraform-provider-dx/dx"
	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iancoleman/strcase"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ScorecardResource{}
	_ resource.ResourceWithImportState = &ScorecardResource{}
)

func NewScorecardResource() resource.Resource {
	return &ScorecardResource{}
}

// scorecardResource defines the resource implementation.
type ScorecardResource struct {
	client *dxapi.Client
}

func (r *ScorecardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scorecard"
}

func (r *ScorecardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dxapi.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The API client was not configured. This is a bug in the provider.")
		return
	}
}

func (r *ScorecardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating scorecard resource!")

	// Retrieve values from plan
	var plan ScorecardModel
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

	// Create Scorecard (apiResp is a struct of type APIResponse)
	apiResp, err := r.client.CreateScorecard(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating scorecard", err.Error())
		return
	}

	// Shallow copy of plan to preserve values
	oldPlan := plan
	responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ScorecardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading scorecard resource")

	var state ScorecardModel

	// Load existing state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Prior state, before reading from API: %v", state))

	// Extract ID
	id := state.Id.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing ID", "The resource ID is missing from the state")
		return
	}

	// Call the API to get the latest scorecard data
	apiResp, err := r.client.GetScorecard(ctx, id)
	if err != nil {
		// TODO - implement resource not found error handling
		// 	// Resource no longer exists remotely — remove from state
		// 	resp.State.RemoveResource(ctx)
		// 	return
		// }
		resp.Diagnostics.AddError(
			"Error reading scorecard",
			fmt.Sprintf("Could not read scorecard ID %s: %s", id, err.Error()),
		)
		return
	}

	// Map API response to Terraform state model
	// Shallow copy of plan to preserve values
	oldState := state
	responseBodyToModel(ctx, apiResp, &state, &oldState)
	// state.Id = types.StringValue(apiResp.Scorecard.Id)
	// state.Name = types.StringValue(apiResp.Scorecard.Name)
	// // state.Description = types.StringValue(apiResp.Scorecard.Description)
	// state.Type = types.StringValue(apiResp.Scorecard.Type)
	// Map other fields as needed
	// ...

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ScorecardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScorecardModel
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

	apiResp, err := r.client.UpdateScorecard(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating scorecard", err.Error())
		return
	}

	// Map API response to Terraform state model
	oldPlan := plan
	responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ScorecardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScorecardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get the current state
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing ID", "The resource ID is missing from the state")
		return
	}

	success, err := r.client.DeleteScorecard(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting scorecard", err.Error())
		return
	}
	if !success {
		resp.Diagnostics.AddError("Error deleting scorecard", "API did not confirm deletion.")
		return
	}
	// No need to set state, resource will be removed by Terraform if this method returns successfully
}

func (r *ScorecardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Importing scorecard state")

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func ValidateModel(plan ScorecardModel, diags *diag.Diagnostics) {
	// Validate required fields for CREATE endpoint
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		diags.AddError("Missing required field", "The 'name' field must be specified.")
		return
	}
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		diags.AddError("Missing required field", "The 'type' field must be specified.")
		return
	}
	if plan.EntityFilterType.IsNull() || plan.EntityFilterType.IsUnknown() {
		diags.AddError("Missing required field", "The 'entity_filter_type' field must be specified.")
		return
	}
	if plan.EvaluationFrequency.IsNull() || plan.EvaluationFrequency.IsUnknown() {
		diags.AddError("Missing required field", "The 'evaluation_frequency_hours' field must be specified.")
		return
	}

	// Validate tags
	if len(plan.Tags) > 0 {
		for _, tag := range plan.Tags {
			if tag.Value.IsNull() || tag.Value.IsUnknown() {
				diags.AddError("Missing required field", "The 'tags.value' field must be specified.")
				return
			}
		}
	}

	// Validate required fields based on scorecard type
	scorecardType := plan.Type.ValueString()
	switch scorecardType {
	case "LEVEL":
		if plan.EmptyLevelLabel.IsNull() || plan.EmptyLevelLabel.IsUnknown() {
			diags.AddError("Missing required field", "The 'empty_level_label' field must be specified for LEVEL scorecards.")
		}
		if plan.EmptyLevelColor.IsNull() || plan.EmptyLevelColor.IsUnknown() {
			diags.AddError("Missing required field", "The 'empty_level_color' field must be specified for LEVEL scorecards.")
		}
		if len(plan.Levels) == 0 {
			diags.AddError("Missing required field", "At least one 'level' must be specified for LEVEL scorecards.")
		}

		levelKeys := make(map[string]bool)
		for levelKey := range plan.Levels {
			levelKeys[levelKey] = true
		}

		for _, check := range plan.Checks {
			if check.ScorecardLevelKey.IsNull() {
				diags.AddError("Missing required field", "The 'scorecard_level_key' field must be specified for checks in LEVEL scorecards.")
			}

			levelKey := check.ScorecardLevelKey.ValueString()
			if !levelKeys[levelKey] {
				diags.AddError("Invalid value", fmt.Sprintf("The 'scorecard_level_key' field value of `%s` does not match any level keys", levelKey))
			}
		}

		// Validate that there are no duplicate ordering values for checks within the same level
		validateNoDuplicateOrdering(plan.Checks, func(check CheckModel) string {
			if !check.ScorecardLevelKey.IsNull() {
				return check.ScorecardLevelKey.ValueString()
			}
			return ""
		}, "Level", diags)

	case "POINTS":
		if len(plan.CheckGroups) == 0 {
			diags.AddError("Missing required field", "At least one 'check_group' must be specified for POINTS scorecards.")
		}

		checkGroupKeys := make(map[string]bool)
		for checkGroupKey := range plan.CheckGroups {
			checkGroupKeys[checkGroupKey] = true
		}

		for _, check := range plan.Checks {
			if check.ScorecardCheckGroupKey.IsNull() {
				diags.AddError("Missing required field", "The 'scorecard_check_group_key' field must be specified for checks in POINTS scorecards.")
			}

			checkGroupKey := check.ScorecardCheckGroupKey.ValueString()
			if !checkGroupKeys[checkGroupKey] {
				diags.AddError("Invalid value", fmt.Sprintf("The 'scorecard_check_group_key' field value of `%s` does not match any check group keys", checkGroupKey))
			}
		}

		// Validate that there are no duplicate ordering values for checks within the same check group
		validateNoDuplicateOrdering(plan.Checks, func(check CheckModel) string {
			if !check.ScorecardCheckGroupKey.IsNull() {
				return check.ScorecardCheckGroupKey.ValueString()
			}
			return ""
		}, "Check group", diags)
	default:
		diags.AddError("Invalid scorecard type", fmt.Sprintf("Unsupported scorecard type: %s", scorecardType))
	}
}

// Validates that there are no duplicate ordering values for checks within the same container (level or check group).
func validateNoDuplicateOrdering(checks map[string]CheckModel, getContainerKey func(CheckModel) string, containerType string, diags *diag.Diagnostics) {
	containerOrderingMap := make(map[string]map[int32][]string) // containerKey -> ordering -> []checkKey

	for checkKey, check := range checks {
		if !check.Ordering.IsNull() {
			containerKey := getContainerKey(check)
			if containerKey != "" {
				ordering := check.Ordering.ValueInt32()

				if containerOrderingMap[containerKey] == nil {
					containerOrderingMap[containerKey] = make(map[int32][]string)
				}

				containerOrderingMap[containerKey][ordering] = append(containerOrderingMap[containerKey][ordering], checkKey)
			}
		}
	}

	// Check for duplicates and report errors
	for containerKey, orderingMap := range containerOrderingMap {
		for ordering, checkKeys := range orderingMap {
			if len(checkKeys) > 1 {
				// Sort check keys for consistent error messages
				sort.Strings(checkKeys)

				// Create a comma-separated list of check keys with backticks
				var formattedCheckKeys []string
				for _, key := range checkKeys {
					formattedCheckKeys = append(formattedCheckKeys, fmt.Sprintf("`%s`", key))
				}

				// Join the check keys with ", "
				checkKeysStr := ""
				for i, key := range formattedCheckKeys {
					if i > 0 {
						checkKeysStr += ", "
					}
					checkKeysStr += key
				}

				errorMsg := fmt.Sprintf("%s `%s`: the following checks have a duplicate ordering of %d: %s",
					containerType, containerKey, ordering, checkKeysStr)

				diags.AddError("Duplicate check ordering", errorMsg)
			}
		}
	}
}

func modelToRequestBody(ctx context.Context, plan ScorecardModel, setIds bool) (map[string]interface{}, error) {
	tflog.Debug(ctx, "Converting plan to request body")

	scorecardType := plan.Type.ValueString()

	// Construct API request payload
	payload := map[string]interface{}{
		// Required fields
		"name":                       plan.Name.ValueString(),
		"type":                       scorecardType,
		"entity_filter_type":         plan.EntityFilterType.ValueString(),
		"evaluation_frequency_hours": plan.EvaluationFrequency.ValueInt32(),
	}

	if len(plan.Tags) > 0 {
		tags := []map[string]interface{}{}
		for _, tag := range plan.Tags {
			tags = append(tags, map[string]interface{}{"value": tag.Value.ValueString()})
		}
		payload["tags"] = tags
	}

	if setIds {
		payload["id"] = plan.Id.ValueString()
	}

	// Add LEVEL-specific required fields
	if scorecardType == "LEVEL" {
		payload["empty_level_label"] = plan.EmptyLevelLabel.ValueString()
		payload["empty_level_color"] = plan.EmptyLevelColor.ValueString()

		levels := []map[string]interface{}{}
		for planLevelKey, planLevel := range plan.Levels {
			level := map[string]interface{}{
				"key":   planLevelKey,
				"name":  planLevel.Name.ValueString(),
				"color": planLevel.Color.ValueString(),
				"rank":  planLevel.Rank.ValueInt32(),
			}
			if setIds {
				level["id"] = planLevel.Id.ValueString()
			}
			levels = append(levels, level)
		}
		payload["levels"] = levels
	}

	// Add POINTS-specific required fields
	if scorecardType == "POINTS" {
		checkGroups := []map[string]interface{}{}
		for planCheckGroupKey, planCheckGroup := range plan.CheckGroups {
			checkGroup := map[string]interface{}{
				"key":      planCheckGroupKey,
				"name":     planCheckGroup.Name.ValueString(),
				"ordering": planCheckGroup.Ordering.ValueInt32(),
			}
			if setIds {
				checkGroup["id"] = planCheckGroup.Id.ValueString()
			}
			checkGroups = append(checkGroups, checkGroup)
		}
		payload["check_groups"] = checkGroups
	}

	// Add optional fields if they're present
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Published.IsNull() && !plan.Published.IsUnknown() {
		payload["published"] = plan.Published.ValueBool()
	}
	if len(plan.EntityFilterTypeIdentifiers) > 0 {
		identifiers := make([]string, 0, len(plan.EntityFilterTypeIdentifiers))
		for _, id := range plan.EntityFilterTypeIdentifiers {
			if !id.IsNull() && !id.IsUnknown() {
				identifiers = append(identifiers, id.ValueString())
			}
		}
		payload["entity_filter_type_identifiers"] = identifiers
	}
	if !plan.EntityFilterSql.IsNull() && !plan.EntityFilterSql.IsUnknown() {
		payload["entity_filter_sql"] = plan.EntityFilterSql.ValueString()
	}

	// Add checks
	checks := []map[string]interface{}{}
	for _, planCheck := range plan.Checks {
		var estimatedDevDaysValue interface{}
		if planCheck.EstimatedDevDays.IsNull() || planCheck.EstimatedDevDays.IsUnknown() {
			estimatedDevDaysValue = nil
		} else {
			estimatedDevDaysValue = planCheck.EstimatedDevDays.ValueFloat32()
		}

		checkPayload := map[string]interface{}{
			"name":               planCheck.Name.ValueString(),
			"description":        planCheck.Description.ValueString(),
			"ordering":           planCheck.Ordering.ValueInt32(),
			"sql":                planCheck.Sql.ValueString(),
			"filter_sql":         planCheck.FilterSql.ValueString(),
			"filter_message":     planCheck.FilterMessage.ValueString(),
			"output_enabled":     planCheck.OutputEnabled.ValueBool(),
			"output_type":        nil,
			"output_aggregation": nil,
			"estimated_dev_days": estimatedDevDaysValue,
			"external_url":       planCheck.ExternalUrl.ValueString(),
			"published":          planCheck.Published.ValueBool(),
		}

		if setIds {
			checkPayload["id"] = planCheck.Id.ValueString()
		}

		if checkPayload["output_enabled"] == true {
			checkPayload["output_type"] = planCheck.OutputType.ValueString()
			if !planCheck.OutputAggregation.IsNull() {
				checkPayload["output_aggregation"] = planCheck.OutputAggregation.ValueString()
			}
		}

		if planCheck.OutputType.ValueString() == "custom" {
			planCustomOptions := planCheck.OutputCustomOptions
			if planCustomOptions == nil {
				return nil, fmt.Errorf("output_custom_options is required when output_type is `custom`")
			}

			planCustomOptionsVal := *planCustomOptions
			customOptionsPayload := map[string]interface{}{
				"unit":     planCustomOptionsVal.Unit.ValueString(),
				"decimals": "auto",
			}
			if !planCustomOptionsVal.Decimals.IsNull() {
				customOptionsPayload["decimals"] = planCustomOptionsVal.Decimals.ValueInt32()
			}

			checkPayload["output_custom_options"] = customOptionsPayload
		}

		// Add LEVEL-specific check fields
		if scorecardType == "LEVEL" {
			checkPayload["scorecard_level_key"] = planCheck.ScorecardLevelKey.ValueString()
		}

		// Add POINTS-specific check fields
		if scorecardType == "POINTS" {
			checkPayload["scorecard_check_group_key"] = planCheck.ScorecardCheckGroupKey.ValueString()
			checkPayload["points"] = planCheck.Points.ValueInt32()
		}

		checks = append(checks, checkPayload)
	}
	payload["checks"] = checks

	return payload, nil
}

func responseBodyToModel(ctx context.Context, apiResp *dxapi.APIResponse, state *ScorecardModel, oldPlan *ScorecardModel) {
	tflog.Debug(ctx, "Mapping API response to Terraform model")

	// ************** Required fields **************
	state.Id = types.StringValue(apiResp.Scorecard.Id)
	state.Name = types.StringValue(apiResp.Scorecard.Name)
	state.Type = types.StringValue(apiResp.Scorecard.Type)
	state.EntityFilterType = types.StringValue(apiResp.Scorecard.EntityFilterType)
	state.EvaluationFrequency = types.Int32Value(apiResp.Scorecard.EvaluationFrequency)

	// ************** Conditionally required fields for levels based scorecards **************
	state.EmptyLevelLabel = dx.StringOrNull(apiResp.Scorecard.EmptyLevelLabel)
	state.EmptyLevelColor = dx.StringOrNull(apiResp.Scorecard.EmptyLevelColor)

	// If there are levels in the API response, update the plan.Levels
	if len(apiResp.Scorecard.Levels) > 0 {
		state.Levels = make(map[string]LevelModel)
		orderedLevelKeys := getOrderedLevelKeys(*oldPlan)
		for idxResp, lvl := range apiResp.Scorecard.Levels {
			levelName := *lvl.Name
			levelKey := nameToKey(ctx, levelName)

			if idxResp < len(orderedLevelKeys) {
				levelKey = orderedLevelKeys[idxResp]
			}

			state.Levels[levelKey] = LevelModel{
				Id:    types.StringValue(*lvl.Id),
				Name:  types.StringValue(levelName),
				Color: types.StringValue(*lvl.Color),
				Rank:  types.Int32Value(*lvl.Rank),
			}
		}
	} else {
		state.Levels = oldPlan.Levels
	}

	// ************** Conditionally required fields for points based scorecards **************

	// If there are check groups in the API response, update the state.CheckGroups
	if len(apiResp.Scorecard.CheckGroups) > 0 {
		state.CheckGroups = make(map[string]CheckGroupModel)
		orderedCheckGroupKeys := getOrderedCheckGroupKeys(*oldPlan)
		for idxResp, grp := range apiResp.Scorecard.CheckGroups {
			groupName := *grp.Name
			groupKey := nameToKey(ctx, groupName)

			if idxResp < len(orderedCheckGroupKeys) {
				groupKey = orderedCheckGroupKeys[idxResp]
			}

			state.CheckGroups[groupKey] = CheckGroupModel{
				Id:       types.StringValue(*grp.Id),
				Name:     types.StringValue(groupName),
				Ordering: types.Int32Value(*grp.Ordering),
			}
		}
	} else {
		state.CheckGroups = oldPlan.CheckGroups
	}

	// ************** Optional fields **************
	state.Description = dx.StringOrNull(apiResp.Scorecard.Description)
	state.EntityFilterSql = dx.StringOrNullConvertEmpty(apiResp.Scorecard.EntityFilterSql)
	state.Published = dx.BoolApiToTF(apiResp.Scorecard.Published, state.Published)

	// If there are entity filter type identifiers, update the state.EntityFilterTypeIdentifiers
	if len(apiResp.Scorecard.EntityFilterTypeIdentifiers) > 0 {
		identifiers := make([]types.String, len(apiResp.Scorecard.EntityFilterTypeIdentifiers))
		for i, id := range apiResp.Scorecard.EntityFilterTypeIdentifiers {
			identifiers[i] = dx.StringOrNull(id)
		}
		state.EntityFilterTypeIdentifiers = identifiers
	} else {
		state.EntityFilterTypeIdentifiers = oldPlan.EntityFilterTypeIdentifiers
	}

	// Update the state.Checks
	orderedCheckKeys := getOrderedCheckKeys(*oldPlan)
	state.Checks = make(map[string]CheckModel)
	for idxResp, chk := range apiResp.Scorecard.Checks {
		var levelKey *string = nil
		var checkGroupKey *string = nil

		// Find the previous check, based on mapping the response index back to the check's key
		var prevCheck *CheckModel
		checkKey := nameToKey(ctx, *chk.Name)
		if idxResp < len(orderedCheckKeys) {
			// Grouping keys are not returned by the API, but we have found previous values to fallback to
			checkKey = orderedCheckKeys[idxResp]
			foundPrevCheck := oldPlan.Checks[checkKey]
			prevCheck = &foundPrevCheck

			var prevLevelKey *string = nil
			if !prevCheck.ScorecardLevelKey.IsNull() {
				prevLevelKeyVal := prevCheck.ScorecardLevelKey.ValueString()
				prevLevelKey = &prevLevelKeyVal
			}
			levelKey = prevLevelKey

			var prevCheckGroupKey *string = nil
			if !prevCheck.ScorecardCheckGroupKey.IsNull() {
				prevCheckGroupKeyVal := prevCheck.ScorecardCheckGroupKey.ValueString()
				prevCheckGroupKey = &prevCheckGroupKeyVal
			}
			checkGroupKey = prevCheckGroupKey

			tflog.Info(
				ctx,
				fmt.Sprintf(
					"Response check with index %d has key `%s`, found previous check with name `%s`",
					idxResp,
					checkKey,
					prevCheck.Name.ValueString(),
				),
			)
		}

		var outputCustomOptions *OutputCustomOptionsModel = nil
		if chk.OutputCustomOptions != nil {
			decimals := chk.OutputCustomOptions.Decimals
			if decimals.IsAuto {
				outputCustomOptions = &OutputCustomOptionsModel{
					Unit:     types.StringValue(chk.OutputCustomOptions.Unit),
					Decimals: types.Int32Null(),
				}
			} else {
				decimalsValue := *decimals.FixedValue
				outputCustomOptions = &OutputCustomOptionsModel{
					Unit:     types.StringValue(chk.OutputCustomOptions.Unit),
					Decimals: types.Int32Value(decimalsValue),
				}
			}
		}

		state.Checks[checkKey] = CheckModel{
			Id:                  dx.StringOrNull(chk.Id),
			Name:                dx.StringOrNull(chk.Name),
			Description:         dx.StringOrNullConvertEmpty(chk.Description),
			Ordering:            types.Int32Value(chk.Ordering),
			Sql:                 dx.StringOrNull(chk.Sql),
			FilterSql:           dx.StringOrNullConvertEmpty(chk.FilterSql),
			FilterMessage:       dx.StringOrNullConvertEmpty(chk.FilterMessage),
			OutputEnabled:       types.BoolValue(chk.OutputEnabled),
			OutputType:          dx.StringOrNull(chk.OutputType),
			OutputAggregation:   dx.StringOrNull(chk.OutputAggregation),
			OutputCustomOptions: outputCustomOptions,
			EstimatedDevDays:    dx.Float32OrNull(chk.EstimatedDevDays),
			ExternalUrl:         dx.StringOrNullConvertEmpty(chk.ExternalUrl),
			Published:           types.BoolValue(chk.Published),
			Points:              dx.Int32OrNull(chk.Points),

			ScorecardLevelKey:      dx.StringOrNull(levelKey),
			ScorecardCheckGroupKey: dx.StringOrNull(checkGroupKey),
		}
	}
}

// Create a list of level keys, ordered by their rank.
func getOrderedLevelKeys(plan ScorecardModel) []string {
	type levelInfo struct {
		key  string
		rank int32
	}

	// Create slice to hold level information for sorting
	levels := make([]levelInfo, 0, len(plan.Levels))

	// Collect level information
	for levelKey, level := range plan.Levels {
		levels = append(levels, levelInfo{
			key:  levelKey,
			rank: level.Rank.ValueInt32(),
		})
	}

	// Sort the levels based on rank
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].rank < levels[j].rank
	})

	// Extract just the keys in sorted order
	orderedKeys := make([]string, len(levels))
	for i, level := range levels {
		orderedKeys[i] = level.key
	}

	return orderedKeys
}

// Create a list of check group keys, ordered by their ordering.
func getOrderedCheckGroupKeys(plan ScorecardModel) []string {
	type checkGroupInfo struct {
		key      string
		ordering int32
	}

	// Create slice to hold check group information for sorting
	checkGroups := make([]checkGroupInfo, 0, len(plan.CheckGroups))

	// Collect check group information
	for groupKey, group := range plan.CheckGroups {
		checkGroups = append(checkGroups, checkGroupInfo{
			key:      groupKey,
			ordering: group.Ordering.ValueInt32(),
		})
	}

	// Sort the check groups based on ordering
	sort.Slice(checkGroups, func(i, j int) bool {
		return checkGroups[i].ordering < checkGroups[j].ordering
	})

	// Extract just the keys in sorted order
	orderedKeys := make([]string, len(checkGroups))
	for i, group := range checkGroups {
		orderedKeys[i] = group.key
	}

	return orderedKeys
}

// Create a list of check keys, ordered by their level/check-group, then their ordering within that grouping.
func getOrderedCheckKeys(plan ScorecardModel) []string {
	type checkInfo struct {
		key          string
		ordering     int32
		groupingRank int32
	}

	// Create slice to hold check information for sorting
	checks := make([]checkInfo, 0, len(plan.Checks))

	if plan.Type.ValueString() == "LEVEL" {
		// Create a map of level keys to their ranks for efficient lookup
		levelRanks := make(map[string]int32)
		for levelKey, level := range plan.Levels {
			levelRanks[levelKey] = level.Rank.ValueInt32()
		}

		// Collect check information with level ranks
		for key, check := range plan.Checks {
			levelKey := check.ScorecardLevelKey.ValueString()
			checks = append(checks, checkInfo{
				key:          key,
				ordering:     check.Ordering.ValueInt32(),
				groupingRank: levelRanks[levelKey],
			})
		}
	} else if plan.Type.ValueString() == "POINTS" {
		// Create a map of check group keys to their ordering for efficient lookup
		groupOrderings := make(map[string]int32)
		for groupKey, group := range plan.CheckGroups {
			groupOrderings[groupKey] = group.Ordering.ValueInt32()
		}

		// Collect check information with group orderings
		for key, check := range plan.Checks {
			groupKey := check.ScorecardCheckGroupKey.ValueString()
			checks = append(checks, checkInfo{
				key:          key,
				ordering:     check.Ordering.ValueInt32(),
				groupingRank: groupOrderings[groupKey],
			})
		}
	}

	// Sort the checks based on rank and ordering
	sort.Slice(checks, func(i, j int) bool {
		if checks[i].groupingRank != checks[j].groupingRank {
			return checks[i].groupingRank < checks[j].groupingRank
		}
		return checks[i].ordering < checks[j].ordering
	})

	// Extract just the keys in sorted order
	orderedKeys := make([]string, len(checks))
	for i, check := range checks {
		orderedKeys[i] = check.key
	}

	return orderedKeys
}

// Convert a level/check-group/check name to a key.
func nameToKey(ctx context.Context, name string) string {
	result := strcase.ToSnake(name)
	tflog.Info(ctx, fmt.Sprintf("Converted name `%s` to key `%s`", name, result))
	return result
}
