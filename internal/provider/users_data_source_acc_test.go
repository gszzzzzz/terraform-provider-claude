package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUsersDataSource_basic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{
					"id":       "user-1",
					"email":    "alice@example.com",
					"name":     "Alice",
					"role":     "admin",
					"added_at": "2024-01-01T00:00:00Z",
					"type":     "user",
				},
				{
					"id":       "user-2",
					"email":    "bob@example.com",
					"name":     "Bob",
					"role":     "developer",
					"added_at": "2024-02-01T00:00:00Z",
					"type":     "user",
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
data "claude_users" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_users.test", "users.#", "2"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.id", "user-1"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.email", "alice@example.com"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.name", "Alice"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.role", "admin"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.added_at", "2024-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.id", "user-2"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.email", "bob@example.com"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.name", "Bob"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.role", "developer"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.added_at", "2024-02-01T00:00:00Z"),
				),
			},
		},
	})
}

func TestAccUsersDataSource_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/users", func(w http.ResponseWriter, r *http.Request) {
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
data "claude_users" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_users.test", "users.#", "0"),
				),
			},
		},
	})
}

func TestAccUsersDataSource_pagination(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		afterID := r.URL.Query().Get("after_id")

		if callCount == 0 && afterID == "" {
			callCount++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{
					{
						"id":       "user-page1",
						"email":    "page1@example.com",
						"name":     "Page1",
						"role":     "admin",
						"added_at": "2024-01-01T00:00:00Z",
						"type":     "user",
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
						"id":       "user-page2",
						"email":    "page2@example.com",
						"name":     "Page2",
						"role":     "developer",
						"added_at": "2024-02-01T00:00:00Z",
						"type":     "user",
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
data "claude_users" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_users.test", "users.#", "2"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.id", "user-page1"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.role", "admin"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.id", "user-page2"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.1.role", "developer"),
				),
			},
		},
	})
}

func TestAccUsersDataSource_emailFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/organizations/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		email := r.URL.Query().Get("email")
		if email == "alice@example.com" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{
					{
						"id":       "user-1",
						"email":    "alice@example.com",
						"name":     "Alice",
						"role":     "admin",
						"added_at": "2024-01-01T00:00:00Z",
						"type":     "user",
					},
				},
				"first_id": "user-1",
				"last_id":  "user-1",
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
data "claude_users" "test" {
  email = "alice@example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.claude_users.test", "email", "alice@example.com"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.#", "1"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.id", "user-1"),
					resource.TestCheckResourceAttr("data.claude_users.test", "users.0.email", "alice@example.com"),
				),
			},
		},
	})
}
