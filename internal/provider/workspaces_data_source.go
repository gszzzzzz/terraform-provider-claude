package provider

import (
	"context"
	"fmt"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &workspacesDataSource{}

type workspacesDataSource struct {
	client *client.Client
}

type workspacesDataSourceModel struct {
	IncludeArchived types.Bool        `tfsdk:"include_archived"`
	Workspaces      []workspacesModel `tfsdk:"workspaces"`
}

type workspacesModel struct {
	ID            types.String        `tfsdk:"id"`
	Name          types.String        `tfsdk:"name"`
	DisplayColor  types.String        `tfsdk:"display_color"`
	CreatedAt     types.String        `tfsdk:"created_at"`
	ArchivedAt    types.String        `tfsdk:"archived_at"`
	DataResidency *dataResidencyModel `tfsdk:"data_residency"`
}

func NewWorkspacesDataSource() datasource.DataSource {
	return &workspacesDataSource{}
}

func (d *workspacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspaces"
}

func (d *workspacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list workspaces in the Claude organization.",
		Attributes: map[string]schema.Attribute{
			"include_archived": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to include archived workspaces in the results.",
			},
			"workspaces": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of workspaces.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the workspace.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the workspace.",
						},
						"display_color": schema.StringAttribute{
							Computed:    true,
							Description: "The display color of the workspace.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the workspace was created.",
						},
						"archived_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the workspace was archived, if applicable.",
						},
						"data_residency": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "The data residency configuration for the workspace.",
							Attributes: map[string]schema.Attribute{
								"workspace_geo": schema.StringAttribute{
									Computed:    true,
									Description: "The geographic region for workspace data.",
								},
								"default_inference_geo": schema.StringAttribute{
									Computed:    true,
									Description: "The default geographic region for inference.",
								},
								"allowed_inference_geos": schema.ListAttribute{
									Computed:    true,
									ElementType: types.StringType,
									Description: "The allowed geographic regions for inference.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *workspacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *workspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workspacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allWorkspaces []client.Workspace
	params := client.ListWorkspacesParams{}
	if !config.IncludeArchived.IsNull() && config.IncludeArchived.ValueBool() {
		params.IncludeArchived = true
	}

	for {
		result, err := d.client.ListWorkspaces(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list workspaces", err.Error())
			return
		}

		allWorkspaces = append(allWorkspaces, result.Data...)

		if !result.HasMore || result.LastID == "" {
			break
		}
		params.AfterID = result.LastID
	}

	workspaces := make([]workspacesModel, len(allWorkspaces))
	for i, w := range allWorkspaces {
		model := workspacesModel{
			ID:           types.StringValue(w.ID),
			Name:         types.StringValue(w.Name),
			DisplayColor: types.StringValue(w.DisplayColor),
			CreatedAt:    types.StringValue(w.CreatedAt),
		}

		if w.ArchivedAt != nil {
			model.ArchivedAt = types.StringValue(*w.ArchivedAt)
		} else {
			model.ArchivedAt = types.StringNull()
		}

		if w.DataResidency != nil {
			dr := &dataResidencyModel{
				WorkspaceGeo:        types.StringValue(w.DataResidency.WorkspaceGeo),
				DefaultInferenceGeo: types.StringValue(w.DataResidency.DefaultInferenceGeo),
			}
			geoList, geoDiags := parseAllowedInferenceGeos(ctx, w.DataResidency.AllowedInferenceGeos)
			resp.Diagnostics.Append(geoDiags...)
			dr.AllowedInferenceGeos = geoList
			model.DataResidency = dr
		}

		workspaces[i] = model
	}

	state := workspacesDataSourceModel{
		IncludeArchived: config.IncludeArchived,
		Workspaces:      workspaces,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
