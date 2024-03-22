#!/bin/bash

CURRENT_DIR="$(dirname "${BASH_SOURCE[0]}")"
DEMO_DIR="$(cd ${CURRENT_DIR} && pwd)"

export KUBECONFIG=${DEMO_DIR}/kubeconfig
export CTX_HUB_CLUSTER=kind-demo

# clenup env
. $DEMO_DIR/cleanup.sh

# prepare certs for mqtt
# pushd ${DEMO_DIR}
# cfssl gencert -initca config/ca-csr.json | cfssljson -bare ca -
# echo '{"CN":"meastro-mqtt-server","hosts":["maestro-mqtt","maestro-mqtt.maestro","maestro-mqtt.maestro.svc","localhost"],"key":{"algo":"rsa","size":2048}}' | \
#   cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=config/ca-config.json -profile=server - | \
#   cfssljson -bare mqtt-server
# echo '{"CN":"meastro-mqtt-client","hosts":["maestro-addon.addon.open-cluster-management.io"],"key":{"algo":"rsa","size":2048}}' | \
#   cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=config/ca-config.json -profile=client - | \
#   cfssljson -bare mqtt-client
# popd

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

# install maestro addon with customized certs
# helm install maestro-addon charts/maestro-addon \
#   --set global.messageQueue.useCustomizedCerts=true \
#   --set-file global.messageQueue.certs.ca=${DEMO_DIR}/ca.pem \
#   --set-file global.messageQueue.certs.caKey=${DEMO_DIR}/ca-key.pem \
#   --set-file global.messageQueue.certs.serverCert=${DEMO_DIR}/mqtt-server.pem \
#   --set-file global.messageQueue.certs.serverKey=${DEMO_DIR}/mqtt-server-key.pem \
#   --set-file global.messageQueue.certs.clientCert=${DEMO_DIR}/mqtt-client.pem \
#   --set-file global.messageQueue.certs.clientKey=${DEMO_DIR}/mqtt-client-key.pem
