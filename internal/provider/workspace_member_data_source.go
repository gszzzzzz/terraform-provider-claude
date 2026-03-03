package provider

import (
	"context"
	"fmt"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &workspaceMemberDataSource{}

type workspaceMemberDataSource struct {
	client *client.Client
}

type workspaceMemberDataSourceModel struct {
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	UserID        types.String `tfsdk:"user_id"`
	WorkspaceRole types.String `tfsdk:"workspace_role"`
}

func NewWorkspaceMemberDataSource() datasource.DataSource {
	return &workspaceMemberDataSource{}
}

func (d *workspaceMemberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (d *workspaceMemberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a Claude workspace member.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the workspace.",
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the user.",
			},
			"workspace_role": schema.StringAttribute{
				Computed:    true,
				Description: "The role of the member in the workspace.",
			},
		},
	}
}

func (d *workspaceMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workspaceMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workspaceMemberDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := d.client.GetWorkspaceMember(ctx, config.WorkspaceID.ValueString(), config.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workspace member", err.Error())
		return
	}

	state := workspaceMemberDataSourceModel{
		WorkspaceID:   types.StringValue(member.WorkspaceID),
		UserID:        types.StringValue(member.UserID),
		WorkspaceRole: types.StringValue(member.WorkspaceRole),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
