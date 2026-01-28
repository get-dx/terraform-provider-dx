package entity

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &EntitiesDataSource{}
	_ datasource.DataSourceWithConfigure = &EntitiesDataSource{}
)

func NewEntitiesDataSource() datasource.DataSource {
	return &EntitiesDataSource{}
}

type EntitiesDataSource struct {
	client *dxapi.Client
}

type EntitiesDataSourceModel struct {
	Type     types.String          `tfsdk:"type"`
	Entities []EntitiesEntityModel `tfsdk:"entities"`
}

type EntitiesEntityModel struct {
	Id          types.String            `tfsdk:"id"`
	Identifier  types.String            `tfsdk:"identifier"`
	Type        types.String            `tfsdk:"type"`
	Name        types.String            `tfsdk:"name"`
	Description types.String            `tfsdk:"description"`
	OwnerTeams  []OwnerTeamModel        `tfsdk:"owner_teams"`
	OwnerUsers  []OwnerUserModel        `tfsdk:"owner_users"`
	Domain      types.String            `tfsdk:"domain"`
	Properties  types.String            `tfsdk:"properties"`
	Aliases     map[string][]AliasModel `tfsdk:"aliases"`
	CreatedAt   types.String            `tfsdk:"created_at"`
	UpdatedAt   types.String            `tfsdk:"updated_at"`
}

func (d *EntitiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entities"
}

func (d *EntitiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all DX entities of a given type from the catalog.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The entity type identifier to filter by (e.g., 'service', 'api', 'domain').",
			},
			"entities": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of entities matching the given type.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the entity (same as 'identifier').",
						},
						"identifier": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the entity.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The identifier of the entity type.",
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
							Description: "Array of owner teams assigned to the entity.",
						},
						"owner_users": schema.ListAttribute{
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"id":    types.StringType,
									"email": types.StringType,
								},
							},
							Computed:    true,
							Description: "Array of owner users assigned to the entity.",
						},
						"domain": schema.StringAttribute{
							Computed:    true,
							Description: "The identifier of the domain entity parent assigned to the entity.",
						},
						"properties": schema.StringAttribute{
							Computed:    true,
							Description: "JSON-encoded key-value pairs of entity properties. Use jsondecode() to access values.",
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
							Description: "Key-value pairs of aliases assigned to the entity.",
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
				},
			},
		},
	}
}

func (d *EntitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EntitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading entities data source")

	var config EntitiesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := config.Type.ValueString()
	if entityType == "" {
		resp.Diagnostics.AddError("Missing type", "The entity type is required")
		return
	}

	apiEntities, err := d.client.ListEntities(ctx, entityType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing entities",
			fmt.Sprintf("Could not list entities of type %s: %s", entityType, err.Error()),
		)
		return
	}

	state := EntitiesDataSourceModel{
		Type:     config.Type,
		Entities: make([]EntitiesEntityModel, 0, len(apiEntities)),
	}

	for i := range apiEntities {
		var entityModel EntitiesEntityModel
		mapAPIEntityToEntitiesModel(ctx, &apiEntities[i], &entityModel)
		state.Entities = append(state.Entities, entityModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// mapAPIEntityToEntitiesModel converts an APIEntity to the EntitiesEntityModel,
// encoding properties as a JSON string since Dynamic types cannot be nested
// inside collection attributes.
func mapAPIEntityToEntitiesModel(ctx context.Context, entity *dxapi.APIEntity, state *EntitiesEntityModel) {
	state.Id = types.StringValue(entity.Identifier)
	state.Identifier = types.StringValue(entity.Identifier)
	state.Type = types.StringValue(entity.Type)

	if entity.Name != nil {
		state.Name = types.StringValue(*entity.Name)
	} else {
		state.Name = types.StringNull()
	}

	if entity.Description != nil {
		state.Description = types.StringValue(*entity.Description)
	} else {
		state.Description = types.StringNull()
	}

	if len(entity.OwnerTeams) > 0 {
		teams := make([]OwnerTeamModel, 0, len(entity.OwnerTeams))
		for _, team := range entity.OwnerTeams {
			teams = append(teams, OwnerTeamModel{
				Id:   types.StringValue(team.Id),
				Name: types.StringValue(team.Name),
			})
		}
		state.OwnerTeams = teams
	} else {
		state.OwnerTeams = []OwnerTeamModel{}
	}

	if len(entity.OwnerUsers) > 0 {
		users := make([]OwnerUserModel, 0, len(entity.OwnerUsers))
		for _, user := range entity.OwnerUsers {
			users = append(users, OwnerUserModel{
				Id:    types.StringValue(user.Id),
				Email: types.StringValue(user.Email),
			})
		}
		state.OwnerUsers = users
	} else {
		state.OwnerUsers = []OwnerUserModel{}
	}

	if entity.Domain != nil {
		state.Domain = types.StringValue(entity.Domain.Identifier)
	} else {
		state.Domain = types.StringNull()
	}

	if len(entity.Properties) > 0 {
		propsJSON, err := json.Marshal(entity.Properties)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to marshal properties to JSON: %v", err))
			state.Properties = types.StringNull()
		} else {
			state.Properties = types.StringValue(string(propsJSON))
		}
	} else {
		state.Properties = types.StringNull()
	}

	if len(entity.Aliases) > 0 {
		aliases := make(map[string][]AliasModel)
		for aliasType, aliasArray := range entity.Aliases {
			aliasModels := make([]AliasModel, 0, len(aliasArray))
			for _, alias := range aliasArray {
				aliasModels = append(aliasModels, AliasModel{
					Identifier: types.StringValue(alias.Identifier),
				})
			}
			if len(aliasModels) > 0 {
				aliases[aliasType] = aliasModels
			}
		}
		state.Aliases = aliases
	} else {
		state.Aliases = nil
	}

	state.CreatedAt = types.StringValue(entity.CreatedAt)
	state.UpdatedAt = types.StringValue(entity.UpdatedAt)
}
