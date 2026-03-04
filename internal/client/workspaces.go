package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Workspace represents a Claude workspace.
type Workspace struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	DisplayColor  string         `json:"display_color"`
	CreatedAt     string         `json:"created_at"`
	ArchivedAt    *string        `json:"archived_at"`
	DataResidency *DataResidency `json:"data_residency,omitempty"`
}

// DataResidency represents the data residency configuration for a workspace.
type DataResidency struct {
	WorkspaceGeo         string          `json:"workspace_geo"`
	DefaultInferenceGeo  string          `json:"default_inference_geo"`
	AllowedInferenceGeos json.RawMessage `json:"allowed_inference_geos"`
}

// CreateWorkspaceRequest is the request body for creating a workspace.
type CreateWorkspaceRequest struct {
	Name          string                      `json:"name"`
	DataResidency *CreateDataResidencyRequest `json:"data_residency,omitempty"`
}

// CreateDataResidencyRequest is the data residency config for workspace creation.
type CreateDataResidencyRequest struct {
	WorkspaceGeo         string `json:"workspace_geo,omitempty"`
	DefaultInferenceGeo  string `json:"default_inference_geo,omitempty"`
	AllowedInferenceGeos any    `json:"allowed_inference_geos,omitempty"`
}

// UpdateWorkspaceRequest is the request body for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name          string                      `json:"name,omitempty"`
	DataResidency *UpdateDataResidencyRequest `json:"data_residency,omitempty"`
}

// UpdateDataResidencyRequest is the data residency config for workspace updates.
type UpdateDataResidencyRequest struct {
	DefaultInferenceGeo  string `json:"default_inference_geo,omitempty"`
	AllowedInferenceGeos any    `json:"allowed_inference_geos,omitempty"`
}

// ListWorkspacesParams are the optional query parameters for listing workspaces.
type ListWorkspacesParams struct {
	AfterID         string
	BeforeID        string
	Limit           int
	IncludeArchived bool
}

// ListWorkspacesResponse is the paginated response from listing workspaces.
type ListWorkspacesResponse struct {
	Data    []Workspace `json:"data"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
	HasMore bool        `json:"has_more"`
}

// ListWorkspaces lists workspaces with optional filters.
func (c *Client) ListWorkspaces(ctx context.Context, params ListWorkspacesParams) (*ListWorkspacesResponse, error) {
	v := url.Values{}
	if params.AfterID != "" {
		v.Set("after_id", params.AfterID)
	}
	if params.BeforeID != "" {
		v.Set("before_id", params.BeforeID)
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.IncludeArchived {
		v.Set("include_archived", "true")
	}

	path := "/v1/organizations/workspaces"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp ListWorkspacesResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}
	return &resp, nil
}

// CreateWorkspace creates a new workspace.
func (c *Client) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (*Workspace, error) {
	var workspace Workspace
	err := c.doRequest(ctx, http.MethodPost, "/v1/organizations/workspaces", req, &workspace)
	if err != nil {
		return nil, fmt.Errorf("creating workspace: %w", err)
	}
	return &workspace, nil
}

// GetWorkspace gets a workspace by ID.
func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var workspace Workspace
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/organizations/workspaces/%s", id), nil, &workspace)
	if err != nil {
		return nil, fmt.Errorf("getting workspace: %w", err)
	}
	return &workspace, nil
}

// UpdateWorkspace updates a workspace.
func (c *Client) UpdateWorkspace(ctx context.Context, id string, req UpdateWorkspaceRequest) (*Workspace, error) {
	var workspace Workspace
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/organizations/workspaces/%s", id), req, &workspace)
	if err != nil {
		return nil, fmt.Errorf("updating workspace: %w", err)
	}
	return &workspace, nil
}

// ArchiveWorkspace archives (soft-deletes) a workspace.
func (c *Client) ArchiveWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var workspace Workspace
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/organizations/workspaces/%s/archive", id), nil, &workspace)
	if err != nil {
		return nil, fmt.Errorf("archiving workspace: %w", err)
	}
	return &workspace, nil
}
