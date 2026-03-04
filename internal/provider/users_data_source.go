package provider

import (
	"context"
	"fmt"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &usersDataSource{}

type usersDataSource struct {
	client *client.Client
}

type usersDataSourceModel struct {
	Email types.String `tfsdk:"email"`
	Users []usersModel `tfsdk:"users"`
}

type usersModel struct {
	ID      types.String `tfsdk:"id"`
	Email   types.String `tfsdk:"email"`
	Name    types.String `tfsdk:"name"`
	Role    types.String `tfsdk:"role"`
	AddedAt types.String `tfsdk:"added_at"`
}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list users in the Claude organization.",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "Filter users by email address.",
			},
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the user.",
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "The email address of the user.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the user.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "The role of the user in the organization.",
						},
						"added_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the user was added to the organization.",
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allUsers []client.User
	params := client.ListUsersParams{}
	if !config.Email.IsNull() {
		params.Email = config.Email.ValueString()
	}

	for {
		result, err := d.client.ListUsers(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list users", err.Error())
			return
		}

		allUsers = append(allUsers, result.Data...)

		if !result.HasMore || result.LastID == "" {
			break
		}
		params.AfterID = result.LastID
	}

	users := make([]usersModel, len(allUsers))
	for i, u := range allUsers {
		users[i] = usersModel{
			ID:      types.StringValue(u.ID),
			Email:   types.StringValue(u.Email),
			Name:    types.StringValue(u.Name),
			Role:    types.StringValue(u.Role),
			AddedAt: types.StringValue(u.AddedAt),
		}
	}

	state := usersDataSourceModel{
		Email: config.Email,
		Users: users,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
