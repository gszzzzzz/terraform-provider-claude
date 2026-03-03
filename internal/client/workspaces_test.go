package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient("test-key", server.URL)
}

func TestCreateWorkspace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces" {
				t.Errorf("path = %q, want /v1/organizations/workspaces", r.URL.Path)
			}

			body, _ := io.ReadAll(r.Body)
			var req CreateWorkspaceRequest
			json.Unmarshal(body, &req)
			if req.Name != "my-workspace" {
				t.Errorf("Name = %q, want %q", req.Name, "my-workspace")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Workspace{
				ID:           "ws-1",
				Name:         "my-workspace",
				DisplayColor: "blue",
				CreatedAt:    "2024-01-01T00:00:00Z",
			})
		})

		ws, err := c.CreateWorkspace(context.Background(), CreateWorkspaceRequest{Name: "my-workspace"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ID != "ws-1" {
			t.Errorf("ID = %q, want %q", ws.ID, "ws-1")
		}
		if ws.Name != "my-workspace" {
			t.Errorf("Name = %q, want %q", ws.Name, "my-workspace")
		}
		if ws.DisplayColor != "blue" {
			t.Errorf("DisplayColor = %q, want %q", ws.DisplayColor, "blue")
		}
	})

	t.Run("with data residency", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req CreateWorkspaceRequest
			json.Unmarshal(body, &req)
			if req.DataResidency == nil {
				t.Fatal("expected data_residency in request")
			}
			if req.DataResidency.WorkspaceGeo != "us" {
				t.Errorf("WorkspaceGeo = %q, want %q", req.DataResidency.WorkspaceGeo, "us")
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"ws-2","name":"geo-ws","display_color":"red","created_at":"2024-01-01T00:00:00Z","data_residency":{"workspace_geo":"us","default_inference_geo":"us","allowed_inference_geos":["us"]}}`))
		})

		ws, err := c.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
			Name: "geo-ws",
			DataResidency: &CreateDataResidencyRequest{
				WorkspaceGeo: "us",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.DataResidency == nil {
			t.Fatal("expected DataResidency in response")
		}
		if ws.DataResidency.WorkspaceGeo != "us" {
			t.Errorf("WorkspaceGeo = %q, want %q", ws.DataResidency.WorkspaceGeo, "us")
		}
	})

	t.Run("api error", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"type":"invalid_request","message":"bad"}`))
		})

		ws, err := c.CreateWorkspace(context.Background(), CreateWorkspaceRequest{Name: "bad"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if ws != nil {
			t.Errorf("expected nil workspace, got %+v", ws)
		}
	})
}

func TestGetWorkspace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"ws-1","name":"test","display_color":"green","created_at":"2024-01-01T00:00:00Z","archived_at":null}`))
		})

		ws, err := c.GetWorkspace(context.Background(), "ws-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ID != "ws-1" {
			t.Errorf("ID = %q, want %q", ws.ID, "ws-1")
		}
		if ws.ArchivedAt != nil {
			t.Errorf("ArchivedAt = %v, want nil", ws.ArchivedAt)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"type":"not_found","message":"workspace not found"}`))
		})

		_, err := c.GetWorkspace(context.Background(), "ws-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true, got false")
		}
	})

	t.Run("archived workspace", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"ws-1","name":"archived","display_color":"gray","created_at":"2024-01-01T00:00:00Z","archived_at":"2024-06-01T00:00:00Z"}`))
		})

		ws, err := c.GetWorkspace(context.Background(), "ws-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ArchivedAt == nil {
			t.Fatal("expected ArchivedAt to be non-nil")
		}
		if *ws.ArchivedAt != "2024-06-01T00:00:00Z" {
			t.Errorf("ArchivedAt = %q, want %q", *ws.ArchivedAt, "2024-06-01T00:00:00Z")
		}
	})

	t.Run("string allowed_inference_geos", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"ws-1","name":"test","display_color":"blue","created_at":"2024-01-01T00:00:00Z","data_residency":{"workspace_geo":"us","default_inference_geo":"us","allowed_inference_geos":"unrestricted"}}`))
		})

		ws, err := c.GetWorkspace(context.Background(), "ws-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.DataResidency == nil {
			t.Fatal("expected DataResidency to be non-nil")
		}
		// allowed_inference_geos is a json.RawMessage; verify it's a string
		var s string
		if err := json.Unmarshal(ws.DataResidency.AllowedInferenceGeos, &s); err != nil {
			t.Errorf("expected string, got unmarshal error: %v", err)
		}
		if s != "unrestricted" {
			t.Errorf("AllowedInferenceGeos = %q, want %q", s, "unrestricted")
		}
	})

	t.Run("array allowed_inference_geos", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"ws-1","name":"test","display_color":"blue","created_at":"2024-01-01T00:00:00Z","data_residency":{"workspace_geo":"eu","default_inference_geo":"eu","allowed_inference_geos":["us","eu"]}}`))
		})

		ws, err := c.GetWorkspace(context.Background(), "ws-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.DataResidency == nil {
			t.Fatal("expected DataResidency to be non-nil")
		}
		var arr []string
		if err := json.Unmarshal(ws.DataResidency.AllowedInferenceGeos, &arr); err != nil {
			t.Errorf("expected array, got unmarshal error: %v", err)
		}
		if len(arr) != 2 || arr[0] != "us" || arr[1] != "eu" {
			t.Errorf("AllowedInferenceGeos = %v, want [us eu]", arr)
		}
	})
}

func TestUpdateWorkspace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1", r.URL.Path)
			}

			body, _ := io.ReadAll(r.Body)
			var req UpdateWorkspaceRequest
			json.Unmarshal(body, &req)
			if req.Name != "updated" {
				t.Errorf("Name = %q, want %q", req.Name, "updated")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Workspace{
				ID:           "ws-1",
				Name:         "updated",
				DisplayColor: "blue",
				CreatedAt:    "2024-01-01T00:00:00Z",
			})
		})

		ws, err := c.UpdateWorkspace(context.Background(), "ws-1", UpdateWorkspaceRequest{Name: "updated"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.Name != "updated" {
			t.Errorf("Name = %q, want %q", ws.Name, "updated")
		}
	})

	t.Run("api error", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"type":"invalid_request","message":"bad"}`))
		})

		_, err := c.UpdateWorkspace(context.Background(), "ws-1", UpdateWorkspaceRequest{Name: "bad"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestArchiveWorkspace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		archived := "2024-06-01T00:00:00Z"
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/v1/organizations/workspaces/ws-1/archive" {
				t.Errorf("path = %q, want /v1/organizations/workspaces/ws-1/archive", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Workspace{
				ID:           "ws-1",
				Name:         "archived",
				DisplayColor: "gray",
				CreatedAt:    "2024-01-01T00:00:00Z",
				ArchivedAt:   &archived,
			})
		})

		ws, err := c.ArchiveWorkspace(context.Background(), "ws-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ArchivedAt == nil {
			t.Fatal("expected ArchivedAt to be non-nil")
		}
		if *ws.ArchivedAt != archived {
			t.Errorf("ArchivedAt = %q, want %q", *ws.ArchivedAt, archived)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"type":"not_found","message":"not found"}`))
		})

		_, err := c.ArchiveWorkspace(context.Background(), "ws-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true")
		}
	})
}
