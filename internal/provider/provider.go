package provider

import (
	"context"
	"os"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &claudeProvider{}

type claudeProvider struct {
	version string
}

type claudeProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

// New creates a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &claudeProvider{
			version: version,
		}
	}
}

func (p *claudeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "claude"
	resp.Version = p.version
}

func (p *claudeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Claude resources via the Admin API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API key for the Claude Admin API. Can also be set via ANTHROPIC_ADMIN_API_KEY environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "The base URL for the Claude Admin API. Defaults to https://api.anthropic.com. Can also be set via ANTHROPIC_BASE_URL environment variable.",
			},
		},
	}
}

func (p *claudeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config claudeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown API Key",
			"The provider cannot be configured with an unknown API key. Set a known value or use the ANTHROPIC_ADMIN_API_KEY environment variable.",
		)
		return
	}

	apiKey := os.Getenv("ANTHROPIC_ADMIN_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires an API key. Set api_key in the provider configuration or set the ANTHROPIC_ADMIN_API_KEY environment variable.",
		)
		return
	}

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown Base URL",
			"The provider cannot be configured with an unknown base URL. Set a known value or use the ANTHROPIC_BASE_URL environment variable.",
		)
		return
	}

	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	c := client.NewClient(apiKey, baseURL)

	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *claudeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWorkspaceResource,
	}
}

func (p *claudeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
	}
}
