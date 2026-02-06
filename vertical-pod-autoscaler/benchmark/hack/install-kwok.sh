#!/bin/bash

# Copyright The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Installs KWOK (Kubernetes WithOut Kubelet) into the current cluster using Helm.
# Creates a fake KWOK node for benchmark pods to be scheduled on.
#
# Prerequisites: kubectl, helm
#
# Usage: ./install-kwok.sh

set -euo pipefail

KWOK_NAMESPACE=kube-system
KWOK_NODE_NAME=kwok-node

echo "=== Installing KWOK via Helm ==="

# Add KWOK Helm repo
echo "  Adding KWOK Helm repository..."
helm repo add kwok https://kwok.sigs.k8s.io/charts/ 2>/dev/null || true
helm repo update kwok

echo "  Installing KWOK chart..."
helm upgrade --install kwok kwok/kwok \
  --namespace "${KWOK_NAMESPACE}" \
  --set hostNetwork=true \
  --wait

echo "  Installing stage-fast chart..."
helm upgrade --install kwok-stage-fast kwok/stage-fast \
  --namespace "${KWOK_NAMESPACE}" \
  --wait

echo "  Waiting for KWOK controller to be ready..."
kubectl wait --for=condition=Available deployment/kwok-controller -n "${KWOK_NAMESPACE}" --timeout=60s

# Create fake KWOK node
if kubectl get node "${KWOK_NODE_NAME}" &>/dev/null; then
  echo "  KWOK node '${KWOK_NODE_NAME}' already exists, skipping"
else
  echo "  Creating KWOK fake node '${KWOK_NODE_NAME}'..."
  kubectl apply -f - <<EOF
apiVersion: v1
kind: Node
metadata:
  name: ${KWOK_NODE_NAME}
  annotations:
    node.alpha.kubernetes.io/ttl: "0"
    kwok.x-k8s.io/node: fake
    node.kubernetes.io/exclude-from-external-load-balancers: "true"
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: ${KWOK_NODE_NAME}
    kubernetes.io/os: linux
    kubernetes.io/role: agent
    node-role.kubernetes.io/agent: ""
    type: kwok
spec:
  taints:
    - key: kwok.x-k8s.io/node
      value: fake
      effect: NoSchedule
EOF
  echo "  Created KWOK fake node"
fi

echo "=== KWOK installation complete ==="
