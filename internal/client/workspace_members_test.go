package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateWorkspaceMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1/members" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1/members", r.URL.Path)
			}

			body, _ := io.ReadAll(r.Body)
			var req CreateWorkspaceMemberRequest
			_ = json.Unmarshal(body, &req)
			if req.UserID != "user-1" {
				t.Errorf("UserID = %q, want %q", req.UserID, "user-1")
			}
			if req.WorkspaceRole != "workspace_developer" {
				t.Errorf("WorkspaceRole = %q, want %q", req.WorkspaceRole, "workspace_developer")
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(WorkspaceMember{
				Type:          "workspace_member",
				UserID:        "user-1",
				WorkspaceID:   "ws-1",
				WorkspaceRole: "workspace_developer",
			})
		})

		member, err := c.CreateWorkspaceMember(context.Background(), "ws-1", CreateWorkspaceMemberRequest{
			UserID:        "user-1",
			WorkspaceRole: "workspace_developer",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if member.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", member.UserID, "user-1")
		}
		if member.WorkspaceID != "ws-1" {
			t.Errorf("WorkspaceID = %q, want %q", member.WorkspaceID, "ws-1")
		}
		if member.WorkspaceRole != "workspace_developer" {
			t.Errorf("WorkspaceRole = %q, want %q", member.WorkspaceRole, "workspace_developer")
		}
	})

	t.Run("api error", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"type":"invalid_request","message":"invalid role"}`))
		})

		_, err := c.CreateWorkspaceMember(context.Background(), "ws-1", CreateWorkspaceMemberRequest{
			UserID:        "user-1",
			WorkspaceRole: "invalid",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetWorkspaceMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1/members/user-1" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1/members/user-1", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(WorkspaceMember{
				Type:          "workspace_member",
				UserID:        "user-1",
				WorkspaceID:   "ws-1",
				WorkspaceRole: "workspace_admin",
			})
		})

		member, err := c.GetWorkspaceMember(context.Background(), "ws-1", "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if member.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", member.UserID, "user-1")
		}
		if member.WorkspaceRole != "workspace_admin" {
			t.Errorf("WorkspaceRole = %q, want %q", member.WorkspaceRole, "workspace_admin")
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"not_found","message":"member not found"}`))
		})

		_, err := c.GetWorkspaceMember(context.Background(), "ws-1", "user-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true, got false")
		}
	})
}

func TestListWorkspaceMembers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1/members" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1/members", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListWorkspaceMembersResponse{
				Data: []WorkspaceMember{
					{Type: "workspace_member", UserID: "user-1", WorkspaceID: "ws-1", WorkspaceRole: "workspace_admin"},
					{Type: "workspace_member", UserID: "user-2", WorkspaceID: "ws-1", WorkspaceRole: "workspace_user"},
				},
				FirstID: "user-1",
				LastID:  "user-2",
				HasMore: false,
			})
		})

		resp, err := c.ListWorkspaceMembers(context.Background(), "ws-1", ListWorkspaceMembersParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 2 {
			t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
		}
		if resp.Data[0].UserID != "user-1" {
			t.Errorf("Data[0].UserID = %q, want %q", resp.Data[0].UserID, "user-1")
		}
		if resp.HasMore {
			t.Errorf("HasMore = true, want false")
		}
	})

	t.Run("with pagination params", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("limit") != "10" {
				t.Errorf("limit = %q, want %q", r.URL.Query().Get("limit"), "10")
			}
			if r.URL.Query().Get("after_id") != "user-1" {
				t.Errorf("after_id = %q, want %q", r.URL.Query().Get("after_id"), "user-1")
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListWorkspaceMembersResponse{
				Data:    []WorkspaceMember{},
				HasMore: false,
			})
		})

		_, err := c.ListWorkspaceMembers(context.Background(), "ws-1", ListWorkspaceMembersParams{
			AfterID: "user-1",
			Limit:   10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListWorkspaceMembersResponse{
				Data:    []WorkspaceMember{},
				HasMore: false,
			})
		})

		resp, err := c.ListWorkspaceMembers(context.Background(), "ws-1", ListWorkspaceMembersParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 0 {
			t.Errorf("len(Data) = %d, want 0", len(resp.Data))
		}
	})
}

func TestUpdateWorkspaceMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1/members/user-1" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1/members/user-1", r.URL.Path)
			}

			body, _ := io.ReadAll(r.Body)
			var req UpdateWorkspaceMemberRequest
			_ = json.Unmarshal(body, &req)
			if req.WorkspaceRole != "workspace_admin" {
				t.Errorf("WorkspaceRole = %q, want %q", req.WorkspaceRole, "workspace_admin")
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(WorkspaceMember{
				Type:          "workspace_member",
				UserID:        "user-1",
				WorkspaceID:   "ws-1",
				WorkspaceRole: "workspace_admin",
			})
		})

		member, err := c.UpdateWorkspaceMember(context.Background(), "ws-1", "user-1", UpdateWorkspaceMemberRequest{
			WorkspaceRole: "workspace_admin",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if member.WorkspaceRole != "workspace_admin" {
			t.Errorf("WorkspaceRole = %q, want %q", member.WorkspaceRole, "workspace_admin")
		}
	})

	t.Run("api error", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"type":"invalid_request","message":"invalid role"}`))
		})

		_, err := c.UpdateWorkspaceMember(context.Background(), "ws-1", "user-1", UpdateWorkspaceMemberRequest{
			WorkspaceRole: "invalid",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDeleteWorkspaceMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1/members/user-1" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1/members/user-1", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(DeleteWorkspaceMemberResponse{
				Type:        "workspace_member_deleted",
				UserID:      "user-1",
				WorkspaceID: "ws-1",
			})
		})

		resp, err := c.DeleteWorkspaceMember(context.Background(), "ws-1", "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", resp.UserID, "user-1")
		}
		if resp.Type != "workspace_member_deleted" {
			t.Errorf("Type = %q, want %q", resp.Type, "workspace_member_deleted")
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"not_found","message":"member not found"}`))
		})

		_, err := c.DeleteWorkspaceMember(context.Background(), "ws-1", "user-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true, got false")
		}
	})
}
