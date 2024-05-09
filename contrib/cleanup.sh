#!/usr/bin/env bash

CURRENT_DIR="$(dirname "${BASH_SOURCE[0]}")"
WORK_DIR="$(cd ${CURRENT_DIR} && pwd)"

oc delete -f ${WORK_DIR}/examples/manifestworkreplicaset.yaml --ignore-not-found

oc patch clustermanager cluster-manager --type=merge -p='{"spec":{"workConfiguration":{"featureGates":[{"feature":"CloudEventsDrivers","mode":"Disable"},{"feature":"ManifestWorkReplicaSet","mode":"Disable"}]}}}'

oc -n multicluster-engine delete -f ${WORK_DIR}/examples/work-driver-config.yaml --ignore-not-found

helm uninstall maestro-addon

oc delete -f ${WORK_DIR}/examples/placement --ignore-not-found
