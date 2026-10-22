package relationedge

import (
	"context"
	"fmt"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &RelationEdgeResource{}
)

func NewRelationEdgeResource() resource.Resource {
	return &RelationEdgeResource{}
}

type RelationEdgeResource struct {
	client *dxapi.Client
}

func (r *RelationEdgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog_relation_edge"
}

func (r *RelationEdgeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RelationEdgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating catalog relation edge resource")

	var plan RelationEdgeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddRelationEdge(
		ctx,
		plan.RelationIdentifier.ValueString(),
		plan.SourceEntityIdentifier.ValueString(),
		plan.TargetEntityIdentifier.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating catalog relation edge", err.Error())
		return
	}

	plan.Id = types.StringValue(edgeID(plan))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RelationEdgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading catalog relation edge resource")

	var state RelationEdgeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exists, err := r.client.RelationEdgeExists(
		ctx,
		state.RelationIdentifier.ValueString(),
		state.SourceEntityIdentifier.ValueString(),
		state.TargetEntityIdentifier.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading catalog relation edge",
			fmt.Sprintf("Could not read catalog relation edge %s: %s", edgeID(state), err.Error()),
		)
		return
	}

	if !exists {
		tflog.Info(ctx, fmt.Sprintf("Catalog relation edge %s not found, removing from state", edgeID(state)))
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(edgeID(state))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is never called with changes because every attribute requires replacement.
func (r *RelationEdgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RelationEdgeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(edgeID(plan))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RelationEdgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RelationEdgeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveRelationEdge(
		ctx,
		state.RelationIdentifier.ValueString(),
		state.SourceEntityIdentifier.ValueString(),
		state.TargetEntityIdentifier.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting catalog relation edge", err.Error())
		return
	}
}

func edgeID(m RelationEdgeModel) string {
	return fmt.Sprintf("%s+%s+%s",
		m.RelationIdentifier.ValueString(),
		m.SourceEntityIdentifier.ValueString(),
		m.TargetEntityIdentifier.ValueString(),
	)
}
