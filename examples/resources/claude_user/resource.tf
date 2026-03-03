# Users cannot be created via the API.
# Import an existing user, then manage their role.
resource "claude_user" "example" {
  role = "organization_admin"
}
