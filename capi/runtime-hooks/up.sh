#!/bin/bash

if [ ! -d cluster-api ]; then
  git clone https://github.com/kubernetes-sigs/cluster-api
fi
cd cluster-api
KIND_NETWORK_IPFAMILY=ipv4 make kind-cluster
cd ..

export CLUSTER_TOPOLOGY="true"
export EXP_KUBEADM_BOOTSTRAP_FORMAT_IGNITION="true"
export EXP_RUNTIME_SDK="true"
export EXP_MACHINE_SET_PREFLIGHT_CHECKS="true"
clusterctl init --infrastructure docker --addon helm
tilt up
