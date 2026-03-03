package provider

import (
	"testing"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
)

func TestFlattenWorkspaceMember(t *testing.T) {
	tests := []struct {
		name       string
		member     *client.WorkspaceMember
		checkModel func(t *testing.T, m workspaceMemberResourceModel)
	}{
		{
			name: "basic member",
			member: &client.WorkspaceMember{
				Type:          "workspace_member",
				UserID:        "user-1",
				WorkspaceID:   "ws-1",
				WorkspaceRole: "workspace_developer",
			},
			checkModel: func(t *testing.T, m workspaceMemberResourceModel) {
				if m.ID.ValueString() != "ws-1/user-1" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "ws-1/user-1")
				}
				if m.WorkspaceID.ValueString() != "ws-1" {
					t.Errorf("WorkspaceID = %q, want %q", m.WorkspaceID.ValueString(), "ws-1")
				}
				if m.UserID.ValueString() != "user-1" {
					t.Errorf("UserID = %q, want %q", m.UserID.ValueString(), "user-1")
				}
				if m.WorkspaceRole.ValueString() != "workspace_developer" {
					t.Errorf("WorkspaceRole = %q, want %q", m.WorkspaceRole.ValueString(), "workspace_developer")
				}
			},
		},
		{
			name: "admin member",
			member: &client.WorkspaceMember{
				Type:          "workspace_member",
				UserID:        "user-2",
				WorkspaceID:   "ws-abc",
				WorkspaceRole: "workspace_admin",
			},
			checkModel: func(t *testing.T, m workspaceMemberResourceModel) {
				if m.ID.ValueString() != "ws-abc/user-2" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "ws-abc/user-2")
				}
				if m.WorkspaceRole.ValueString() != "workspace_admin" {
					t.Errorf("WorkspaceRole = %q, want %q", m.WorkspaceRole.ValueString(), "workspace_admin")
				}
			},
		},
		{
			name: "user role member",
			member: &client.WorkspaceMember{
				Type:          "workspace_member",
				UserID:        "user-3",
				WorkspaceID:   "ws-xyz",
				WorkspaceRole: "workspace_user",
			},
			checkModel: func(t *testing.T, m workspaceMemberResourceModel) {
				if m.ID.ValueString() != "ws-xyz/user-3" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "ws-xyz/user-3")
				}
				if m.WorkspaceRole.ValueString() != "workspace_user" {
					t.Errorf("WorkspaceRole = %q, want %q", m.WorkspaceRole.ValueString(), "workspace_user")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := flattenWorkspaceMember(tt.member)
			tt.checkModel(t, model)
		})
	}
}
