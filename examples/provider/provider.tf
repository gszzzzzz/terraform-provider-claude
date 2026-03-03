provider "claude" {
  # Can also be set via ANTHROPIC_ADMIN_API_KEY environment variable.
  api_key = var.anthropic_admin_api_key

  # base_url = "https://api.anthropic.com"  # default
}
