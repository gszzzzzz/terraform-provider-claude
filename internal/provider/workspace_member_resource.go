package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &workspaceMemberResource{}
	_ resource.ResourceWithImportState = &workspaceMemberResource{}
)

type workspaceMemberResource struct {
	client *client.Client
}

type workspaceMemberResourceModel struct {
	ID            types.String `tfsdk:"id"`
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	UserID        types.String `tfsdk:"user_id"`
	WorkspaceRole types.String `tfsdk:"workspace_role"`
}

func NewWorkspaceMemberResource() resource.Resource {
	return &workspaceMemberResource{}
}

func (r *workspaceMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (r *workspaceMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Claude workspace member.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The composite identifier of the workspace member (workspace_id/user_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the workspace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_role": schema.StringAttribute{
				Required:    true,
				Description: "The role of the member in the workspace. Valid values: workspace_user, workspace_developer, workspace_admin.",
			},
		},
	}
}

func (r *workspaceMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *workspaceMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspaceMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.CreateWorkspaceMember(ctx, plan.WorkspaceID.ValueString(), client.CreateWorkspaceMemberRequest{
		UserID:        plan.UserID.ValueString(),
		WorkspaceRole: plan.WorkspaceRole.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create workspace member", err.Error())
		return
	}

	state := flattenWorkspaceMember(member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetWorkspaceMember(ctx, state.WorkspaceID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read workspace member", err.Error())
		return
	}

	newState := flattenWorkspaceMember(member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *workspaceMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workspaceMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.UpdateWorkspaceMember(ctx, plan.WorkspaceID.ValueString(), plan.UserID.ValueString(), client.UpdateWorkspaceMemberRequest{
		WorkspaceRole: plan.WorkspaceRole.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update workspace member", err.Error())
		return
	}

	state := flattenWorkspaceMember(member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workspaceMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteWorkspaceMember(ctx, state.WorkspaceID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete workspace member", err.Error())
	}
}

func (r *workspaceMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format 'workspace_id/user_id', got: %q", req.ID),
		)
		return
	}

	state := workspaceMemberResourceModel{
		ID:          types.StringValue(req.ID),
		WorkspaceID: types.StringValue(parts[0]),
		UserID:      types.StringValue(parts[1]),
	}
	// Set workspace_role to unknown so it gets populated by Read.
	state.WorkspaceRole = types.StringUnknown()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// flattenWorkspaceMember converts an API workspace member to the Terraform model.
func flattenWorkspaceMember(m *client.WorkspaceMember) workspaceMemberResourceModel {
	return workspaceMemberResourceModel{
		ID:            types.StringValue(m.WorkspaceID + "/" + m.UserID),
		WorkspaceID:   types.StringValue(m.WorkspaceID),
		UserID:        types.StringValue(m.UserID),
		WorkspaceRole: types.StringValue(m.WorkspaceRole),
	}
}
