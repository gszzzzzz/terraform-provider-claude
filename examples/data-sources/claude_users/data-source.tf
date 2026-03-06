data "claude_users" "all" {}

data "claude_users" "filtered" {
  email = "alice@example.com"
}

output "all_user_ids" {
  value = [for user in data.claude_users.all.users : user.id]
}

output "filtered_user_ids" {
  value = [for user in data.claude_users.filtered.users : user.id]
}
