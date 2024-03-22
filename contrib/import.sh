#!/bin/bash

sh -c "$(cat join.sh) cluster1 --force-internal-endpoint-lookup --context kind-demo"

clusteradm accept --clusters cluster1 --context kind-demo
