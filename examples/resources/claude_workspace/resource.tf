resource "claude_workspace" "example" {
  name = "my-workspace"

  data_residency {
    workspace_geo         = "us"
    default_inference_geo = "us"
    allowed_inference_geos = ["us"]
  }
}
