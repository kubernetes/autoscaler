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

# This script is the entrypoint for the CA presubmit E2E job.
# It detects the target branch from PULL_BASE_REF (set by Prow) and
# derives the correct Kubernetes version from go.mod, then passes
# --extract=ci/latest-X.Y to kubernetes_e2e.py so that the cluster
# matches the k8s version the branch is built against.

set -o nounset
set -o pipefail
set -o errexit

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
CA_ROOT="$(readlink -f "${SCRIPT_DIR}/../..")"

# Derive --extract value from go.mod.
# go.mod always has the correct k8s version for the current branch.
# e.g. "k8s.io/kubernetes v1.33.2" → "ci/latest-1.33"
K8S_VERSION=$(grep "k8s.io/kubernetes" "${CA_ROOT}/go.mod" | awk '{print $2}' | sed 's/v\([0-9]*\.[0-9]*\)\..*/\1/')

if [[ -z "${K8S_VERSION}" ]]; then
  echo "ERROR: Could not determine k8s version from go.mod" >&2
  exit 1
fi

EXTRACT="ci/latest-${K8S_VERSION}"
echo "### Detected Kubernetes version: ${K8S_VERSION} (branch: ${PULL_BASE_REF:-master})"
echo "### Using --extract=${EXTRACT}"

# Call kubernetes_e2e.py with the correct --extract flag.
exec /workspace/scenarios/kubernetes_e2e.py \
  --cluster=ca \
  --extract="${EXTRACT}" \
  --gcp-node-image=gci \
  --gcp-nodes=3 \
  --gcp-zone=us-central1-b \
  --provider=gce \
  --test=false \
  --test-cmd="${CA_ROOT}/hack/e2e/run-e2e.sh" \
  --test-cmd-args=--presubmit \
  --timeout=400m
