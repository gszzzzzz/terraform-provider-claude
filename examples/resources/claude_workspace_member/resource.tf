resource "claude_workspace_member" "example" {
  workspace_id   = claude_workspace.example.id
  user_id        = claude_user.example.id
  workspace_role = "workspace_developer"
}
