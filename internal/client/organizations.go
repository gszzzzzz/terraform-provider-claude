package client

import (
	"context"
	"fmt"
	"net/http"
)

// Organization represents a Claude organization.
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetOrganization gets the current organization.
func (c *Client) GetOrganization(ctx context.Context) (*Organization, error) {
	var org Organization
	err := c.doRequest(ctx, http.MethodGet, "/v1/organizations/me", nil, &org)
	if err != nil {
		return nil, fmt.Errorf("getting organization: %w", err)
	}
	return &org, nil
}
