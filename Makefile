SHELL :=/bin/bash

container_tool?=podman

IMAGE_REGISTRY?=quay.io/stolostron
IMAGE_TAG?=latest
IMAGE_NAME?=$(IMAGE_REGISTRY)/maestro-addon:$(IMAGE_TAG)

build:
	go build -trimpath -ldflags="-s -w" -o maestroaddon cmd/manager/main.go
.PHONY: build

image:
	$(container_tool) build -f Dockerfile -t $(IMAGE_NAME) .
.PHONY: image
