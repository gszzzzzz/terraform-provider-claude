terraform {
  required_providers {
    claude = {
      source = "registry.terraform.io/gszzzzzz/claude"
    }
  }
}

provider "claude" {
  # api_key  = "sk-ant-admin-..."  # Or set ANTHROPIC_ADMIN_API_KEY env var
  # base_url = "https://api.anthropic.com"  # Optional
}
