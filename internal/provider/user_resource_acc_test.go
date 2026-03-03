package provider

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// mockUser is the in-memory representation used by the mock handler.
type mockUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	AddedAt string `json:"added_at"`
	Type    string `json:"type"`
}

// userStore is a concurrency-safe in-memory store for mock users.
type userStore struct {
	mu    sync.Mutex
	items map[string]*mockUser
}

func newUserStore() *userStore {
	return &userStore{
		items: make(map[string]*mockUser),
	}
}

// seedUser adds a pre-existing user to the store (since there is no Create API).
func (s *userStore) seedUser(u *mockUser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[u.ID] = u
}

// newUserMockMux creates an http.ServeMux that handles user Read/Update/Delete
// backed by the given store.
func newUserMockMux(t *testing.T, store *userStore) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()

	// Read
	mux.HandleFunc("GET /v1/organizations/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		store.mu.Lock()
		u, ok := store.items[id]
		store.mu.Unlock()

		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "user not found",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u)
	})

	// Update
	mux.HandleFunc("POST /v1/organizations/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		store.mu.Lock()
		u, ok := store.items[id]
		if !ok {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "user not found",
			})
			return
		}

		var body struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			store.mu.Unlock()
			http.Error(w, `{"type":"invalid_request","message":"bad json"}`, http.StatusBadRequest)
			return
		}

		if body.Role != "" {
			u.Role = body.Role
		}
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u)
	})

	// Delete
	mux.HandleFunc("DELETE /v1/organizations/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		store.mu.Lock()
		_, ok := store.items[id]
		if !ok {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "user not found",
			})
			return
		}

		delete(store.items, id)
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":   id,
			"type": "user_deleted",
		})
	})

	return mux
}

func TestAccUserResource_importAndUpdateRole(t *testing.T) {
	store := newUserStore()
	store.seedUser(&mockUser{
		ID:      "user-abc123",
		Email:   "alice@example.com",
		Name:    "Alice",
		Role:    "user",
		AddedAt: time.Now().UTC().Format(time.RFC3339),
		Type:    "user",
	})

	mux := newUserMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Import the seeded user (persist state for subsequent steps)
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_user" "alice" {
  role = "user"
}
`,
				ResourceName:       "claude_user.alice",
				ImportState:        true,
				ImportStateId:      "user-abc123",
				ImportStatePersist: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_user.alice", "id", "user-abc123"),
					resource.TestCheckResourceAttr("claude_user.alice", "email", "alice@example.com"),
					resource.TestCheckResourceAttr("claude_user.alice", "name", "Alice"),
					resource.TestCheckResourceAttr("claude_user.alice", "role", "user"),
					resource.TestCheckResourceAttrSet("claude_user.alice", "added_at"),
				),
			},
			// Step 2: Update role to developer
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_user" "alice" {
  role = "developer"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_user.alice", "id", "user-abc123"),
					resource.TestCheckResourceAttr("claude_user.alice", "role", "developer"),
					resource.TestCheckResourceAttr("claude_user.alice", "email", "alice@example.com"),
				),
			},
		},
	})
}

func TestAccUserResource_import(t *testing.T) {
	store := newUserStore()
	store.seedUser(&mockUser{
		ID:      "user-import1",
		Email:   "bob@example.com",
		Name:    "Bob",
		Role:    "developer",
		AddedAt: time.Now().UTC().Format(time.RFC3339),
		Type:    "user",
	})

	mux := newUserMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Import and persist state
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_user" "bob" {
  role = "developer"
}
`,
				ResourceName:       "claude_user.bob",
				ImportState:        true,
				ImportStateId:      "user-import1",
				ImportStatePersist: true,
			},
			// Step 2: Re-import and verify state matches
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_user" "bob" {
  role = "developer"
}
`,
				ResourceName:      "claude_user.bob",
				ImportState:       true,
				ImportStateId:     "user-import1",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_createReturnsError(t *testing.T) {
	store := newUserStore()
	mux := newUserMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_user" "new" {
  role = "developer"
}
`,
				ExpectError: regexp.MustCompile("User Creation Not Supported"),
			},
		},
	})
}

func TestAccUserResource_readDeletedRemovesFromState(t *testing.T) {
	store := newUserStore()
	store.seedUser(&mockUser{
		ID:      "user-del1",
		Email:   "charlie@example.com",
		Name:    "Charlie",
		Role:    "user",
		AddedAt: time.Now().UTC().Format(time.RFC3339),
		Type:    "user",
	})

	mux := newUserMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Import the user (persist state)
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_user" "charlie" {
  role = "user"
}
`,
				ResourceName:       "claude_user.charlie",
				ImportState:        true,
				ImportStateId:      "user-del1",
				ImportStatePersist: true,
			},
			// Step 2: Delete externally, then refresh → Read removes from state
			{
				PreConfig: func() {
					store.mu.Lock()
					defer store.mu.Unlock()
					delete(store.items, "user-del1")
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
