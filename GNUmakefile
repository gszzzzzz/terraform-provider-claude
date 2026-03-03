default: build

BINARY=terraform-provider-claude
HOSTNAME=registry.terraform.io
NAMESPACE=gszzzzzz
NAME=claude
VERSION=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

build:
	go build -o $(BINARY)

install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/

test:
	go test ./... -v $(TESTARGS) -timeout 120s

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -s -w .

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate

docs-lint:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest validate

.PHONY: build install test testacc fmt lint docs docs-lint
