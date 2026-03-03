package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationDataSource_basic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":   "org-123",
			"name": "Test Organization",
		})
	})

	server := testAccMockServer(t, mux)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
data "claude_organization" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_organization.test", "id", "org-123"),
					resource.TestCheckResourceAttr("data.claude_organization.test", "name", "Test Organization"),
				),
			},
		},
	})
}
