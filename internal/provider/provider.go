package provider

import (
	"context"
	"os"

	"terraform-provider-dx/dx/dxapi"
	"terraform-provider-dx/dx/scorecard"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure DxProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &DxProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DxProvider{
			Version: version,
		}
	}
}

// DxProvider defines the provider implementation.
type DxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	Client  *dxapi.Client
	Token   string
	Version string
}

// DxProviderModel describes the provider data model.
type DxProviderModel struct {
	ApiToken types.String `tfsdk:"api_token"`
}

func (p *DxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dx"
	resp.Version = p.Version
}

func (p *DxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "DX Web API token for authentication.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring DX provider")

	var config DxProviderModel

	// Load provider config
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := config.ApiToken.ValueString()

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider could not retrieve an API token. This is required to authenticate with the DX API.",
		)
		return
	}

	// Initialize HTTP client
	baseURL := os.Getenv("DX_WEB_API_URL")
	if baseURL == "" {
		baseURL = "https://api.getdx.com"
	}
	client := dxapi.NewClient(baseURL, token)
	// p.client = client

	resp.ResourceData = client
	// Set if we create a data source
	// resp.DataSourceData = client
}

func (p *DxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		scorecard.NewScorecardResource,
	}
}

func (p *DxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}
