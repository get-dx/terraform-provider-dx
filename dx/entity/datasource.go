package entity

import (
	"context"
	"fmt"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EntityDataSource{}
	_ datasource.DataSourceWithConfigure = &EntityDataSource{}
)

func NewEntityDataSource() datasource.DataSource {
	return &EntityDataSource{}
}

// EntityDataSource defines the data source implementation.
type EntityDataSource struct {
	client *dxapi.Client
}

func (d *EntityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (d *EntityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a DX Entity from the catalog. Use this to reference existing entities without managing them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the entity (same as 'identifier').",
			},
			"identifier": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the entity to look up.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the entity type (e.g., 'service', 'api', 'domain').",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Display name for the entity.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the entity.",
			},
			"owner_teams": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				},
				Computed:    true,
				Description: "Array of owner teams assigned to the entity. Each team has 'id' and 'name' fields.",
			},
			"owner_users": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":    types.StringType,
						"email": types.StringType,
					},
				},
				Computed:    true,
				Description: "Array of owner users assigned to the entity. Each user has 'id' and 'email' fields.",
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the domain entity parent assigned to the entity.",
			},
			"properties": schema.DynamicAttribute{
				Computed:    true,
				Description: "Key-value pairs of entity properties and their values. Values can be strings, numbers, null, objects, or lists of any of those types.",
			},
			"aliases": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"identifier": types.StringType,
						},
					},
				},
				Computed:    true,
				Description: "Key-value pairs of aliases assigned to the entity. Keys are alias types (e.g., 'github_repo'), values are arrays of alias objects with 'identifier' field.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the entity was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the entity was last updated.",
			},
		},
	}
}

func (d *EntityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dxapi.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *dxapi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The API client was not configured. This is a bug in the provider.")
		return
	}
}

func (d *EntityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading entity data source")

	var config EntityDataSourceModel

	// Read configuration
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract identifier from config
	identifier := config.Identifier.ValueString()
	if identifier == "" {
		resp.Diagnostics.AddError("Missing identifier", "The entity identifier is required")
		return
	}

	// Call the API to get the entity data
	apiResp, err := d.client.GetEntity(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading entity",
			fmt.Sprintf("Could not read entity with identifier %s: %s", identifier, err.Error()),
		)
		return
	}

	// Map API response to data source model
	var state EntityDataSourceModel
	mapAPIResponseToDataSourceModel(ctx, apiResp, &state)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// EntityDataSourceModel describes the data source data model.
type EntityDataSourceModel struct {
	Id          types.String            `tfsdk:"id"`
	Identifier  types.String            `tfsdk:"identifier"`
	Type        types.String            `tfsdk:"type"`
	Name        types.String            `tfsdk:"name"`
	Description types.String            `tfsdk:"description"`
	OwnerTeams  []OwnerTeamModel        `tfsdk:"owner_teams"`
	OwnerUsers  []OwnerUserModel        `tfsdk:"owner_users"`
	Domain      types.String            `tfsdk:"domain"`
	Properties  types.Dynamic           `tfsdk:"properties"`
	Aliases     map[string][]AliasModel `tfsdk:"aliases"`
	CreatedAt   types.String            `tfsdk:"created_at"`
	UpdatedAt   types.String            `tfsdk:"updated_at"`
}

// OwnerTeamModel describes an owner team.
type OwnerTeamModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// OwnerUserModel describes an owner user.
type OwnerUserModel struct {
	Id    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

// mapAPIResponseToDataSourceModel converts API response to data source model.
func mapAPIResponseToDataSourceModel(ctx context.Context, apiResp *dxapi.APIEntityResponse, state *EntityDataSourceModel) {
	tflog.Debug(ctx, "Mapping API response to data source model")

	// Required fields
	state.Id = types.StringValue(apiResp.Entity.Identifier)
	state.Identifier = types.StringValue(apiResp.Entity.Identifier)
	state.Type = types.StringValue(apiResp.Entity.Type)

	// Optional fields - use StringPointerValue for proper null handling
	if apiResp.Entity.Name != nil {
		state.Name = types.StringValue(*apiResp.Entity.Name)
	} else {
		state.Name = types.StringNull()
	}

	if apiResp.Entity.Description != nil {
		state.Description = types.StringValue(*apiResp.Entity.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Owner teams
	if len(apiResp.Entity.OwnerTeams) > 0 {
		teams := make([]OwnerTeamModel, 0, len(apiResp.Entity.OwnerTeams))
		for _, team := range apiResp.Entity.OwnerTeams {
			teams = append(teams, OwnerTeamModel{
				Id:   types.StringValue(team.Id),
				Name: types.StringValue(team.Name),
			})
		}
		state.OwnerTeams = teams
	} else {
		state.OwnerTeams = []OwnerTeamModel{}
	}

	// Owner users
	if len(apiResp.Entity.OwnerUsers) > 0 {
		users := make([]OwnerUserModel, 0, len(apiResp.Entity.OwnerUsers))
		for _, user := range apiResp.Entity.OwnerUsers {
			users = append(users, OwnerUserModel{
				Id:    types.StringValue(user.Id),
				Email: types.StringValue(user.Email),
			})
		}
		state.OwnerUsers = users
	} else {
		state.OwnerUsers = []OwnerUserModel{}
	}

	// Domain
	if apiResp.Entity.Domain != nil {
		state.Domain = types.StringValue(apiResp.Entity.Domain.Identifier)
	} else {
		state.Domain = types.StringNull()
	}

	// Properties - convert from API response to types.Dynamic
	if len(apiResp.Entity.Properties) > 0 {
		dynamicVal, err := interfaceToDynamic(ctx, apiResp.Entity.Properties)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to convert properties to Dynamic: %v", err))
			state.Properties = types.DynamicNull()
		} else {
			state.Properties = dynamicVal
		}
	} else {
		state.Properties = types.DynamicNull()
	}

	// Aliases - convert from API response to map[string][]AliasModel
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
		state.Aliases = nil
	}

	// Computed fields
	state.CreatedAt = types.StringValue(apiResp.Entity.CreatedAt)
	state.UpdatedAt = types.StringValue(apiResp.Entity.UpdatedAt)
}
