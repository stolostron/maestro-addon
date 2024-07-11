#!/usr/bin/env bash
REPO_DIR="$(cd "$(dirname ${BASH_SOURCE[0]})/../../.." ; pwd -P)"

function wait_command() {
  local command="$1"; shift
  local wait_seconds="${1:-300}"; shift # 300 seconds as default timeout
  until [[ $((wait_seconds--)) -eq 0 ]] || eval "$command &> /dev/null" ; do sleep 1; done
  ((++wait_seconds))
}

# TODO using stolostron image instead of this image once the stolostron image is available
maestro_image=${MAESTRO_IMAGE_NAME:-quay.io/redhat-user-workloads/crt-redhat-acm-tenant/maestro-main/maestro-main:351a2648395ec8bd080c6c16ca35ec3ab2514c2f}
maestro_addon_image=${IMAGE_NAME:-quay.io/stolostron/maestro-addon:latest}

echo "=== Maestro Image: $maestro_image"
echo "=== Maestro AddOn Image: $maestro_addon_image"

kind_cluster="maestro-addon-e2e-test"

cluster="loopback"

ocm_version="latest"

output="${REPO_DIR}/_output"
tools_dir="${output}/tools"

kube_config="${output}/kube.config"

kind="${tools_dir}/kind"
kubectl="${tools_dir}/kubectl"
helm="${tools_dir}/helm"
clusteradm="${tools_dir}/clusteradm"

mkdir -p ${tools_dir}

source ${REPO_DIR}/test/e2e/hack/deps.sh

echo "=== Prepare KinD test environment"
${kind} delete clusters ${kind_cluster}
${kind} create cluster --name ${kind_cluster} --kubeconfig ${kube_config}
${kind} load docker-image --name=${kind_cluster} ${maestro_addon_image}

export KUBECONFIG=${kube_config}

echo "=== Prepare Kafaka cluster"
${kubectl} create namespace strimzi
${kubectl} -n strimzi create -f 'https://strimzi.io/install/latest?namespace=strimzi'
wait_command "${kubectl} -n strimzi get deploy strimzi-cluster-operator"
${kubectl} -n strimzi wait deploy/strimzi-cluster-operator --for=condition=Available --timeout=300s
${kubectl} -n strimzi create -f ${REPO_DIR}/test/e2e/hack/manifests/kafka.yaml
${kubectl} -n strimzi wait kafka/kafka --for=condition=Ready --timeout=300s

echo "=== Prepare OCM hub"
${clusteradm} init --wait --bundle-version="${ocm_version}" \
  --feature-gates=CloudEventsDrivers=true,ManifestWorkReplicaSet=true \
  --output-join-command-file="${output}/join.sh"

echo "=== Deploy Maestro AddOn on the hub"
${helm} install maestro-addon charts/maestro-addon \
  --set global.imageOverrides.maestroImage=${maestro_image} \
  --set global.imageOverrides.maestroAddOnImage=${maestro_addon_image} \
  --set messageQueue.amqStreams.namespace=strimzi \
  --set messageQueue.amqStreams.listener.type=internal \
  --set messageQueue.amqStreams.listener.port=9093
${kubectl} -n maestro wait deploy/maestro --for=condition=Available --timeout=300s
${kubectl} -n maestro wait deploy/maestro-addon-manager --for=condition=Available --timeout=300s

echo "=== Join a cluster to the hub"
sh -c "$(cat ${output}/join.sh) ${cluster} --force-internal-endpoint-lookup"
${clusteradm} accept --wait --clusters ${cluster}

echo "=== Deploy Maestro Agent with ManagedClusterAddOn on the managed cluster"
${kubectl} -n ${cluster} apply -f ${REPO_DIR}/test/e2e/hack/manifests/managedclusteraddon.yaml
${kubectl} -n ${cluster} wait managedclusteraddons/maestro-addon --for=condition=Available --timeout=300s

echo "=== Enable manifestworkreplicaset controller on the hub"
${kubectl} -n open-cluster-management create -f ${REPO_DIR}/test/e2e/hack/manifests/work-driver-config.secret.yaml
${kubectl} patch clustermanager cluster-manager -p='{"spec":{"workConfiguration":{"workDriver":"grpc"}}}' --type=merge
${kubectl} -n open-cluster-management-hub wait deploy/cluster-manager-work-controller --for=jsonpath='{.status.replicas}'=2 --timeout=300s
${kubectl} -n open-cluster-management-hub wait deploy/cluster-manager-work-controller --for=jsonpath='{.status.replicas}'=1 --timeout=300s
${kubectl} -n open-cluster-management-hub wait deploy/cluster-manager-work-controller --for=condition=Available --timeout=300s

echo "=== Prepare the placement on the hub"
${kubectl} apply -f ${REPO_DIR}/test/e2e/hack/manifests/managedclustersetbinding.yaml
${kubectl} apply -f ${REPO_DIR}/test/e2e/hack/manifests/placement.yaml

${output}/e2e.test -test.v -ginkgo.v
