package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// mockWorkspace is the in-memory representation used by the mock handler.
type mockWorkspace struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	DisplayColor  string             `json:"display_color"`
	CreatedAt     string             `json:"created_at"`
	ArchivedAt    *string            `json:"archived_at"`
	DataResidency *mockDataResidency `json:"data_residency,omitempty"`
}

type mockDataResidency struct {
	WorkspaceGeo         string          `json:"workspace_geo"`
	DefaultInferenceGeo  string          `json:"default_inference_geo"`
	AllowedInferenceGeos json.RawMessage `json:"allowed_inference_geos"`
}

// workspaceStore is a concurrency-safe in-memory store for mock workspaces.
type workspaceStore struct {
	mu     sync.Mutex
	items  map[string]*mockWorkspace
	nextID int
}

func newWorkspaceStore() *workspaceStore {
	return &workspaceStore{
		items:  make(map[string]*mockWorkspace),
		nextID: 1,
	}
}

// newWorkspaceMockMux creates an http.ServeMux that handles workspace CRUD
// backed by the given store.
func newWorkspaceMockMux(t *testing.T, store *workspaceStore) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()

	// Create
	mux.HandleFunc("POST /v1/organizations/workspaces", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name          string `json:"name"`
			DataResidency *struct {
				WorkspaceGeo         string          `json:"workspace_geo"`
				DefaultInferenceGeo  string          `json:"default_inference_geo"`
				AllowedInferenceGeos json.RawMessage `json:"allowed_inference_geos"`
			} `json:"data_residency"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"type":"invalid_request","message":"bad json"}`, http.StatusBadRequest)
			return
		}

		store.mu.Lock()
		id := fmt.Sprintf("wrkspc_%d", store.nextID)
		store.nextID++

		ws := &mockWorkspace{
			ID:           id,
			Name:         body.Name,
			DisplayColor: "blue",
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		if body.DataResidency != nil {
			ws.DataResidency = &mockDataResidency{
				WorkspaceGeo:         body.DataResidency.WorkspaceGeo,
				DefaultInferenceGeo:  body.DataResidency.DefaultInferenceGeo,
				AllowedInferenceGeos: body.DataResidency.AllowedInferenceGeos,
			}
		}

		store.items[id] = ws
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ws)
	})

	// Read
	mux.HandleFunc("GET /v1/organizations/workspaces/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		store.mu.Lock()
		ws, ok := store.items[id]
		store.mu.Unlock()

		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "workspace not found",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	})

	// Update
	mux.HandleFunc("POST /v1/organizations/workspaces/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		store.mu.Lock()
		ws, ok := store.items[id]
		if !ok {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "workspace not found",
			})
			return
		}

		var body struct {
			Name                 string          `json:"name"`
			DefaultInferenceGeo  string          `json:"default_inference_geo"`
			AllowedInferenceGeos json.RawMessage `json:"allowed_inference_geos"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			store.mu.Unlock()
			http.Error(w, `{"type":"invalid_request","message":"bad json"}`, http.StatusBadRequest)
			return
		}

		if body.Name != "" {
			ws.Name = body.Name
		}
		if body.DefaultInferenceGeo != "" {
			if ws.DataResidency == nil {
				ws.DataResidency = &mockDataResidency{}
			}
			ws.DataResidency.DefaultInferenceGeo = body.DefaultInferenceGeo
		}
		if body.AllowedInferenceGeos != nil {
			if ws.DataResidency == nil {
				ws.DataResidency = &mockDataResidency{}
			}
			ws.DataResidency.AllowedInferenceGeos = body.AllowedInferenceGeos
		}
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	})

	// Archive
	mux.HandleFunc("POST /v1/organizations/workspaces/{id}/archive", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		store.mu.Lock()
		ws, ok := store.items[id]
		if !ok {
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"type":    "not_found",
				"message": "workspace not found",
			})
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		ws.ArchivedAt = &now
		store.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	})

	return mux
}

func TestAccWorkspaceResource_basic(t *testing.T) {
	store := newWorkspaceStore()
	mux := newWorkspaceMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "test-workspace"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["us"]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("claude_workspace.test", "id"),
					resource.TestCheckResourceAttr("claude_workspace.test", "name", "test-workspace"),
					resource.TestCheckResourceAttr("claude_workspace.test", "display_color", "blue"),
					resource.TestCheckResourceAttrSet("claude_workspace.test", "created_at"),
				),
			},
			// Step 2: Update name (in-place)
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "test-workspace-updated"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["us"]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_workspace.test", "name", "test-workspace-updated"),
					resource.TestCheckResourceAttr("claude_workspace.test", "display_color", "blue"),
				),
			},
		},
	})
}

func TestAccWorkspaceResource_dataResidency(t *testing.T) {
	store := newWorkspaceStore()
	mux := newWorkspaceMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with data residency
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "geo-workspace"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["us"]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("claude_workspace.test", "id"),
					resource.TestCheckResourceAttr("claude_workspace.test", "name", "geo-workspace"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.workspace_geo", "us"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.default_inference_geo", "us"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.#", "1"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.0", "us"),
				),
			},
			// Step 2: Update inference geos (workspace_geo stays the same)
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "geo-workspace"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "eu"
    allowed_inference_geos = ["us", "eu"]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.workspace_geo", "us"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.default_inference_geo", "eu"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.#", "2"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.0", "us"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.1", "eu"),
				),
			},
		},
	})
}

func TestAccWorkspaceResource_unrestrictedGeos(t *testing.T) {
	store := newWorkspaceStore()
	mux := newWorkspaceMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "unrestricted-workspace"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["unrestricted"]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("claude_workspace.test", "name", "unrestricted-workspace"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.#", "1"),
					resource.TestCheckResourceAttr("claude_workspace.test", "data_residency.allowed_inference_geos.0", "unrestricted"),
				),
			},
		},
	})
}

func TestAccWorkspaceResource_import(t *testing.T) {
	store := newWorkspaceStore()
	mux := newWorkspaceMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "import-workspace"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["us"]
  }
}
`,
				Check: resource.TestCheckResourceAttrSet("claude_workspace.test", "id"),
			},
			// Step 2: Import
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "import-workspace"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["us"]
  }
}
`,
				ResourceName:      "claude_workspace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkspaceResource_readArchivedRemovesFromState(t *testing.T) {
	store := newWorkspaceStore()
	mux := newWorkspaceMockMux(t, store)
	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create workspace
			{
				Config: testAccProviderConfig(server.URL) + `
resource "claude_workspace" "test" {
  name = "will-be-archived"
  data_residency = {
    workspace_geo          = "us"
    default_inference_geo  = "us"
    allowed_inference_geos = ["us"]
  }
}
`,
				Check: resource.TestCheckResourceAttrSet("claude_workspace.test", "id"),
			},
			// Step 2: Archive externally, then refresh → Read removes from state
			{
				PreConfig: func() {
					store.mu.Lock()
					defer store.mu.Unlock()
					for _, ws := range store.items {
						now := time.Now().UTC().Format(time.RFC3339)
						ws.ArchivedAt = &now
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
