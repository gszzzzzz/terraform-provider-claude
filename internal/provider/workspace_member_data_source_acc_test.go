package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspaceMemberDataSource_basic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/workspaces/{workspace_id}/members/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"type":           "workspace_member",
			"user_id":        r.PathValue("user_id"),
			"workspace_id":   r.PathValue("workspace_id"),
			"workspace_role": "workspace_developer",
		})
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_workspace_member" "test" {
  workspace_id = "ws-ds1"
  user_id      = "user-ds1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_workspace_member.test", "workspace_id", "ws-ds1"),
					resource.TestCheckResourceAttr("data.claude_workspace_member.test", "user_id", "user-ds1"),
					resource.TestCheckResourceAttr("data.claude_workspace_member.test", "workspace_role", "workspace_developer"),
				),
			},
		},
	})
}
