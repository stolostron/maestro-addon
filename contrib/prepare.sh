#!/bin/bash

CURRENT_DIR="$(dirname "${BASH_SOURCE[0]}")"
DEMO_DIR="$(cd ${CURRENT_DIR} && pwd)"

export KUBECONFIG=${DEMO_DIR}/kubeconfig
export CTX_HUB_CLUSTER=kind-demo

# clenup env
. $DEMO_DIR/cleanup.sh

# prepare a demo cluster
kind create cluster --name demo --kubeconfig ${KUBECONFIG}

kind load docker-image quay.io/stolostron/maestro-addon:latest --name demo
kind load docker-image image-registry.testing/maestro/maestro:latest --name demo
kind load docker-image docker.io/library/eclipse-mosquitto:2.0.18 --name demo
kind load docker-image docker.io/library/postgres:14.2 --name demo

# deploy ocm hub
clusteradm init --wait --context ${CTX_HUB_CLUSTER} --bundle-version=latest --output-join-command-file ${DEMO_DIR}/join.sh

sh -c "$(cat ${DEMO_DIR}/join.sh) cluster1 --force-internal-endpoint-lookup --context ${CTX_HUB_CLUSTER}"

# enable manifestworkreplicaset controller
# TODO using clustermanager to enable manifestworkreplicaset controller
kubectl apply -f ${DEMO_DIR}/deploy/mwrs

# apply placement
kubectl apply -f ${DEMO_DIR}/manifest/placement
