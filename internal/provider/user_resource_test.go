package provider

import (
	"testing"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
)

func TestFlattenUser(t *testing.T) {
	tests := []struct {
		name       string
		user       *client.User
		checkModel func(t *testing.T, m userResourceModel)
	}{
		{
			name: "basic user",
			user: &client.User{
				ID:      "user-1",
				Email:   "alice@example.com",
				Name:    "Alice",
				Role:    "developer",
				AddedAt: "2024-01-01T00:00:00Z",
				Type:    "user",
			},
			checkModel: func(t *testing.T, m userResourceModel) {
				if m.ID.ValueString() != "user-1" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "user-1")
				}
				if m.Email.ValueString() != "alice@example.com" {
					t.Errorf("Email = %q, want %q", m.Email.ValueString(), "alice@example.com")
				}
				if m.Name.ValueString() != "Alice" {
					t.Errorf("Name = %q, want %q", m.Name.ValueString(), "Alice")
				}
				if m.Role.ValueString() != "developer" {
					t.Errorf("Role = %q, want %q", m.Role.ValueString(), "developer")
				}
				if m.AddedAt.ValueString() != "2024-01-01T00:00:00Z" {
					t.Errorf("AddedAt = %q, want %q", m.AddedAt.ValueString(), "2024-01-01T00:00:00Z")
				}
			},
		},
		{
			name: "admin user",
			user: &client.User{
				ID:      "user-2",
				Email:   "bob@example.com",
				Name:    "Bob",
				Role:    "admin",
				AddedAt: "2024-06-15T12:30:00Z",
				Type:    "user",
			},
			checkModel: func(t *testing.T, m userResourceModel) {
				if m.ID.ValueString() != "user-2" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "user-2")
				}
				if m.Role.ValueString() != "admin" {
					t.Errorf("Role = %q, want %q", m.Role.ValueString(), "admin")
				}
			},
		},
		{
			name: "empty fields",
			user: &client.User{
				ID:   "user-3",
				Role: "user",
			},
			checkModel: func(t *testing.T, m userResourceModel) {
				if m.ID.ValueString() != "user-3" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "user-3")
				}
				if m.Email.ValueString() != "" {
					t.Errorf("Email = %q, want empty", m.Email.ValueString())
				}
				if m.Name.ValueString() != "" {
					t.Errorf("Name = %q, want empty", m.Name.ValueString())
				}
				if m.Role.ValueString() != "user" {
					t.Errorf("Role = %q, want %q", m.Role.ValueString(), "user")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := flattenUser(tt.user)
			tt.checkModel(t, model)
		})
	}
}
