package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// WorkspaceMember represents a Claude workspace member.
type WorkspaceMember struct {
	Type          string `json:"type"`
	UserID        string `json:"user_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceRole string `json:"workspace_role"`
}

// CreateWorkspaceMemberRequest is the request body for adding a member to a workspace.
type CreateWorkspaceMemberRequest struct {
	UserID        string `json:"user_id"`
	WorkspaceRole string `json:"workspace_role"`
}

// UpdateWorkspaceMemberRequest is the request body for updating a workspace member's role.
type UpdateWorkspaceMemberRequest struct {
	WorkspaceRole string `json:"workspace_role"`
}

// DeleteWorkspaceMemberResponse is the response from removing a member from a workspace.
type DeleteWorkspaceMemberResponse struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id"`
	WorkspaceID string `json:"workspace_id"`
}

// ListWorkspaceMembersParams are the optional query parameters for listing workspace members.
type ListWorkspaceMembersParams struct {
	AfterID  string
	BeforeID string
	Limit    int
}

// ListWorkspaceMembersResponse is the paginated response from listing workspace members.
type ListWorkspaceMembersResponse struct {
	Data    []WorkspaceMember `json:"data"`
	FirstID string            `json:"first_id"`
	LastID  string            `json:"last_id"`
	HasMore bool              `json:"has_more"`
}

// CreateWorkspaceMember adds a member to a workspace.
func (c *Client) CreateWorkspaceMember(ctx context.Context, workspaceID string, req CreateWorkspaceMemberRequest) (*WorkspaceMember, error) {
	var member WorkspaceMember
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/organizations/workspaces/%s/members", workspaceID), req, &member)
	if err != nil {
		return nil, fmt.Errorf("creating workspace member: %w", err)
	}
	return &member, nil
}

// GetWorkspaceMember gets a workspace member by workspace ID and user ID.
func (c *Client) GetWorkspaceMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error) {
	var member WorkspaceMember
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/organizations/workspaces/%s/members/%s", workspaceID, userID), nil, &member)
	if err != nil {
		return nil, fmt.Errorf("getting workspace member: %w", err)
	}
	return &member, nil
}

// ListWorkspaceMembers lists members of a workspace with optional pagination.
func (c *Client) ListWorkspaceMembers(ctx context.Context, workspaceID string, params ListWorkspaceMembersParams) (*ListWorkspaceMembersResponse, error) {
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

	path := fmt.Sprintf("/v1/organizations/workspaces/%s/members", workspaceID)
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp ListWorkspaceMembersResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing workspace members: %w", err)
	}
	return &resp, nil
}

// UpdateWorkspaceMember updates a workspace member's role.
func (c *Client) UpdateWorkspaceMember(ctx context.Context, workspaceID, userID string, req UpdateWorkspaceMemberRequest) (*WorkspaceMember, error) {
	var member WorkspaceMember
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/organizations/workspaces/%s/members/%s", workspaceID, userID), req, &member)
	if err != nil {
		return nil, fmt.Errorf("updating workspace member: %w", err)
	}
	return &member, nil
}

// DeleteWorkspaceMember removes a member from a workspace.
func (c *Client) DeleteWorkspaceMember(ctx context.Context, workspaceID, userID string) (*DeleteWorkspaceMemberResponse, error) {
	var resp DeleteWorkspaceMemberResponse
	err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/v1/organizations/workspaces/%s/members/%s", workspaceID, userID), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("deleting workspace member: %w", err)
	}
	return &resp, nil
}
