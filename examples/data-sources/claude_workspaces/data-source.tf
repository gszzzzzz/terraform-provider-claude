data "claude_workspaces" "all" {}

data "claude_workspaces" "including_archived" {
  include_archived = true
}

output "workspace_ids" {
  value = [for workspace in data.claude_workspaces.all.workspaces : workspace.id]
}

output "including_archived_workspace_ids" {
  value = [for workspace in data.claude_workspaces.including_archived.workspaces : workspace.id]
}
