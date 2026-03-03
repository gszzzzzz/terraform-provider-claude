package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// User represents a Claude organization user.
type User struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	AddedAt string `json:"added_at"`
	Type    string `json:"type"`
}

// ListUsersParams are the optional query parameters for listing users.
type ListUsersParams struct {
	AfterID  string
	BeforeID string
	Email    string
	Limit    int
}

// ListUsersResponse is the paginated response from listing users.
type ListUsersResponse struct {
	Data    []User `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	Role string `json:"role"`
}

// DeleteUserResponse is the response from deleting a user.
type DeleteUserResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// GetUser gets a user by ID.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/organizations/users/%s", id), nil, &user)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return &user, nil
}

// ListUsers lists users with optional filters.
func (c *Client) ListUsers(ctx context.Context, params ListUsersParams) (*ListUsersResponse, error) {
	v := url.Values{}
	if params.AfterID != "" {
		v.Set("after_id", params.AfterID)
	}
	if params.BeforeID != "" {
		v.Set("before_id", params.BeforeID)
	}
	if params.Email != "" {
		v.Set("email", params.Email)
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}

	path := "/v1/organizations/users"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp ListUsersResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return &resp, nil
}

// UpdateUser updates a user's role.
func (c *Client) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/organizations/users/%s", id), req, &user)
	if err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}
	return &user, nil
}

// DeleteUser removes a user from the organization.
func (c *Client) DeleteUser(ctx context.Context, id string) (*DeleteUserResponse, error) {
	var resp DeleteUserResponse
	err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/v1/organizations/users/%s", id), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("deleting user: %w", err)
	}
	return &resp, nil
}
