package provider

import (
	"context"
	"fmt"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &workspaceMembersDataSource{}

type workspaceMembersDataSource struct {
	client *client.Client
}

type workspaceMembersDataSourceModel struct {
	WorkspaceID types.String           `tfsdk:"workspace_id"`
	Members     []workspaceMemberModel `tfsdk:"members"`
}

type workspaceMemberModel struct {
	UserID        types.String `tfsdk:"user_id"`
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	WorkspaceRole types.String `tfsdk:"workspace_role"`
}

func NewWorkspaceMembersDataSource() datasource.DataSource {
	return &workspaceMembersDataSource{}
}

func (d *workspaceMembersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_members"
}

func (d *workspaceMembersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all members of a Claude workspace.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the workspace.",
			},
			"members": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of workspace members.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the user.",
						},
						"workspace_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the workspace.",
						},
						"workspace_role": schema.StringAttribute{
							Computed:    true,
							Description: "The role of the member in the workspace.",
						},
					},
				},
			},
		},
	}
}

func (d *workspaceMembersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workspaceMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workspaceMembersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allMembers []client.WorkspaceMember
	params := client.ListWorkspaceMembersParams{}

	for {
		result, err := d.client.ListWorkspaceMembers(ctx, config.WorkspaceID.ValueString(), params)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list workspace members", err.Error())
			return
		}

		allMembers = append(allMembers, result.Data...)

		if !result.HasMore {
			break
		}
		params.AfterID = result.LastID
	}

	members := make([]workspaceMemberModel, len(allMembers))
	for i, m := range allMembers {
		members[i] = workspaceMemberModel{
			UserID:        types.StringValue(m.UserID),
			WorkspaceID:   types.StringValue(m.WorkspaceID),
			WorkspaceRole: types.StringValue(m.WorkspaceRole),
		}
	}

	state := workspaceMembersDataSourceModel{
		WorkspaceID: config.WorkspaceID,
		Members:     members,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
