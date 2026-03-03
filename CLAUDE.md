# Terraform Provider for Claude Admin API

## Build & Test Commands

```bash
make build        # 바이너리 빌드
make install      # 로컬 Terraform 플러그인 디렉토리에 설치
make test         # 유닛 테스트 실행
make testacc      # 인수 테스트 실행 (TF_ACC=1)
make fmt          # gofmt 실행
make lint         # golangci-lint 실행
```

## Architecture

- **Plugin Framework**: `terraform-plugin-framework` (not SDKv2)
- **Go Module**: `github.com/gszzzzzz/terraform-provider-claude`
- **Provider Name**: `claude` → resources are `claude_workspace`, `claude_organization`, etc.

### Directory Structure

- `main.go` — entrypoint (providerserver)
- `internal/client/` — HTTP client for Claude Admin API (separate from provider logic)
- `internal/provider/` — Terraform provider, resources, data sources

### Conventions

- API client methods live in `internal/client/`, one file per resource type
- Provider resources/data sources live in `internal/provider/`, one file per resource
- Admin API uses POST for updates (not PUT/PATCH)
- Workspace "delete" is archive (soft delete)
- `x-api-key` header for auth, `anthropic-version: 2023-06-01` required header
- Environment variables: `ANTHROPIC_ADMIN_API_KEY`, `ANTHROPIC_BASE_URL`
