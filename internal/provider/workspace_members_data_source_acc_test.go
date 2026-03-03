package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspaceMembersDataSource_basic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces/{workspace_id}/members", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{
					"type":           "workspace_member",
					"user_id":        "user-1",
					"workspace_id":   r.PathValue("workspace_id"),
					"workspace_role": "workspace_admin",
				},
				{
					"type":           "workspace_member",
					"user_id":        "user-2",
					"workspace_id":   r.PathValue("workspace_id"),
					"workspace_role": "workspace_developer",
				},
			},
			"first_id": "user-1",
			"last_id":  "user-2",
			"has_more": false,
		})
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_workspace_members" "test" {
  workspace_id = "ws-list1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "workspace_id", "ws-list1"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.#", "2"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.0.user_id", "user-1"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.0.workspace_id", "ws-list1"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.0.workspace_role", "workspace_admin"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.1.user_id", "user-2"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.1.workspace_id", "ws-list1"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.1.workspace_role", "workspace_developer"),
				),
			},
		},
	})
}

func TestAccWorkspaceMembersDataSource_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces/{workspace_id}/members", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":     []map[string]string{},
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
data "claude_workspace_members" "test" {
  workspace_id = "ws-empty"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "workspace_id", "ws-empty"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.#", "0"),
				),
			},
		},
	})
}

func TestAccWorkspaceMembersDataSource_pagination(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces/{workspace_id}/members", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		afterID := r.URL.Query().Get("after_id")

		if callCount == 0 && afterID == "" {
			callCount++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{
					{
						"type":           "workspace_member",
						"user_id":        "user-page1",
						"workspace_id":   r.PathValue("workspace_id"),
						"workspace_role": "workspace_admin",
					},
				},
				"first_id": "user-page1",
				"last_id":  "user-page1",
				"has_more": true,
			})
		} else if afterID == "user-page1" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{
					{
						"type":           "workspace_member",
						"user_id":        "user-page2",
						"workspace_id":   r.PathValue("workspace_id"),
						"workspace_role": "workspace_developer",
					},
				},
				"first_id": "user-page2",
				"last_id":  "user-page2",
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
data "claude_workspace_members" "test" {
  workspace_id = "ws-paginated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.#", "2"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.0.user_id", "user-page1"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.0.workspace_role", "workspace_admin"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.1.user_id", "user-page2"),
					resource.TestCheckResourceAttr("data.claude_workspace_members.test", "members.1.workspace_role", "workspace_developer"),
				),
			},
		},
	})
}
