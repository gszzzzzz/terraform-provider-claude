package provider

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// mockWorkspaceMember is the in-memory representation used by the mock handler.
type mockWorkspaceMember struct {
	Type          string `json:"type"`
	UserID        string `json:"user_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceRole string `json:"workspace_role"`
}

// workspaceMemberStore is a concurrency-safe in-memory store for mock workspace members.
type workspaceMemberStore struct {
	mu    sync.Mutex
	items map[string]*mockWorkspaceMember // key: "workspace_id/user_id"
}

func newWorkspaceMemberStore() *workspaceMemberStore {
	return &workspaceMemberStore{
		items: make(map[string]*mockWorkspaceMember),
	}
}

func workspaceMemberKey(workspaceID, userID string) string {
	return workspaceID + "/" + userID
}

// newWorkspaceMemberMockMux creates an http.ServeMux that handles workspace member CRUD
// backed by the given store.
func newWorkspaceMemberMockMux(t *testing.T, store *workspaceMemberStore) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()

	// Create
	mux.HandleFunc("POST /v1/organizations/workspaces/{workspace_id}/members", func(w http.ResponseWriter, r *http.Request) {
		workspaceID := r.PathValue("workspace_id")

		var body struct {
			UserID        string `json:"user_id"`
			WorkspaceRole string `json:"workspace_role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"type":"invalid_request","message":"bad json"}`, http.StatusBadRequest)
			return
		}

		key := workspaceMemberKey(workspaceID, body.UserID)

		store.mu.Lock()
		if _, exists := store.items[key]; exists {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "conflict",
				"message": "member already exists",
			})
			return
		}

		m := &mockWorkspaceMember{
			Type:          "workspace_member",
			UserID:        body.UserID,
			WorkspaceID:   workspaceID,
			WorkspaceRole: body.WorkspaceRole,
		}
		store.items[key] = m
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m)
	})

	// Read
	mux.HandleFunc("GET /v1/organizations/workspaces/{workspace_id}/members/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		key := workspaceMemberKey(r.PathValue("workspace_id"), r.PathValue("user_id"))

		store.mu.Lock()
		m, ok := store.items[key]
		store.mu.Unlock()

		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "member not found",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m)
	})

	// Update
	mux.HandleFunc("POST /v1/organizations/workspaces/{workspace_id}/members/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		key := workspaceMemberKey(r.PathValue("workspace_id"), r.PathValue("user_id"))

		store.mu.Lock()
		m, ok := store.items[key]
		if !ok {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "member not found",
			})
			return
		}

		var body struct {
			WorkspaceRole string `json:"workspace_role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			store.mu.Unlock()
			http.Error(w, `{"type":"invalid_request","message":"bad json"}`, http.StatusBadRequest)
			return
		}

		if body.WorkspaceRole != "" {
			m.WorkspaceRole = body.WorkspaceRole
		}
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m)
	})

	// Delete
	mux.HandleFunc("DELETE /v1/organizations/workspaces/{workspace_id}/members/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		workspaceID := r.PathValue("workspace_id")
		userID := r.PathValue("user_id")
		key := workspaceMemberKey(workspaceID, userID)

		store.mu.Lock()
		_, ok := store.items[key]
		if !ok {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "member not found",
			})
			return
		}

		delete(store.items, key)
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"type":         "workspace_member_deleted",
			"user_id":      userID,
			"workspace_id": workspaceID,
		})
	})

	return mux
}

func TestAccWorkspaceMemberResource_basic(t *testing.T) {
	store := newWorkspaceMemberStore()
	mux := newWorkspaceMemberMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create workspace member
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace_member" "test" {
  workspace_id   = "ws-test1"
  user_id        = "user-test1"
  workspace_role = "workspace_developer"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_workspace_member.test", "id", "ws-test1/user-test1"),
					resource.TestCheckResourceAttr("claude_workspace_member.test", "workspace_id", "ws-test1"),
					resource.TestCheckResourceAttr("claude_workspace_member.test", "user_id", "user-test1"),
					resource.TestCheckResourceAttr("claude_workspace_member.test", "workspace_role", "workspace_developer"),
				),
			},
			// Step 2: Update role
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace_member" "test" {
  workspace_id   = "ws-test1"
  user_id        = "user-test1"
  workspace_role = "workspace_admin"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_workspace_member.test", "id", "ws-test1/user-test1"),
					resource.TestCheckResourceAttr("claude_workspace_member.test", "workspace_role", "workspace_admin"),
				),
			},
		},
	})
}

func TestAccWorkspaceMemberResource_import(t *testing.T) {
	store := newWorkspaceMemberStore()
	mux := newWorkspaceMemberMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create a member first
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace_member" "import_test" {
  workspace_id   = "ws-imp1"
  user_id        = "user-imp1"
  workspace_role = "workspace_user"
}
`,
			},
			// Step 2: Import and verify
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace_member" "import_test" {
  workspace_id   = "ws-imp1"
  user_id        = "user-imp1"
  workspace_role = "workspace_user"
}
`,
				ResourceName:      "claude_workspace_member.import_test",
				ImportState:       true,
				ImportStateId:     "ws-imp1/user-imp1",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkspaceMemberResource_readDeletedRemovesFromState(t *testing.T) {
	store := newWorkspaceMemberStore()
	mux := newWorkspaceMemberMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create a member
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace_member" "del_test" {
  workspace_id   = "ws-del1"
  user_id        = "user-del1"
  workspace_role = "workspace_developer"
}
`,
			},
			// Step 2: Delete externally, then refresh → Read removes from state
			{
				PreConfig: func() {
					store.mu.Lock()
					defer store.mu.Unlock()
					delete(store.items, workspaceMemberKey("ws-del1", "user-del1"))
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
