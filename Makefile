SHELL :=/bin/bash

container_tool?=podman

IMAGE_REGISTRY?=quay.io/stolostron
IMAGE_TAG?=latest
IMAGE_NAME?=$(IMAGE_REGISTRY)/maestro-addon:$(IMAGE_TAG)

GOLANGCI_LINT_VERSION=v1.54.1

verify:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go vet ./...
	golangci-lint run --timeout=3m ./...
.PHONY: verify

build:
	go build -trimpath -ldflags="-s -w" -o maestroaddon cmd/manager/main.go
.PHONY: build

image:
	$(container_tool) build -f Dockerfile -t $(IMAGE_NAME) .
.PHONY: image

test:
	go test ./pkg/...
.PHONY: test
