// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/big"

	"terraform-provider-dx/internal/provider/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &scorecardResource{}

// var _ resource.ResourceWithImportState = &scorecardResource{}

func NewScorecardResource() resource.Resource {
	return &scorecardResource{}
}

// scorecardResource defines the resource implementation.
type scorecardResource struct {
	client *dxapi.Client
}

// scorecardModel describes the resource data model.
type scorecardModel struct {
	// Required fields
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	EntityFilterType    types.String `tfsdk:"entity_filter_type"`
	EvaluationFrequency types.Number `tfsdk:"evaluation_frequency_hours"`

	// Conditionally required fields for levels based scorecards
	EmptyLevelLabel types.String `tfsdk:"empty_level_label"`
	EmptyLevelColor types.String `tfsdk:"empty_level_color"`
	Levels          []levelModel `tfsdk:"levels"`

	// Conditionally required fields for points based scorecards
	CheckGroups []checkGroupModel `tfsdk:"check_groups"`

	// Optional fields
	Description                 types.String   `tfsdk:"description"`
	Published                   types.Bool     `tfsdk:"published"`
	EntityFilterTypeIdentifiers []types.String `tfsdk:"entity_filter_type_identifiers"`
	EntityFilterSql             types.String   `tfsdk:"entity_filter_sql"`
	Checks                      []checkModel   `tfsdk:"checks"`
}

type levelModel struct {
	Key   types.String `tfsdk:"key"`
	Id    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
	Rank  types.Number `tfsdk:"rank"`
}

type checkGroupModel struct {
	Key      types.String `tfsdk:"key"`
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Ordering types.Number `tfsdk:"ordering"`
}

type checkModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Ordering      types.Number `tfsdk:"ordering"`
	Sql           types.String `tfsdk:"sql"`
	FilterSql     types.String `tfsdk:"filter_sql"`
	FilterMessage types.String `tfsdk:"filter_message"`
	OutputEnabled types.Bool   `tfsdk:"output_enabled"`

	OutputType          types.String `tfsdk:"output_type"`
	OutputAggregation   types.String `tfsdk:"output_aggregation"`
	OutputCustomOptions types.String `tfsdk:"output_custom_options"` //TODO figure out how to model this

	EstimatedDevDays types.Number `tfsdk:"estimated_dev_days"`
	ExternalUrl      types.String `tfsdk:"external_url"`
	Published        types.Bool   `tfsdk:"published"`

	// Additional fields for level based scorecards
	ScorecardLevelKey types.String `tfsdk:"scorecard_level_key"`
	Level             levelModel   `tfsdk:"level"`

	// Additional fields for points based scorecards
	ScorecardCheckGroupKey types.String    `tfsdk:"scorecard_check_group_key"`
	CheckGroup             checkGroupModel `tfsdk:"check_group"`
	Points                 types.Number    `tfsdk:"points"`
}

func (r *scorecardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scorecard"
}

func (r *scorecardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scorecardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DX Scorecard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique ID of the scorecard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the scorecard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of scorecard. Options: 'LEVEL', 'POINTS'.",
				// Validators: []validator.String{
				// 	stringvalidator.OneOf("LEVEL", "POINTS"),
				// },
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity_filter_type": schema.StringAttribute{
				Required:    true,
				Description: "The filtering strategy when deciding what entities this scorecard should assess. Options: 'entity_types', 'sql'",
				// Validators: []validator.String{
				// 	stringvalidator.OneOf("entity_types", "sql"),
				// },
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"evaluation_frequency_hours": schema.NumberAttribute{
				Required:    true,
				Description: "How often the scorecard is evaluated (in hours). [2|4|8|24]",
				// Validators: []validator.Number{
				// 	numbervalidator.OneOf(2, 4, 8, 24),
				// },
				PlanModifiers: []planmodifier.Number{
					numberplanmodifier.UseStateForUnknown(),
				},
			},

			// Conditionally required for levels-based scorecards
			"empty_level_label": schema.StringAttribute{
				Optional:    true,
				Description: "The label to display when an entity has not achieved any levels in the scorecard (levels scorecards only).",
			},
			"empty_level_color": schema.StringAttribute{
				Optional:    true,
				Description: "The color hex code to display when an entity has not achieved any levels in the scorecard (levels scorecards only).",
			},
			"levels": schema.ListNestedAttribute{
				Optional:    true,
				Description: "The levels that can be achieved in this scorecard (levels scorecards only).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{Required: true},
						"id":    schema.StringAttribute{Computed: true},
						"name":  schema.StringAttribute{Required: true},
						"color": schema.StringAttribute{Required: true},
						"rank":  schema.NumberAttribute{Required: true},
					},
				},
			},

			// Conditionally required for points-based scorecards
			"check_groups": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Groups of checks, to help organize the scorecard for entity owners (points scorecards only).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":      schema.StringAttribute{Required: true},
						"id":       schema.StringAttribute{Computed: true},
						"name":     schema.StringAttribute{Required: true},
						"ordering": schema.NumberAttribute{Required: true},
					},
				},
			},

			// Optional metadata
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the scorecard.",
			},
			"published": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the scorecard is published.",
			},
			"entity_filter_type_identifiers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of entity type identifiers that the scorecard should run against.",
			},
			"entity_filter_sql": schema.StringAttribute{
				Optional:    true,
				Description: "Custom SQL used to filter entities that the scorecard should run against.",
			},

			// For now, all check field are required. This may change in the future.
			"checks": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of checks that are applied to entities in the scorecard.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true},
						"name":                  schema.StringAttribute{Required: true},
						"description":           schema.StringAttribute{Required: true},
						"ordering":              schema.NumberAttribute{Required: true},
						"sql":                   schema.StringAttribute{Required: true},
						"filter_sql":            schema.StringAttribute{Required: true},
						"filter_message":        schema.StringAttribute{Required: true},
						"output_enabled":        schema.BoolAttribute{Required: true},
						"output_type":           schema.StringAttribute{Required: true},
						"output_aggregation":    schema.StringAttribute{Required: true},
						"output_custom_options": schema.StringAttribute{Required: true}, // JSON string (you may eventually want to use a map)
						"estimated_dev_days":    schema.NumberAttribute{Required: true},
						"external_url":          schema.StringAttribute{Required: true},
						"published":             schema.BoolAttribute{Required: true},

						// Fields for level-based scorecards
						"scorecard_level_key": schema.StringAttribute{Optional: true},
						"level": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"key":   schema.StringAttribute{Required: true},
								"id":    schema.StringAttribute{Computed: true},
								"name":  schema.StringAttribute{Required: true},
								"color": schema.StringAttribute{Required: true},
								"rank":  schema.NumberAttribute{Required: true},
							},
						},

						// Fields for points-based scorecards
						"scorecard_check_group_key": schema.StringAttribute{Optional: true},
						"check_group": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Optional check group. If provided, all its fields (except 'id') are required.",
							Attributes: map[string]schema.Attribute{
								"key":      schema.StringAttribute{Required: true},
								"id":       schema.StringAttribute{Computed: true},
								"name":     schema.StringAttribute{Required: true},
								"ordering": schema.NumberAttribute{Required: true},
							},
						},
						"points": schema.NumberAttribute{Optional: true},
					},
				},
			},
		},
	}
}

func (r *scorecardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan scorecardModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// Construct API request payload
	payload := map[string]interface{}{
		// Required fields
		"name":                       plan.Name.ValueString(),
		"type":                       scorecardType,
		"entity_filter_type":         plan.EntityFilterType.ValueString(),
		"evaluation_frequency_hours": plan.EvaluationFrequency.ValueBigFloat(),
	}

	// Add LEVEL-specific required fields
	if scorecardType == "LEVEL" {
		payload["empty_level_label"] = plan.EmptyLevelLabel.ValueString()
		payload["empty_level_color"] = plan.EmptyLevelColor.ValueString()

		levels := []map[string]interface{}{}
		for _, level := range plan.Levels {
			levels = append(levels, map[string]interface{}{
				"key":   level.Key.ValueString(),
				"name":  level.Name.ValueString(),
				"color": level.Color.ValueString(),
				"rank":  level.Rank.ValueBigFloat(),
			})
		}
		payload["levels"] = levels
	}

	// Add POINTS-specific required fields
	if scorecardType == "POINTS" {
		checkGroups := []map[string]interface{}{}
		for _, group := range plan.CheckGroups {
			checkGroups = append(checkGroups, map[string]interface{}{
				"key":      group.Key.ValueString(),
				"name":     group.Name.ValueString(),
				"ordering": group.Ordering,
			})
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
	for _, check := range plan.Checks {
		checkPayload := map[string]interface{}{
			"name":                  check.Name.ValueString(),
			"description":           check.Description.ValueString(),
			"ordering":              check.Ordering,
			"sql":                   check.Sql.ValueString(),
			"filter_sql":            check.FilterSql.ValueString(),
			"filter_message":        check.FilterMessage.ValueString(),
			"output_enabled":        check.OutputEnabled.ValueBool(),
			"output_type":           check.OutputType.ValueString(),
			"output_aggregation":    check.OutputAggregation.ValueString(),
			"output_custom_options": check.OutputCustomOptions.ValueString(),
			"estimated_dev_days":    check.EstimatedDevDays,
			"external_url":          check.ExternalUrl.ValueString(),
			"published":             check.Published.ValueBool(),
		}

		// Add LEVEL-specific check fields
		if scorecardType == "LEVEL" {
			checkPayload["scorecard_level_key"] = check.ScorecardLevelKey.ValueString()
			checkPayload["level"] = map[string]interface{}{
				"key":   check.Level.Key.ValueString(),
				"id":    check.Level.Id.ValueString(),
				"name":  check.Level.Name.ValueString(),
				"color": check.Level.Color.ValueString(),
				"rank":  check.Level.Rank.ValueBigFloat(),
			}
		}

		// Add POINTS-specific check fields
		if scorecardType == "POINTS" {
			checkPayload["scorecard_check_group_key"] = check.ScorecardCheckGroupKey.ValueString()
			checkPayload["check_group"] = map[string]interface{}{
				"key":      check.CheckGroup.Key.ValueString(),
				"name":     check.CheckGroup.Name.ValueString(),
				"ordering": check.CheckGroup.Ordering,
			}
			checkPayload["points"] = check.Points
		}

		checks = append(checks, checkPayload)
	}
	payload["checks"] = checks

	// Create Scorecard (apiResp is a struct of type APIResponse)
	apiResp, err := r.client.CreateScorecard(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating scorecard", err.Error())
		return
	}

	// Shallow copy of plan to preserve values
	oldPlan := plan
	mapApiResponseToTerraformModel(apiResp, &plan, &oldPlan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func mapApiResponseToTerraformModel(apiResp *dxapi.APIResponse, plan *scorecardModel, oldPlan *scorecardModel) {

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
	numberOrNull := func(n *int) types.Number {
		if n != nil {
			return types.NumberValue(big.NewFloat(float64(*n)))
		}
		return types.NumberNull()
	}

	// ************** Required fields **************
	plan.Id = types.StringValue(apiResp.Scorecard.Id)
	plan.Name = types.StringValue(apiResp.Scorecard.Name)
	plan.Type = types.StringValue(apiResp.Scorecard.Type)
	plan.EntityFilterType = types.StringValue(apiResp.Scorecard.EntityFilterType)
	plan.EvaluationFrequency = types.NumberValue(big.NewFloat(float64(apiResp.Scorecard.EvaluationFrequency)))

	// ************** Conditionally required fields for levels based scorecards **************
	plan.EmptyLevelLabel = stringOrNull(apiResp.Scorecard.EmptyLevelLabel)
	plan.EmptyLevelColor = stringOrNull(apiResp.Scorecard.EmptyLevelColor)

	// If there are levels in the API response, update the plan.Levels
	if len(apiResp.Scorecard.Levels) > 0 {

		plan.Levels = make([]levelModel, len(apiResp.Scorecard.Levels))
		for i, lvl := range apiResp.Scorecard.Levels {
			var oldLevel levelModel
			if i < len(oldPlan.Levels) {
				oldLevel = oldPlan.Levels[i]
			}
			plan.Levels[i] = levelModel{
				// Key not returned by API. Leave same as plan.
				Key:   oldLevel.Key,
				Id:    stringOrNull(lvl.Id),
				Name:  stringOrNull(lvl.Name),
				Color: stringOrNull(lvl.Color),
				Rank:  numberOrNull(lvl.Rank),
			}
		}
	} else {
		plan.Levels = oldPlan.Levels
	}

	// ************** Conditionally required fields for points based scorecards **************

	// If there are check groups in the API response, update the plan.CheckGroups
	if len(apiResp.Scorecard.CheckGroups) > 0 {

		plan.CheckGroups = make([]checkGroupModel, len(apiResp.Scorecard.CheckGroups))
		for i, grp := range apiResp.Scorecard.CheckGroups {
			var prevCheckGroup checkGroupModel
			if i < len(oldPlan.CheckGroups) {
				prevCheckGroup = oldPlan.CheckGroups[i]
			}
			plan.CheckGroups[i] = checkGroupModel{
				// Key not returned by API. Leave same as plan.
				Key:      prevCheckGroup.Key,
				Id:       stringOrNull(grp.Id),
				Name:     stringOrNull(grp.Name),
				Ordering: numberOrNull(grp.Ordering),
			}
		}
	} else {
		plan.CheckGroups = oldPlan.CheckGroups
	}

	// ************** Optional fields **************
	plan.Description = stringOrNull(apiResp.Scorecard.Description)
	plan.EntityFilterSql = stringOrNull(apiResp.Scorecard.EntityFilterSql)
	plan.Published = boolApiToTF(apiResp.Scorecard.Published, plan.Published)

	// If there are entity filter type identifiers, update the plan.EntityFilterTypeIdentifiers
	if len(apiResp.Scorecard.EntityFilterTypeIdentifiers) > 0 {
		identifiers := make([]types.String, len(apiResp.Scorecard.EntityFilterTypeIdentifiers))
		for i, id := range apiResp.Scorecard.EntityFilterTypeIdentifiers {
			identifiers[i] = stringOrNull(id)
		}
		plan.EntityFilterTypeIdentifiers = identifiers
	} else {
		plan.EntityFilterTypeIdentifiers = oldPlan.EntityFilterTypeIdentifiers
	}

	// If there are checks in the API response, update the plan.Checks
	if len(apiResp.Scorecard.Checks) > 0 {
		plan.Checks = make([]checkModel, len(apiResp.Scorecard.Checks))
		for i, chk := range apiResp.Scorecard.Checks {
			var prevCheck checkModel
			if i < len(oldPlan.Checks) {
				prevCheck = oldPlan.Checks[i]
			}
			plan.Checks[i] = checkModel{
				Id:                  stringOrNull(chk.Id),
				Name:                stringOrNull(chk.Name),
				Description:         stringOrNull(chk.Description),
				Ordering:            numberOrNull(chk.Ordering),
				Sql:                 stringOrNull(chk.Sql),
				FilterSql:           stringOrNull(chk.FilterSql),
				FilterMessage:       stringOrNull(chk.FilterMessage),
				OutputEnabled:       boolApiToTF(chk.OutputEnabled, plan.Checks[i].OutputEnabled),
				OutputType:          stringOrNull(chk.OutputType),
				OutputAggregation:   stringOrNull(chk.OutputAggregation),
				OutputCustomOptions: stringOrNull(chk.OutputCustomOptions),
				EstimatedDevDays:    numberOrNull(chk.EstimatedDevDays),
				ExternalUrl:         stringOrNull(chk.ExternalUrl),
				Published:           boolApiToTF(chk.Published, plan.Checks[i].Published),
				// Key not returned by API. Leave same as plan.
				ScorecardLevelKey: prevCheck.ScorecardLevelKey,
				Level: levelModel{
					// Key not returned by API. Leave same as plan.
					Key:   prevCheck.Level.Key,
					Id:    stringOrNull(chk.Level.Id),
					Name:  stringOrNull(chk.Level.Name),
					Color: stringOrNull(chk.Level.Color),
					Rank:  numberOrNull(chk.Level.Rank),
				},
				// Key not returned by API. Leave same as plan.
				ScorecardCheckGroupKey: prevCheck.ScorecardCheckGroupKey,
				CheckGroup: checkGroupModel{
					// Key not returned by API. Leave same as plan.
					Key:      prevCheck.CheckGroup.Key,
					Id:       stringOrNull(chk.CheckGroup.Id),
					Name:     stringOrNull(chk.CheckGroup.Name),
					Ordering: numberOrNull(chk.CheckGroup.Ordering),
				},
				Points: numberOrNull(chk.Points),
			}
		}
	} else {
		plan.Checks = oldPlan.Checks
	}
}

func (r *scorecardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scorecardModel

	// Load existing state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	mapApiResponseToTerraformModel(apiResp, &state, &oldState)
	// state.Id = types.StringValue(apiResp.Scorecard.Id)
	// state.Name = types.StringValue(apiResp.Scorecard.Name)
	// // state.Description = types.StringValue(apiResp.Scorecard.Description)
	// state.Type = types.StringValue(apiResp.Scorecard.Type)
	// Map other fields as needed
	// ...

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *scorecardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scorecardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) // Get the desired state
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the payload, similar to Create, but include the id
	payload := map[string]interface{}{
		"id":                         plan.Id.ValueString(),
		"name":                       plan.Name.ValueString(),
		"type":                       plan.Type.ValueString(),
		"entity_filter_type":         plan.EntityFilterType.ValueString(),
		"evaluation_frequency_hours": plan.EvaluationFrequency.ValueBigFloat(),
	}

	scorecardType := plan.Type.ValueString()
	if scorecardType == "LEVEL" {
		payload["empty_level_label"] = plan.EmptyLevelLabel.ValueString()
		payload["empty_level_color"] = plan.EmptyLevelColor.ValueString()
		levels := []map[string]interface{}{}
		for _, level := range plan.Levels {
			levels = append(levels, map[string]interface{}{
				"key":   level.Key.ValueString(),
				"id":    level.Id.ValueString(),
				"name":  level.Name.ValueString(),
				"color": level.Color.ValueString(),
				"rank":  level.Rank.ValueBigFloat(),
			})
		}
		payload["levels"] = levels
	}
	if scorecardType == "POINTS" {
		checkGroups := []map[string]interface{}{}
		for _, group := range plan.CheckGroups {
			checkGroups = append(checkGroups, map[string]interface{}{
				"key":      group.Key.ValueString(),
				"id":       group.Id.ValueString(),
				"name":     group.Name.ValueString(),
				"ordering": group.Ordering,
			})
		}
		payload["check_groups"] = checkGroups
	}
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
	checks := []map[string]interface{}{}
	for _, check := range plan.Checks {
		checkPayload := map[string]interface{}{
			"id":                    check.Id.ValueString(),
			"name":                  check.Name.ValueString(),
			"description":           check.Description.ValueString(),
			"ordering":              check.Ordering,
			"sql":                   check.Sql.ValueString(),
			"filter_sql":            check.FilterSql.ValueString(),
			"filter_message":        check.FilterMessage.ValueString(),
			"output_enabled":        check.OutputEnabled.ValueBool(),
			"output_type":           check.OutputType.ValueString(),
			"output_aggregation":    check.OutputAggregation.ValueString(),
			"output_custom_options": check.OutputCustomOptions.ValueString(),
			"estimated_dev_days":    check.EstimatedDevDays,
			"external_url":          check.ExternalUrl.ValueString(),
			"published":             check.Published.ValueBool(),
		}
		if scorecardType == "LEVEL" {
			checkPayload["scorecard_level_key"] = check.ScorecardLevelKey.ValueString()
			checkPayload["level"] = map[string]interface{}{
				"key":   check.Level.Key.ValueString(),
				"id":    check.Level.Id.ValueString(),
				"name":  check.Level.Name.ValueString(),
				"color": check.Level.Color.ValueString(),
				"rank":  check.Level.Rank.ValueBigFloat(),
			}
		}
		if scorecardType == "POINTS" {
			checkPayload["scorecard_check_group_key"] = check.ScorecardCheckGroupKey.ValueString()
			checkPayload["check_group"] = map[string]interface{}{
				"key":      check.CheckGroup.Key.ValueString(),
				"id":       check.CheckGroup.Id.ValueString(),
				"name":     check.CheckGroup.Name.ValueString(),
				"ordering": check.CheckGroup.Ordering,
			}
			checkPayload["points"] = check.Points
		}
		checks = append(checks, checkPayload)
	}
	payload["checks"] = checks

	apiResp, err := r.client.UpdateScorecard(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating scorecard", err.Error())
		return
	}

	oldPlan := plan
	mapApiResponseToTerraformModel(apiResp, &plan, &oldPlan)

	// Map API response to Terraform state model

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *scorecardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scorecardModel
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

func (r *scorecardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
