#!/bin/bash

set -e

if [ ! -d cluster-api ]; then
  git clone https://github.com/kubernetes-sigs/cluster-api
fi

export KIND_NETWORK_IPFAMILY="ipv4"
export KIND_CLUSTER_NAME=${CAPI_KIND_CLUSTER_NAME:-"capi-test"}

if ! kind get clusters | grep -q "$KIND_CLUSTER_NAME"; then
  cat <<EOF | kind create cluster --name="$KIND_CLUSTER_NAME"  --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: ${KIND_NETWORK_IPFAMILY}
  disableDefaultCNI: true
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: /var/run/docker.sock
        containerPath: /var/run/docker.sock
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry]
      config_path = "/etc/containerd/certs.d"
EOF
fi

kubectl config use kind-$KIND_CLUSTER_NAME

if ! helm repo list | grep -q cilium; then
  helm repo add cilium https://helm.cilium.io/
  helm repo update
fi

docker pull quay.io/cilium/cilium:v1.16.4
kind load docker-image quay.io/cilium/cilium:v1.16.4

helm upgrade --install cilium cilium/cilium --version 1.16.4 \
  --namespace kube-system \
  --set image.pullPolicy=IfNotPresent \
  --set operator.replicas=1 \
  --set ipam.mode=kubernetes \
  --set hubble.relay.enabled=true \
  --set hubble.ui.enabled=true \
  --wait

cat <<EOF | kubectl apply -f -
apiVersion: cilium.io/v2
kind: CiliumClusterwideNetworkPolicy
metadata:
  name: show-fqdn
spec:
  endpointSelector:
    matchExpressions:
    - key: "k8s:io.kubernetes.pod.namespace"
      operator: NotIn
      values:
      - kube-system
  egress:
  - toFQDNs:
    - matchPattern: "*"
  - toCIDR:
    - 0.0.0.0/0
  - toEntities:
    - all
  - toEndpoints:
    - matchLabels:
        "k8s:io.kubernetes.pod.namespace": kube-system
        "k8s:k8s-app": kube-dns
    toPorts:
    - ports:
      - port: "53"
        protocol: ANY
      rules:
        dns:
        - matchPattern: "*"
EOF

cd cluster-api
make kind-cluster
cd ..

export CLUSTER_TOPOLOGY="true"
export EXP_KUBEADM_BOOTSTRAP_FORMAT_IGNITION="true"
export EXP_RUNTIME_SDK="true"
export EXP_MACHINE_SET_PREFLIGHT_CHECKS="true"
clusterctl init --infrastructure docker --addon helm
tilt up

kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/heads/main/test/infrastructure/docker/templates/clusterclass-quick-start.yaml
