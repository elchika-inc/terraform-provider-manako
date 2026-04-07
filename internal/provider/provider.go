package provider

import (
	"context"
	"os"

	"github.com/elchika-inc/terraform-provider-manako/internal/client"
	"github.com/elchika-inc/terraform-provider-manako/internal/resources/monitor"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ManakoProvider{}

type ManakoProvider struct {
	version string
}

type ManakoProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ManakoProvider{version: version}
	}
}

func (p *ManakoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "manako"
	resp.Version = p.version
}

func (p *ManakoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Manako monitoring service.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Manako API Key (mk_*). Can also be set via MANAKO_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Manako API base URL.",
				Optional:    true,
			},
		},
	}
}

func (p *ManakoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ManakoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("MANAKO_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires a Manako API key. Set it via the api_key attribute or the MANAKO_API_KEY environment variable.",
		)
		return
	}

	baseURL := "https://api.manako.dev"
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	c := client.NewClient(baseURL, apiKey, p.version)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *ManakoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		monitor.NewMonitorResource,
	}
}

func (p *ManakoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		monitor.NewMonitorDataSource,
		monitor.NewMonitorsDataSource,
	}
}
