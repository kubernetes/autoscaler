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

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
CA_ROOT="$(readlink -f "${SCRIPT_DIR}/../..")"

# Parse flags
REMAINING_ARGS=()
while [[ $# -gt 0 ]]; do
  case $1 in
    --presubmit)
      export EXTRA_CA_FLAGS="${EXTRA_CA_FLAGS:-} --unremovable-node-recheck-timeout=1m --scale-down-unneeded-time=1m --scale-down-delay-after-add=1m"
      shift
      ;;
    *)
      REMAINING_ARGS+=("$1")
      shift
      ;;
  esac
done

# Build and push image once to be used in both steps.
GIT_COMMIT="$(git describe --always --dirty --exclude '*')"
export TAG="dev-${GIT_COMMIT}-$(date +%s)"

echo "### STEP 1: Standard Autoscaling tests ###"
${CA_ROOT}/hack/e2e/deploy-ca-on-gce-for-e2e.sh
${CA_ROOT}/hack/e2e/run-e2e-tests.sh "Standard Autoscaling"

echo "### STEP 2: DRA Autoscaling tests ###"
echo "Removing Cluster Autoscaler to reset its state..."
kubectl delete deployment cluster-autoscaler -n kube-system

echo "Installing DRA driver..."
kubectl apply -f "${CA_ROOT}/e2e/manifests/dra-driver.yaml"

echo "Redeploying Cluster Autoscaler..."
# Reuse CA image.
REGISTRY="gcr.io/$(gcloud config get core/project)"
export CA_IMAGE="${REGISTRY}/cluster-autoscaler:${TAG}"
${CA_ROOT}/hack/e2e/deploy-ca-on-gce-for-e2e.sh

echo "Running DRA tests..."
${CA_ROOT}/hack/e2e/run-e2e-tests.sh "DRA Autoscaling"
