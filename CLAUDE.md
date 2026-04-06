# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The maestro-addon is a Kubernetes addon that enables Maestro integration with Advanced Cluster Management (ACM). It manages ManifestWorkReplicaSet controllers and provides cloud events-driven communication between hub and managed clusters using either Kafka or gRPC brokers.

## Core Architecture

### Components
- **Hub Manager** (`pkg/hub/manager.go`): Main orchestrator that manages cluster controllers and message queue authorization
- **Cluster Controller** (`pkg/hub/controllers/cluster_controller.go`): Handles ManagedCluster lifecycle and Maestro integration
- **Message Queue Integration** (`pkg/mq/`): Abstraction layer for Kafka and gRPC brokers
- **Helpers** (`pkg/helpers/`): Utilities for Maestro API and Kafka client operations

### Key Concepts
- **Broker Types**: Supports two message queue brokers - `kafka` and `grpc` (constants in `pkg/common/constants.go`)
- **Maestro Service**: External service for cluster management (default: `http://maestro:8000`)
- **ManifestWorkReplicaSet**: Feature for deploying Kubernetes resources across managed clusters using cloud events

### Entry Point
Main binary built from `cmd/manager/main.go` creates a CLI with subcommands for hub management.

## Development Commands

### Build
```bash
make build          # Build the maestroaddon binary
make image          # Build container image
```

### Testing
```bash
make test           # Run unit tests (tests files in pkg/)
make e2e-test       # Run end-to-end tests
make verify         # Run linting and verification (golangci-lint)
```

### Deployment
```bash
# Install on ACM hub
helm install maestro-addon ./charts/maestro-addon

# Example configurations
--set global.imageOverrides.maestroImage=<custom-image>
--set maestro.broker=kafka
--set messageQueue.amqStreams.name=<kafka-cluster-name>
```

## File Structure

- `pkg/hub/`: Hub-side components and controllers
- `pkg/mq/`: Message queue broker abstractions
- `pkg/helpers/`: Shared utilities and mocks
- `charts/maestro-addon/`: Helm chart for deployment
- `contrib/examples/`: Example CRs and configurations
- `test/e2e/`: End-to-end test suite
- `test/performance/`: Performance testing tools

## Key Dependencies

- Open Cluster Management APIs (`open-cluster-management.io/api`)
- Maestro service integration (`github.com/openshift-online/maestro`)
- Kafka client (`github.com/confluentinc/confluent-kafka-go/v2`)
- Kubernetes client libraries

## Configuration

The addon supports configuration via:
- Command-line flags (maestro-service-address, message-queue-broker-type, etc.)
- Helm values (`charts/maestro-addon/values.yaml`)
- Secret-based configs for Kafka and database connections