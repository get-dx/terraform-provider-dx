package scorecard

import (
	"context"
	"fmt"
	"sort"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	// Validate required fields for CREATE endpoint
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		resp.Diagnostics.AddError("Missing required field", "The 'name' field must be specified.")
		return
	}
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		resp.Diagnostics.AddError("Missing required field", "The 'type' field must be specified.")
		return
	}
	if plan.EntityFilterType.IsNull() || plan.EntityFilterType.IsUnknown() {
		resp.Diagnostics.AddError("Missing required field", "The 'entity_filter_type' field must be specified.")
		return
	}
	if plan.EvaluationFrequency.IsNull() || plan.EvaluationFrequency.IsUnknown() {
		resp.Diagnostics.AddError("Missing required field", "The 'evaluation_frequency_hours' field must be specified.")
		return
	}

	// Validate required fields based on scorecard type
	scorecardType := plan.Type.ValueString()
	switch scorecardType {
	case "LEVEL":
		if plan.EmptyLevelLabel.IsNull() || plan.EmptyLevelLabel.IsUnknown() {
			resp.Diagnostics.AddError("Missing required field", "The 'empty_level_label' field must be specified for LEVEL scorecards.")
		}
		if plan.EmptyLevelColor.IsNull() || plan.EmptyLevelColor.IsUnknown() {
			resp.Diagnostics.AddError("Missing required field", "The 'empty_level_color' field must be specified for LEVEL scorecards.")
		}
		if len(plan.Levels) == 0 {
			resp.Diagnostics.AddError("Missing required field", "At least one 'level' must be specified for LEVEL scorecards.")
		}
	case "POINTS":
		if len(plan.CheckGroups) == 0 {
			resp.Diagnostics.AddError("Missing required field", "At least one 'check_group' must be specified for POINTS scorecards.")
		}
	default:
		resp.Diagnostics.AddError("Invalid scorecard type", fmt.Sprintf("Unsupported scorecard type: %s", scorecardType))
	}

	// If there are any errors above, return immediately.
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
			checkPayload["output_aggregation"] = planCheck.OutputAggregation.ValueString()
		}

		if planCheck.OutputType.ValueString() == "custom" {
			return nil, fmt.Errorf("output type of `custom` is not yet supported")
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

	// ************** Helper functions **************

	// Helper checks for and handles nil strings
	stringOrNull := func(s *string) types.String {
		if s != nil {
			return types.StringValue(*s)
		}
		return types.StringNull()
	}

	// Helper preserves the value of a bool field if it's null in the plan
	boolApiToTF := func(apiVal bool, planVal types.Bool) types.Bool {
		if planVal.IsNull() && !apiVal {
			return types.BoolNull()
		}
		return types.BoolValue(apiVal)
	}

	// Helper checks for and handles nil ints
	float32OrNull := func(f *float32) types.Float32 {
		if f != nil {
			return types.Float32Value(*f)
		}
		return types.Float32Null()
	}

	int32OrNull := func(i *int32) types.Int32 {
		if i != nil {
			return types.Int32Value(*i)
		}
		return types.Int32Null()
	}

	// ************** Required fields **************
	state.Id = types.StringValue(apiResp.Scorecard.Id)
	state.Name = types.StringValue(apiResp.Scorecard.Name)
	state.Type = types.StringValue(apiResp.Scorecard.Type)
	state.EntityFilterType = types.StringValue(apiResp.Scorecard.EntityFilterType)
	state.EvaluationFrequency = types.Int32Value(apiResp.Scorecard.EvaluationFrequency)

	// ************** Conditionally required fields for levels based scorecards **************
	state.EmptyLevelLabel = stringOrNull(apiResp.Scorecard.EmptyLevelLabel)
	state.EmptyLevelColor = stringOrNull(apiResp.Scorecard.EmptyLevelColor)

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
	state.Description = stringOrNull(apiResp.Scorecard.Description)
	state.EntityFilterSql = stringOrNull(apiResp.Scorecard.EntityFilterSql)
	state.Published = boolApiToTF(apiResp.Scorecard.Published, state.Published)

	// If there are entity filter type identifiers, update the state.EntityFilterTypeIdentifiers
	if len(apiResp.Scorecard.EntityFilterTypeIdentifiers) > 0 {
		identifiers := make([]types.String, len(apiResp.Scorecard.EntityFilterTypeIdentifiers))
		for i, id := range apiResp.Scorecard.EntityFilterTypeIdentifiers {
			identifiers[i] = stringOrNull(id)
		}
		state.EntityFilterTypeIdentifiers = identifiers
	} else {
		state.EntityFilterTypeIdentifiers = oldPlan.EntityFilterTypeIdentifiers
	}

	// Update the state.Checks
	orderedCheckKeys := getOrderedCheckKeys(*oldPlan)
	state.Checks = make(map[string]CheckModel)
	for idxResp, chk := range apiResp.Scorecard.Checks {
		// Find the previous check, based on mapping the response index back to the check's key
		var prevCheck *CheckModel
		prevCheckKey := orderedCheckKeys[idxResp]
		if idxResp < len(orderedCheckKeys) {
			foundPrevCheck := oldPlan.Checks[prevCheckKey]
			prevCheck = &foundPrevCheck
			tflog.Info(ctx, fmt.Sprintf("Response check with index %d has key `%s`, found previous check with name `%s`", idxResp, prevCheckKey, prevCheck.Name.ValueString()))
		} else {
			prevCheck = nil
		}

		if prevCheck == nil {
			panic(fmt.Sprintf("No previous check found for check %s", *chk.Id))
		}

		state.Checks[prevCheckKey] = CheckModel{
			Id:                stringOrNull(chk.Id),
			Name:              stringOrNull(chk.Name),
			Description:       stringOrNull(chk.Description),
			Ordering:          types.Int32Value(chk.Ordering),
			Sql:               stringOrNull(chk.Sql),
			FilterSql:         stringOrNull(chk.FilterSql),
			FilterMessage:     stringOrNull(chk.FilterMessage),
			OutputEnabled:     types.BoolValue(chk.OutputEnabled),
			OutputType:        stringOrNull(chk.OutputType),
			OutputAggregation: stringOrNull(chk.OutputAggregation),
			OutputCustomOptions: types.ObjectNull(map[string]attr.Type{
				"unit":     types.StringType,
				"decimals": types.NumberType,
			}),
			EstimatedDevDays: float32OrNull(chk.EstimatedDevDays),
			ExternalUrl:      stringOrNull(chk.ExternalUrl),
			Published:        types.BoolValue(chk.Published),
			// Key not returned by API. Leave same as plan.
			ScorecardLevelKey: prevCheck.ScorecardLevelKey,
			// Key not returned by API. Leave same as plan.
			ScorecardCheckGroupKey: prevCheck.ScorecardCheckGroupKey,
			Points:                 int32OrNull(chk.Points),
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
