#!/usr/bin/env bash

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

if [[ $# -ne 2 ]]; then
  echo "Usage: run-e2e-azure.sh IMAGE_REPOSITORY IMAGE_TAG" >&2
  exit 1
fi
CA_IMAGE_REPOSITORY="${1:?IMAGE_REPOSITORY must not be empty}"
CA_IMAGE_TAG="${2:?IMAGE_TAG must not be empty}"
: "${KUBECONFIG:?KUBECONFIG must point to the AKS workload cluster}"
: "${AZURE_SUBSCRIPTION_ID:?AZURE_SUBSCRIPTION_ID must be set}"
: "${AZURE_TENANT_ID:?AZURE_TENANT_ID must be set}"
: "${AZURE_RESOURCE_GROUP:?AZURE_RESOURCE_GROUP must be the AKS node resource group}"
: "${CLUSTER_AUTOSCALER_CLIENT_ID:?CLUSTER_AUTOSCALER_CLIENT_ID must identify the autoscaler workload identity}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CA_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
KUBECTL="${KUBECTL:-kubectl}"
HELM="${HELM:-helm}"
MIN_NODES="${CLUSTER_AUTOSCALER_MIN_NODES:-3}"
MAX_NODES="${CLUSTER_AUTOSCALER_MAX_NODES:-6}"
if [[ ! "${MIN_NODES}" =~ ^[1-9][0-9]*$ || ! "${MAX_NODES}" =~ ^[1-9][0-9]*$ ]] || (( MAX_NODES < MIN_NODES )); then
  echo "ERROR: node-group bounds must be positive integers with max >= min" >&2
  exit 1
fi

mkdir -p "${ARTIFACTS:-${CA_ROOT}/e2e/_artifacts}"
export ARTIFACTS="$(cd "${ARTIFACTS:-${CA_ROOT}/e2e/_artifacts}" && pwd)"
VALUES_FILE="$(mktemp)"
PHASE=standard

collect_autoscaler_logs() {
  "${KUBECTL}" --request-timeout=1m --namespace kube-system logs deployment/cluster-autoscaler \
    > "${ARTIFACTS}/cluster-autoscaler-${PHASE}.log"
}

on_exit() {
  local result=$?
  trap - EXIT
  if ! collect_autoscaler_logs; then
    echo "ERROR: failed to collect ${PHASE} autoscaler logs" >&2
    if [[ "${result}" -eq 0 ]]; then
      result=1
    fi
  fi
  rm -f "${VALUES_FILE}"
  exit "${result}"
}
trap on_exit EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

"${KUBECTL}" wait --for=condition=Ready nodes --all --timeout=10m
VMSS_NAME="$("${KUBECTL}" get nodes --selector=kubernetes.azure.com/mode=user,kubernetes.io/os=linux --output=json |
  jq -er --argjson min "${MIN_NODES}" '
    if (.items | length) != $min then
      error("expected the Linux user pool to start at its minimum node count")
    else
      [.items[].spec.providerID |
        (capture("^azure://.*/virtualMachineScaleSets/(?<name>[^/]+)/virtualMachines/[^/]+$"; "i") //
          error("expected an Azure VMSS provider ID on every Linux user node")).name] |
      unique |
      if length == 1 then .[0] else error("expected exactly one Linux user VMSS") end
    end')"

jq -n \
  --arg repository "${CA_IMAGE_REPOSITORY}" \
  --arg tag "${CA_IMAGE_TAG}" \
  --arg subscription "${AZURE_SUBSCRIPTION_ID}" \
  --arg tenant "${AZURE_TENANT_ID}" \
  --arg resourceGroup "${AZURE_RESOURCE_GROUP}" \
  --arg clientID "${CLUSTER_AUTOSCALER_CLIENT_ID}" \
  --arg vmss "${VMSS_NAME}" \
  --argjson min "${MIN_NODES}" --argjson max "${MAX_NODES}" \
  '{
    fullnameOverride: "cluster-autoscaler",
    cloudProvider: "azure",
    image: {repository: $repository, tag: $tag, pullPolicy: "Always"},
    azureSubscriptionID: $subscription,
    azureTenantID: $tenant,
    azureResourceGroup: $resourceGroup,
    azureVMType: "vmss",
    azureUseWorkloadIdentityExtension: true,
    autoscalingGroups: [{name: $vmss, minSize: $min, maxSize: $max}],
    podLabels: {"azure.workload.identity/use": "true"},
    nodeSelector: {"kubernetes.azure.com/mode": "system", "kubernetes.io/os": "linux"},
    tolerations: [{key: "CriticalAddonsOnly", operator: "Exists", effect: "NoSchedule"}],
    priorityClassName: "system-cluster-critical",
    rbac: {serviceAccount: {
      name: "cluster-autoscaler",
      annotations: {
        "azure.workload.identity/client-id": $clientID,
        "azure.workload.identity/tenant-id": $tenant
      }
    }},
    resources: {requests: {cpu: "100m", memory: "512Mi"}, limits: {memory: "1Gi"}},
    extraArgs: {
      v: 4,
      "logtostderr": true,
      "stderrthreshold": "info",
      "bypassed-scheduler-names": "non-existing-bypassed-scheduler",
      "unremovable-node-recheck-timeout": "1m",
      "scale-down-unneeded-time": "1m",
      "scale-down-delay-after-add": "1m"
    }
  }' > "${VALUES_FILE}"

"${HELM}" upgrade --install cluster-autoscaler "${CA_ROOT}/charts/cluster-autoscaler" \
  --namespace kube-system --values "${VALUES_FILE}" --wait --timeout 10m
"${KUBECTL}" wait --for=create configmap/cluster-autoscaler-status --namespace kube-system --timeout=5m

TIMEOUT=300m "${SCRIPT_DIR}/run-e2e-tests.sh" --report-prefix=standard_ \
  '\[sig-autoscaling\].*Cluster size autoscaling.*Standard Autoscaling' -- \
  --provider=local --non-blocking-taints=CriticalAddonsOnly --kubectl-path="${KUBECTL}"
collect_autoscaler_logs

PHASE=dra
"${KUBECTL}" apply -f "${CA_ROOT}/e2e/manifests/dra-driver.yaml"
"${KUBECTL}" rollout status daemonset/dra-example-driver-kubeletplugin --namespace kube-system --timeout=10m
"${KUBECTL}" rollout restart deployment/cluster-autoscaler --namespace kube-system
"${KUBECTL}" rollout status deployment/cluster-autoscaler --namespace kube-system --timeout=5m
# CA deletes this ConfigMap on exit; rollout readiness does not wait for its recreation.
"${KUBECTL}" wait --for=create configmap/cluster-autoscaler-status --namespace kube-system --timeout=5m

TIMEOUT=90m "${SCRIPT_DIR}/run-e2e-tests.sh" --report-prefix=dra_ \
  '\[sig-autoscaling\].*Cluster size autoscaling.*DRA Autoscaling' -- \
  --provider=local --non-blocking-taints=CriticalAddonsOnly --kubectl-path="${KUBECTL}"
