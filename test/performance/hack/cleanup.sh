#!/usr/bin/env bash

REPO_DIR="$(cd "$(dirname ${BASH_SOURCE[0]})/../../.." ; pwd -P)"

kubectl delete namespaces -l maestro.performance.test=acm --ignore-not-found

kubectl -n amq-streams delete kafka --all --ignore-not-found
kubectl delete ns amq-streams

helm uninstall maestro

rm -rf ${REPO_DIR}/_output/performance/acm
