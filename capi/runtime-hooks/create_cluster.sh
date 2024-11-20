#!/bin/bash

# 現在のkubectlコンテキストを取得
current_context=$(kubectl config current-context)

# 必要なコンテキスト名
required_context="kind-capi-test"

# コンテキストの確認
if [[ "$current_context" != "$required_context" ]]; then
  echo "Error: Current context is '$current_context'. Please switch to '$required_context'." >&2
  exit 1
fi


export CLUSTER_NAME="my-cluster"
export NAMESPACE="default"
export SERVICE_CIDR='["10.128.0.0/12"]'
export POD_CIDR='["192.168.0.0/16"]'
export SERVICE_DOMAIN="cluster.local"
export CONTROL_PLANE_MACHINE_COUNT=1
export WORKER_MACHINE_COUNT=1
export KUBERNETES_VERSION="v1.31.0"
export POD_SECURITY_STANDARD_ENABLED=true

# envsubstだと展開できないので、clusterctl generate yamlを使う必要があります
clusterctl generate yaml --from https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/heads/main/test/infrastructure/docker/templates/cluster-template-development.yaml | kubectl apply -f -
