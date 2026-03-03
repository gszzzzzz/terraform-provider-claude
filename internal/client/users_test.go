package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/v1/organizations/users/user-1" {
				t.Errorf("path = %q, want /v1/organizations/users/user-1", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(User{
				ID:      "user-1",
				Email:   "alice@example.com",
				Name:    "Alice",
				Role:    "developer",
				AddedAt: "2024-01-01T00:00:00Z",
				Type:    "user",
			})
		})

		user, err := c.GetUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != "user-1" {
			t.Errorf("ID = %q, want %q", user.ID, "user-1")
		}
		if user.Email != "alice@example.com" {
			t.Errorf("Email = %q, want %q", user.Email, "alice@example.com")
		}
		if user.Name != "Alice" {
			t.Errorf("Name = %q, want %q", user.Name, "Alice")
		}
		if user.Role != "developer" {
			t.Errorf("Role = %q, want %q", user.Role, "developer")
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"not_found","message":"user not found"}`))
		})

		_, err := c.GetUser(context.Background(), "user-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true, got false")
		}
	})
}

func TestListUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/v1/organizations/users" {
				t.Errorf("path = %q, want /v1/organizations/users", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListUsersResponse{
				Data: []User{
					{ID: "user-1", Email: "alice@example.com", Name: "Alice", Role: "developer"},
					{ID: "user-2", Email: "bob@example.com", Name: "Bob", Role: "admin"},
				},
				FirstID: "user-1",
				LastID:  "user-2",
				HasMore: false,
			})
		})

		resp, err := c.ListUsers(context.Background(), ListUsersParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 2 {
			t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
		}
		if resp.Data[0].ID != "user-1" {
			t.Errorf("Data[0].ID = %q, want %q", resp.Data[0].ID, "user-1")
		}
		if resp.HasMore {
			t.Errorf("HasMore = true, want false")
		}
	})

	t.Run("email filter", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("email") != "alice@example.com" {
				t.Errorf("email query = %q, want %q", r.URL.Query().Get("email"), "alice@example.com")
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListUsersResponse{
				Data: []User{
					{ID: "user-1", Email: "alice@example.com", Name: "Alice", Role: "developer"},
				},
				FirstID: "user-1",
				LastID:  "user-1",
				HasMore: false,
			})
		})

		resp, err := c.ListUsers(context.Background(), ListUsersParams{Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
		}
	})

	t.Run("empty result", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListUsersResponse{
				Data:    []User{},
				HasMore: false,
			})
		})

		resp, err := c.ListUsers(context.Background(), ListUsersParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 0 {
			t.Errorf("len(Data) = %d, want 0", len(resp.Data))
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/v1/organizations/users/user-1" {
				t.Errorf("path = %q, want /v1/organizations/users/user-1", r.URL.Path)
			}

			body, _ := io.ReadAll(r.Body)
			var req UpdateUserRequest
			_ = json.Unmarshal(body, &req)
			if req.Role != "admin" {
				t.Errorf("Role = %q, want %q", req.Role, "admin")
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(User{
				ID:      "user-1",
				Email:   "alice@example.com",
				Name:    "Alice",
				Role:    "admin",
				AddedAt: "2024-01-01T00:00:00Z",
				Type:    "user",
			})
		})

		user, err := c.UpdateUser(context.Background(), "user-1", UpdateUserRequest{Role: "admin"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Role != "admin" {
			t.Errorf("Role = %q, want %q", user.Role, "admin")
		}
	})

	t.Run("api error", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"type":"invalid_request","message":"invalid role"}`))
		})

		_, err := c.UpdateUser(context.Background(), "user-1", UpdateUserRequest{Role: "invalid"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			if r.URL.Path != "/v1/organizations/users/user-1" {
				t.Errorf("path = %q, want /v1/organizations/users/user-1", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(DeleteUserResponse{
				ID:   "user-1",
				Type: "user_deleted",
			})
		})

		resp, err := c.DeleteUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != "user-1" {
			t.Errorf("ID = %q, want %q", resp.ID, "user-1")
		}
		if resp.Type != "user_deleted" {
			t.Errorf("Type = %q, want %q", resp.Type, "user_deleted")
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"not_found","message":"user not found"}`))
		})

		_, err := c.DeleteUser(context.Background(), "user-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true, got false")
		}
	})
}
