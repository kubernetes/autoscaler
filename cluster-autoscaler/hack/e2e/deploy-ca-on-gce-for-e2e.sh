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
# 1. 'gcloud' installed and authenticated.
# 2. A GCE cluster created with 'kubetest2' (or legacy kube-up.sh).
# 3. KUBECONFIG specified of available at $HOME/.kube/config.
#
# This script relies on specific naming conventions for nodes and MIGs common in those setups:
# - Control-plane node can be identified by its taints ("node-role.kubernetes.io/control-plane").
# - Only one MIG, the nodes follow a pattern where stripping the last segment yields the MIG name.
# - ProviderID format is gce://<project>/<zone>/<name>

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
CA_ROOT="$(readlink -f "${SCRIPT_DIR}/../..")"
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

# Default to a tag based on git commit. If there are local changes, add -dirty and a timestamp to avoid push collisions.
GIT_COMMIT="$(git describe --always --dirty --exclude '*')"
TAG="${TAG:-dev-${GIT_COMMIT}-$(date +%s)}"
REGISTRY="gcr.io/$(gcloud config get core/project)"

echo "Configuring registry authentication..."
mkdir -p "${HOME}/.docker"
gcloud auth configure-docker -q

echo "Building and pushing image..."
pushd "${CA_ROOT}" >/dev/null
make execute-release REGISTRY=${REGISTRY} TAG=${TAG}
IMAGE="${REGISTRY}/cluster-autoscaler:${TAG}"
popd >/dev/null

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

# Find a worker node to guess the MIG name
WORKER_NODE="$(kubectl get nodes --no-headers -o custom-columns=NAME:.metadata.name | grep -v -E 'master|control-plane' | head -n 1)"
# MIG name is the worker node name without the last random segment
MIG_NAME="$(echo ${WORKER_NODE} | sed 's/-[^-]*$//')"
echo "Identified MIG name: ${MIG_NAME}"

# 3:6 limits are the defaults used in job configs
MIN_NODES="${MIN_NODES:-3}"
MAX_NODES="${MAX_NODES:-6}"
# Extra cluster-autoscaler flags.
EXTRA_CA_FLAGS="${EXTRA_CA_FLAGS:-""}"

# Construct the full URL as requested
NODES_SPEC="${MIN_NODES}:${MAX_NODES}:https://www.googleapis.com/compute/v1/projects/${PROJECT}/zones/${ZONE}/instanceGroups/${MIG_NAME}"
echo "Nodes spec: ${NODES_SPEC}"

sed -e "s|{{IMAGE}}|${IMAGE}|g" \
    -e "s|{{CONTROL_PLANE_NODE}}|${CONTROL_PLANE_NODE}|g" \
    -e "s|{{KUBERNETES_SERVICE_HOST}}|${KUBERNETES_SERVICE_HOST}|g" \
    -e "s|{{KUBERNETES_SERVICE_PORT}}|${KUBERNETES_SERVICE_PORT}|g" \
    -e "s|{{NODES_SPEC}}|${NODES_SPEC}|g" \
    ${CA_ROOT}/hack/e2e/gce-deployment-template.yaml | while IFS= read -r line; do
    if [[ "${line}" == *"{{EXTRA_CA_FLAGS}}"* ]]; then
        for flag in ${EXTRA_CA_FLAGS}; do
            echo "            - ${flag}"
        done
    else
        printf "%s\n" "${line}"
    fi
done | kubectl apply -f -

echo "Deployed ${IMAGE} to cluster."
