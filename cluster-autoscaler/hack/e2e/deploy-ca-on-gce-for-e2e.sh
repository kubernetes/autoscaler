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

set -o nounset
set -o pipefail
set -o errexit

# PREREQUISITES:
# 1. A GCE cluster created with 'kubetest2' (or legacy kube-up.sh).
# 2. KUBECONFIG specified or available at $HOME/.kube/config.
#
# USAGE: deploy-ca-on-gce-for-e2e.sh <CA_IMAGE>
#
# This script relies on specific naming conventions for nodes and MIGs common in those setups:
# - Control-plane node can be identified by its taints ("node-role.kubernetes.io/control-plane").
# - Only one MIG, the nodes follow a pattern where stripping the last segment yields the MIG name.
# - ProviderID format is gce://<project>/<zone>/<name>

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
CA_ROOT="$(readlink -f "${SCRIPT_DIR}/../..")"
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

CA_IMAGE="${1:-}"

if [[ -z "${CA_IMAGE}" ]]; then
    echo "Error: CA_IMAGE is not set. Please provide it as the first argument."
    echo "Usage: $0 <image>"
    exit 1
fi

echo "Using image: ${CA_IMAGE}"

# Deploy
echo "Deploying to cluster..."

# Detect control-plane node (grep by 'control-plane' taint)
CONTROL_PLANE_NODE="$(kubectl get nodes --no-headers -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints | grep "node-role.kubernetes.io/control-plane" | awk '{print $1; exit}')"
if [[ -z "${CONTROL_PLANE_NODE}" ]]; then
    echo "Error: Could not find a control-plane node."
    exit 1
fi
echo "Identified control-plane node: ${CONTROL_PLANE_NODE}"

# Get the API server URL (e.g., https://34.57.28.81:6443)
# We extract the host/port to inject them as env vars because hostNetwork: true pods
# in this GCE environment often cannot reach the default 10.0.0.1 Service IP due to
# custom routing/firewall rules on the master node.
KUBE_URL="$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')"
# Extract host and port
KUBERNETES_SERVICE_HOST="$(echo ${KUBE_URL} | sed -e 's|https://||' -e 's|:.*||')"
KUBERNETES_SERVICE_PORT="$(echo ${KUBE_URL} | grep -oP '(?<=:)\d+$' || echo "443")"
echo "Identified API server host: ${KUBERNETES_SERVICE_HOST}, port: ${KUBERNETES_SERVICE_PORT}"

# Extract Project and Zone from master node's providerID
PROVIDER_ID="$(kubectl get node ${CONTROL_PLANE_NODE} -o jsonpath='{.spec.providerID}')"
PROJECT="$(echo ${PROVIDER_ID} | cut -d/ -f3)"
ZONE="$(echo ${PROVIDER_ID} | cut -d/ -f4)"
echo "Identified GCE Project: ${PROJECT}, Zone: ${ZONE}"

# Discover all distinct worker MIGs from existing nodes, or accept comma-separated MIG_NAMES
if [[ -n "${MIG_NAMES:-}" ]]; then
    IFS=',' read -ra MIG_LIST <<< "${MIG_NAMES}"
else
    MIG_LIST=($(kubectl get nodes --no-headers -o custom-columns=NAME:.metadata.name | grep -v -E 'master|control-plane' | sed 's/-[^-]*$//' | sort -u))
fi

if [[ ${#MIG_LIST[@]} -eq 0 ]]; then
    WORKER_NODE="$(kubectl get nodes --no-headers -o custom-columns=NAME:.metadata.name | grep -v -E 'master|control-plane' | head -n 1)"
    MIG_NAME="$(echo ${WORKER_NODE} | sed 's/-[^-]*$//')"
    MIG_LIST=("${MIG_NAME}")
fi
echo "Identified MIGs (${#MIG_LIST[@]}): ${MIG_LIST[*]}"

# Default limits per MIG for parallel workers (e.g. 1 to 5 nodes per MIG)
MIN_NODES="${MIN_NODES:-1}"
MAX_NODES="${MAX_NODES:-5}"
# Extra cluster-autoscaler flags.
EXTRA_CA_FLAGS="${EXTRA_CA_FLAGS:-""}"

# Primary MIG specification
PRIMARY_NODES_SPEC="${MIN_NODES}:${MAX_NODES}:https://www.googleapis.com/compute/v1/projects/${PROJECT}/zones/${ZONE}/instanceGroups/${MIG_LIST[0]}"
echo "Primary nodes spec: ${PRIMARY_NODES_SPEC}"

# Additional MIG specifications for parallel workers
EXTRA_NODES_FLAGS=()
for ((i=1; i<${#MIG_LIST[@]}; i++)); do
    mig="${MIG_LIST[$i]}"
    EXTRA_NODES_FLAGS+=("--nodes=${MIN_NODES}:${MAX_NODES}:https://www.googleapis.com/compute/v1/projects/${PROJECT}/zones/${ZONE}/instanceGroups/${mig}")
done

sed -e "s|{{CA_IMAGE}}|${CA_IMAGE}|g" \
    -e "s|{{CONTROL_PLANE_NODE}}|${CONTROL_PLANE_NODE}|g" \
    -e "s|{{KUBERNETES_SERVICE_HOST}}|${KUBERNETES_SERVICE_HOST}|g" \
    -e "s|{{KUBERNETES_SERVICE_PORT}}|${KUBERNETES_SERVICE_PORT}|g" \
    -e "s|{{NODES_SPEC}}|${PRIMARY_NODES_SPEC}|g" \
    ${CA_ROOT}/hack/e2e/gce-deployment-template.yaml | while IFS= read -r line; do
    if [[ "${line}" == *"{{EXTRA_CA_FLAGS}}"* ]]; then
        for node_flag in "${EXTRA_NODES_FLAGS[@]}"; do
            echo "            - ${node_flag}"
        done
        for flag in ${EXTRA_CA_FLAGS}; do
            echo "            - ${flag}"
        done
    else
        printf "%s\n" "${line}"
    fi
done | kubectl apply -f -

echo "Deployed ${CA_IMAGE} to cluster."
