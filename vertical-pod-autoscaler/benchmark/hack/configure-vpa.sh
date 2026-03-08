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

# Configures VPA deployments with benchmark-specific settings:
# - Increased QPS/burst on all components (to avoid client-side throttling)
# - Longer updater interval (steps can take longer than the default 60s at scale)
#
# Prerequisites: helm
#
# Usage: ./configure-vpa.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VPA_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

HELM_RELEASE_NAME="${HELM_RELEASE_NAME:-vpa}"
HELM_NAMESPACE="${HELM_NAMESPACE:-kube-system}"
HELM_CHART_PATH="${VPA_DIR}/charts/vertical-pod-autoscaler"
VALUES_FILE="${SCRIPT_DIR}/values-benchmark.yaml"

echo "=== Configuring VPA deployments for benchmark ==="

helm upgrade "${HELM_RELEASE_NAME}" "${HELM_CHART_PATH}" \
  --namespace "${HELM_NAMESPACE}" \
  --values "${VALUES_FILE}" \
  --reuse-values \
  --wait

echo "=== VPA configuration complete ==="
