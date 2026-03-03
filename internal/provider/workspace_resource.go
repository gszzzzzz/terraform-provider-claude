package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &workspaceResource{}
	_ resource.ResourceWithImportState = &workspaceResource{}
)

type workspaceResource struct {
	client *client.Client
}

type workspaceResourceModel struct {
	ID            types.String        `tfsdk:"id"`
	Name          types.String        `tfsdk:"name"`
	DisplayColor  types.String        `tfsdk:"display_color"`
	CreatedAt     types.String        `tfsdk:"created_at"`
	ArchivedAt    types.String        `tfsdk:"archived_at"`
	DataResidency *dataResidencyModel `tfsdk:"data_residency"`
}

type dataResidencyModel struct {
	WorkspaceGeo         types.String `tfsdk:"workspace_geo"`
	DefaultInferenceGeo  types.String `tfsdk:"default_inference_geo"`
	AllowedInferenceGeos types.List   `tfsdk:"allowed_inference_geos"`
}

func NewWorkspaceResource() resource.Resource {
	return &workspaceResource{}
}

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Claude workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the workspace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the workspace.",
			},
			"display_color": schema.StringAttribute{
				Computed:    true,
				Description: "The display color of the workspace, set by the API.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the workspace was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"archived_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the workspace was archived, or null if active.",
			},
			"data_residency": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Data residency configuration for the workspace.",
				Attributes: map[string]schema.Attribute{
					"workspace_geo": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The geographic region for workspace data. Cannot be changed after creation.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"default_inference_geo": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The default geographic region for inference.",
					},
					"allowed_inference_geos": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "The allowed geographic regions for inference. Use [\"unrestricted\"] to allow all regions.",
					},
				},
			},
		},
	}
}

func (r *workspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateWorkspaceRequest{
		Name: plan.Name.ValueString(),
	}

	if plan.DataResidency != nil {
		dr := &client.CreateDataResidencyRequest{}
		if !plan.DataResidency.WorkspaceGeo.IsNull() && !plan.DataResidency.WorkspaceGeo.IsUnknown() {
			dr.WorkspaceGeo = plan.DataResidency.WorkspaceGeo.ValueString()
		}
		if !plan.DataResidency.DefaultInferenceGeo.IsNull() && !plan.DataResidency.DefaultInferenceGeo.IsUnknown() {
			dr.DefaultInferenceGeo = plan.DataResidency.DefaultInferenceGeo.ValueString()
		}
		if !plan.DataResidency.AllowedInferenceGeos.IsNull() && !plan.DataResidency.AllowedInferenceGeos.IsUnknown() {
			var geos []string
			resp.Diagnostics.Append(plan.DataResidency.AllowedInferenceGeos.ElementsAs(ctx, &geos, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			dr.AllowedInferenceGeos = normalizeAllowedGeosForAPI(geos)
		}
		createReq.DataResidency = dr
	}

	workspace, err := r.client.CreateWorkspace(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create workspace", err.Error())
		return
	}

	state := flattenWorkspace(ctx, workspace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace, err := r.client.GetWorkspace(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read workspace", err.Error())
		return
	}

	if workspace.ArchivedAt != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := flattenWorkspace(ctx, workspace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateWorkspaceRequest{
		Name: plan.Name.ValueString(),
	}

	if plan.DataResidency != nil {
		if !plan.DataResidency.DefaultInferenceGeo.IsNull() && !plan.DataResidency.DefaultInferenceGeo.IsUnknown() {
			updateReq.DefaultInferenceGeo = plan.DataResidency.DefaultInferenceGeo.ValueString()
		}
		if !plan.DataResidency.AllowedInferenceGeos.IsNull() && !plan.DataResidency.AllowedInferenceGeos.IsUnknown() {
			var geos []string
			resp.Diagnostics.Append(plan.DataResidency.AllowedInferenceGeos.ElementsAs(ctx, &geos, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateReq.AllowedInferenceGeos = normalizeAllowedGeosForAPI(geos)
		}
	}

	workspace, err := r.client.UpdateWorkspace(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update workspace", err.Error())
		return
	}

	state := flattenWorkspace(ctx, workspace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ArchiveWorkspace(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to archive workspace", err.Error())
	}
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// flattenWorkspace converts an API workspace to the Terraform model.
func flattenWorkspace(_ context.Context, w *client.Workspace) workspaceResourceModel {
	model := workspaceResourceModel{
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
		dr.AllowedInferenceGeos = parseAllowedInferenceGeos(w.DataResidency.AllowedInferenceGeos)
		model.DataResidency = dr
	}

	return model
}

// parseAllowedInferenceGeos handles the union type from the API:
// - string "unrestricted" → ["unrestricted"]
// - array ["us", "eu"] → ["us", "eu"]
func parseAllowedInferenceGeos(raw json.RawMessage) types.List {
	if raw == nil {
		return types.ListNull(types.StringType)
	}

	// Try as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		elems := []types.String{types.StringValue(str)}
		list, _ := types.ListValueFrom(context.Background(), types.StringType, elems)
		return list
	}

	// Try as array
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		list, _ := types.ListValueFrom(context.Background(), types.StringType, arr)
		return list
	}

	return types.ListNull(types.StringType)
}

// normalizeAllowedGeosForAPI converts Terraform list to API format:
// - ["unrestricted"] → "unrestricted" (string)
// - ["us", "eu"] → ["us", "eu"] (array)
func normalizeAllowedGeosForAPI(geos []string) any {
	if len(geos) == 1 && geos[0] == "unrestricted" {
		return "unrestricted"
	}
	return geos
}
