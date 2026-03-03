# Terraform Provider for Claude Admin API

Manage Anthropic Claude resources with Terraform via the [Claude Admin API](https://docs.anthropic.com/en/docs/administration/administration-api).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Development

```bash
make build     # Build binary
make install   # Install to local Terraform plugin directory
make test      # Run unit tests
make testacc   # Run acceptance tests (TF_ACC=1)
make fmt       # Run gofmt
make lint      # Run golangci-lint
make docs      # Generate tfplugindocs documentation
make docs-lint # Validate tfplugindocs documentation
```
