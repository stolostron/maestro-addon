#!/usr/bin/env bash

go_path=$(go env GOPATH)

host_os=$(go env GOHOSTOS)
host_arch=$(go env GOHOSTARCH)

kind_version="v0.22.0"
kubectl_version="v1.23.6"
helm_version="v3.14.4"
clusteradm_version="latest"

pushd ${tools_dir}

if ! command -v ${tools_dir}/kind >/dev/null 2>&1; then
    echo "=== Installing kind ${kind_version} to ${tools_dir}"
    curl -SsL "https://kind.sigs.k8s.io/dl/${kind_version}/kind-${host_os}-${host_arch}" -o kind
    chmod +x kind
fi

if ! command -v ${tools_dir}/kubectl >/dev/null 2>&1; then
    echo "=== Installing kubectl ${kubectl_version} to ${tools_dir}"
    curl -SsL "https://storage.googleapis.com/kubernetes-release/release/${kubectl_version}/bin/${host_os}/${host_arch}/kubectl" -o kubectl
    chmod +x kubectl
fi

if ! command -v ${tools_dir}/helm >/dev/null 2>&1; then
    echo "=== Installing helm ${helm_version} to ${tools_dir}"
    curl -SsL "https://get.helm.sh/helm-${helm_version}-${host_os}-${host_arch}.tar.gz" -o helm.tar.gz
    tar -xf helm.tar.gz
    cp ${host_os}-${host_arch}/helm ./
fi

if ! command -v ${tools_dir}/clusteradm >/dev/null 2>&1; then
    echo "=== Installing clusteradm ${clusteradm_version} to ${tools_dir}"
    base="https://github.com/open-cluster-management-io/clusteradm"
    if [ "${clusteradm_version}" == "latest" ]; then
        url="${base}/latest/download/${tar}"
        git clone --depth=1 --branch=main ${base}.git
        pushd clusteradm
        make build
        popd
        rm -rf clusteradm
        cp ${go_path}/bin/clusteradm ./clusteradm
    else
        tar="clusteradm_${host_os}_${host_arch}.tar.gz"
        url="${base}/releases/download/${clusteradm_version}/${tar}"
        curl -SsL "${url}/" -o ${tar}
        tar -xf ${tar}
    fi
fi

popd
