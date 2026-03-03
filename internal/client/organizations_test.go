package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestGetOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/v1/organizations/me" {
				t.Errorf("path = %q, want /v1/organizations/me", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(Organization{
				ID:   "org-1",
				Name: "My Org",
			})
		})

		org, err := c.GetOrganization(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if org.ID != "org-1" {
			t.Errorf("ID = %q, want %q", org.ID, "org-1")
		}
		if org.Name != "My Org" {
			t.Errorf("Name = %q, want %q", org.Name, "My Org")
		}
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"type":"auth_error","message":"invalid key"}`))
		})

		org, err := c.GetOrganization(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if org != nil {
			t.Errorf("expected nil org, got %+v", org)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
		}
	})
}
