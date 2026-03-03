# Terraform Provider for Claude Admin API

[Claude Admin API](https://docs.anthropic.com/en/docs/administration/administration-api)를 통해 Anthropic Claude 리소스를 Terraform으로 관리합니다.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Development

```bash
make build     # 바이너리 빌드
make install   # 로컬 Terraform 플러그인 디렉토리에 설치
make test      # 유닛 테스트 실행
make testacc   # 인수 테스트 실행 (TF_ACC=1)
make fmt       # gofmt 실행
make lint      # golangci-lint 실행
```
