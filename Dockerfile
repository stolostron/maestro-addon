FROM golang:1.21 AS builder

ARG OS=linux
ARG ARCH=amd64

WORKDIR /go/src/github.com/stolostron/maestro-addon
COPY . .
ENV GO_PACKAGE github.com/stolostron/maestro-addon

RUN GOOS=${OS} \
    GOARCH=${ARCH} \
    CGO_ENABLED=0 \
    make build --warn-undefined-variables

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
ENV USER_UID=10001

COPY --from=builder /go/src/github.com/stolostron/maestro-addon/maestroaddon /

USER ${USER_UID}
