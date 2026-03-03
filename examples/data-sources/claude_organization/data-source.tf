data "claude_organization" "current" {}

output "organization_id" {
  value = data.claude_organization.current.id
}

output "organization_name" {
  value = data.claude_organization.current.name
}
