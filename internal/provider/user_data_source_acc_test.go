package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_basic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":       "user-ds1",
			"email":    "dana@example.com",
			"name":     "Dana",
			"role":     "admin",
			"added_at": "2024-03-15T10:00:00Z",
			"type":     "user",
		})
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_user" "test" {
  id = "user-ds1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_user.test", "id", "user-ds1"),
					resource.TestCheckResourceAttr("data.claude_user.test", "email", "dana@example.com"),
					resource.TestCheckResourceAttr("data.claude_user.test", "name", "Dana"),
					resource.TestCheckResourceAttr("data.claude_user.test", "role", "admin"),
					resource.TestCheckResourceAttr("data.claude_user.test", "added_at", "2024-03-15T10:00:00Z"),
				),
			},
		},
	})
}
