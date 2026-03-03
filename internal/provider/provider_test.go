package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"claude": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccMockServer creates an httptest.Server with the given handler and
// registers cleanup via t.Cleanup.
func testAccMockServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// testAccProviderConfig returns a provider configuration block pointing at the
// given mock server base URL.
func testAccProviderConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "claude" {
  api_key  = "test-api-key"
  base_url = %q
}
`, baseURL)
}
