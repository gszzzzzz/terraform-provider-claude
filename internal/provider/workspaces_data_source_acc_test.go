package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspacesDataSource_basic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":            "ws-1",
					"name":          "Workspace 1",
					"display_color": "blue",
					"created_at":    "2024-01-01T00:00:00Z",
					"archived_at":   nil,
				},
				{
					"id":            "ws-2",
					"name":          "Workspace 2",
					"display_color": "red",
					"created_at":    "2024-02-01T00:00:00Z",
					"archived_at":   nil,
				},
			},
			"first_id": "ws-1",
			"last_id":  "ws-2",
			"has_more": false,
		})
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_workspaces" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.#", "2"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.id", "ws-1"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.name", "Workspace 1"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.display_color", "blue"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.created_at", "2024-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.id", "ws-2"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.name", "Workspace 2"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.display_color", "red"),
				),
			},
		},
	})
}

func TestAccWorkspacesDataSource_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":     []map[string]any{},
			"first_id": "",
			"last_id":  "",
			"has_more": false,
		})
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_workspaces" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.#", "0"),
				),
			},
		},
	})
}

func TestAccWorkspacesDataSource_pagination(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		afterID := r.URL.Query().Get("after_id")

		if callCount == 0 && afterID == "" {
			callCount++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":            "ws-page1",
						"name":          "Page1",
						"display_color": "blue",
						"created_at":    "2024-01-01T00:00:00Z",
						"archived_at":   nil,
					},
				},
				"first_id": "ws-page1",
				"last_id":  "ws-page1",
				"has_more": true,
			})
		} else if afterID == "ws-page1" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":            "ws-page2",
						"name":          "Page2",
						"display_color": "red",
						"created_at":    "2024-02-01T00:00:00Z",
						"archived_at":   nil,
					},
				},
				"first_id": "ws-page2",
				"last_id":  "ws-page2",
				"has_more": false,
			})
		}
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_workspaces" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.#", "2"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.id", "ws-page1"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.name", "Page1"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.id", "ws-page2"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.name", "Page2"),
				),
			},
		},
	})
}

func TestAccWorkspacesDataSource_withDataResidency(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Use raw JSON to test the union type (allowed_inference_geos as array)
		_, _ = w.Write([]byte(`{
			"data": [
				{
					"id": "ws-geo",
					"name": "Geo Workspace",
					"display_color": "green",
					"created_at": "2024-01-01T00:00:00Z",
					"archived_at": null,
					"data_residency": {
						"workspace_geo": "us",
						"default_inference_geo": "us",
						"allowed_inference_geos": ["us", "eu"]
					}
				},
				{
					"id": "ws-unrestricted",
					"name": "Unrestricted Workspace",
					"display_color": "purple",
					"created_at": "2024-02-01T00:00:00Z",
					"archived_at": null,
					"data_residency": {
						"workspace_geo": "eu",
						"default_inference_geo": "eu",
						"allowed_inference_geos": "unrestricted"
					}
				}
			],
			"first_id": "ws-geo",
			"last_id": "ws-unrestricted",
			"has_more": false
		}`))
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_workspaces" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.#", "2"),
					// Workspace with array allowed_inference_geos
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.id", "ws-geo"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.data_residency.workspace_geo", "us"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.data_residency.default_inference_geo", "us"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.data_residency.allowed_inference_geos.#", "2"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.data_residency.allowed_inference_geos.0", "us"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.0.data_residency.allowed_inference_geos.1", "eu"),
					// Workspace with string "unrestricted"
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.id", "ws-unrestricted"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.data_residency.workspace_geo", "eu"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.data_residency.default_inference_geo", "eu"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.data_residency.allowed_inference_geos.#", "1"),
					resource.TestCheckResourceAttr("data.claude_workspaces.test", "workspaces.1.data_residency.allowed_inference_geos.0", "unrestricted"),
				),
			},
		},
	})
}
