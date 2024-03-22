#!/bin/bash

CURRENT_DIR="$(dirname "${BASH_SOURCE[0]}")"
DEMO_DIR="$(cd ${CURRENT_DIR} && pwd)"

export KUBECONFIG=${DEMO_DIR}/kubeconfig

source ${DEMO_DIR}/demo_magic

echo "kubectl -n open-cluster-management-hub get pods"
kubectl -n open-cluster-management-hub get pods

pushd ${DEMO_DIR}
comment "Deploy maestro addon manager on the hub"
pe "helm install maestro-addon ../charts/maestro-addon"
pe "kubectl -n maestro get pods -w"
pe "kubectl get clustermanagementaddons.addon.open-cluster-management.io"
pe "kubectl get addontemplates.addon.open-cluster-management.io"

comment "Import a cluster into the hub"
pe "./import.sh cluster1"
pe "kubectl get managedclusters -w"

comment "Install maestro addon in the cluster1"
pe "kubectl -n cluster1 apply -f manifest/managedclusteraddon.yaml"
pe "kubectl -n cluster1 get managedclusteraddons.addon.open-cluster-management.io"
pe "kubectl -n open-cluster-management-agent get pods -w"

comment "Apply a manifestworkreplicaset"
pe "kubectl apply -f manifest/manifestworkreplicaset.yaml"
pe "kubectl -n cluster1 get manifestworks"
pe "kubectl -n mwrs-test get deployments.apps"
pe "kubectl get manifestworkreplicasets.work.open-cluster-management.io -w"
popd
