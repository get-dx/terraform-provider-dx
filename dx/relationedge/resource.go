package relationedge

import (
	"context"
	"fmt"
	"strings"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &RelationEdgeResource{}
	_ resource.ResourceWithImportState = &RelationEdgeResource{}
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

func (r *RelationEdgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier, err := parseEdgeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("relation_identifier"), relationIdentifier)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_entity_identifier"), sourceEntityIdentifier)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_entity_identifier"), targetEntityIdentifier)...)
}

// parseEdgeID splits an ID produced by edgeID back into its relation, source
// entity, and target entity identifiers.
func parseEdgeID(id string) (string, string, string, error) {
	parts := strings.Split(id, "+")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("expected import ID in the form relation_identifier+source_entity_identifier+target_entity_identifier, got %q", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func edgeID(m RelationEdgeModel) string {
	return fmt.Sprintf("%s+%s+%s",
		m.RelationIdentifier.ValueString(),
		m.SourceEntityIdentifier.ValueString(),
		m.TargetEntityIdentifier.ValueString(),
	)
}
