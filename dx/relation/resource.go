package relation

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-dx/dx"
	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &RelationResource{}
	_ resource.ResourceWithImportState = &RelationResource{}
)

func NewRelationResource() resource.Resource {
	return &RelationResource{}
}

type RelationResource struct {
	client *dxapi.Client
}

func (r *RelationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog_relation"
}

func (r *RelationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RelationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating catalog relation resource")

	var plan RelationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := modelToCreatePayload(plan)

	apiResp, err := r.client.CreateRelation(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating catalog relation", err.Error())
		return
	}

	responseToModel(apiResp, &plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RelationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading catalog relation resource")

	var state RelationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()
	if identifier == "" {
		resp.Diagnostics.AddError("Missing identifier", "The resource identifier is missing from the state")
		return
	}

	apiResp, err := r.client.GetRelation(ctx, identifier)
	if err != nil {
		var notFound *dxapi.RelationNotFoundError
		if errors.As(err, &notFound) {
			tflog.Info(ctx, fmt.Sprintf("Catalog relation %s not found, removing from state", identifier))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading catalog relation",
			fmt.Sprintf("Could not read catalog relation with identifier %s: %s", identifier, err.Error()),
		)
		return
	}

	responseToModel(apiResp, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RelationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Updating catalog relation resource")

	var plan RelationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := modelToUpdatePayload(plan)

	apiResp, err := r.client.UpdateRelation(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating catalog relation", err.Error())
		return
	}

	responseToModel(apiResp, &plan)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RelationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RelationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()
	if identifier == "" {
		resp.Diagnostics.AddError("Missing identifier", "The resource identifier is missing from the state")
		return
	}

	success, err := r.client.DeleteRelation(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting catalog relation", err.Error())
		return
	}
	if !success {
		resp.Diagnostics.AddError("Error deleting catalog relation", "API did not confirm deletion.")
		return
	}
}

func (r *RelationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Importing catalog relation state")
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func modelToCreatePayload(plan RelationModel) map[string]interface{} {
	payload := map[string]interface{}{
		"identifier":                     plan.Identifier.ValueString(),
		"type":                           plan.Type.ValueString(),
		"cardinality":                    plan.Cardinality.ValueString(),
		"source_entity_type_identifier":  plan.SourceEntityTypeIdentifier.ValueString(),
		"target_entity_type_identifier":  plan.TargetEntityTypeIdentifier.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	return payload
}

func modelToUpdatePayload(plan RelationModel) map[string]interface{} {
	payload := map[string]interface{}{
		"identifier": plan.Identifier.ValueString(),
		"type":       plan.Type.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	} else {
		payload["description"] = ""
	}

	return payload
}

func responseToModel(apiResp *dxapi.APIRelationResponse, state *RelationModel) {
	state.Id = types.StringValue(apiResp.Relation.Identifier)
	state.Identifier = types.StringValue(apiResp.Relation.Identifier)
	state.Type = types.StringValue(apiResp.Relation.Type)
	state.InverseType = types.StringValue(apiResp.Relation.InverseType)
	state.Cardinality = types.StringValue(apiResp.Relation.Cardinality)
	state.Description = dx.StringOrNullConvertEmpty(apiResp.Relation.Description)
	state.SourceEntityTypeIdentifier = types.StringValue(apiResp.Relation.SourceEntityTypeIdentifier)
	state.TargetEntityTypeIdentifier = types.StringValue(apiResp.Relation.TargetEntityTypeIdentifier)
	state.CreatedAt = types.StringValue(apiResp.Relation.CreatedAt)
	state.UpdatedAt = types.StringValue(apiResp.Relation.UpdatedAt)
}
