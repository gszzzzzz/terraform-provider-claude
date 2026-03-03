# Terraform Provider for Claude Admin API

[Claude Admin API](https://docs.anthropic.com/en/docs/administration/administration-api)를 통해 Anthropic Claude 리소스를 Terraform으로 관리합니다.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Resources & Data Sources

### Resources

| Resource | Description |
|---|---|
| `claude_workspace` | Workspace 생성 및 관리 |
| `claude_user` | User 관리 |
| `claude_workspace_member` | Workspace Member 관리 |

### Data Sources

| Data Source | Description |
|---|---|
| `claude_organization` | Organization 정보 조회 |
| `claude_user` | User 정보 조회 |
| `claude_workspace_member` | Workspace Member 정보 조회 |
| `claude_workspace_members` | Workspace Member 목록 조회 |

## Usage

```hcl
terraform {
  required_providers {
    claude = {
      source = "gszzzzzz/claude"
    }
  }
}

provider "claude" {
  # api_key  = "..."  # 또는 ANTHROPIC_ADMIN_API_KEY 환경변수 사용
  # base_url = "..."  # 또는 ANTHROPIC_BASE_URL 환경변수 사용 (기본값: https://api.anthropic.com)
}

resource "claude_workspace" "example" {
  name        = "my-workspace"
  description = "Example workspace"
}
```

## Authentication

Admin API 키를 다음 중 하나의 방법으로 설정합니다:

- Provider 설정에서 `api_key` 속성 지정
- `ANTHROPIC_ADMIN_API_KEY` 환경변수 설정

```bash
export ANTHROPIC_ADMIN_API_KEY="sk-ant-admin01-..."
```

## Development

```bash
make build     # 바이너리 빌드
make install   # 로컬 Terraform 플러그인 디렉토리에 설치
make test      # 유닛 테스트 실행
make testacc   # 인수 테스트 실행 (TF_ACC=1)
make fmt       # gofmt 실행
make lint      # golangci-lint 실행
```
