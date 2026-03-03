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

# Full local benchmark workflow: creates a Kind cluster, deploys VPA, installs
# KWOK, configures VPA for benchmarking, and runs the benchmark.
#
# Prerequisites: go, kind, kubectl, yq, docker
#
# Usage:
#   ./full-benchmark.sh [benchmark flags...]
#
# Examples:
#   ./full-benchmark.sh --profile=small
#   ./full-benchmark.sh --profile=small,medium,large --runs=3 --output=results.csv

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BENCHMARK_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
VPA_DIR="$(cd "${BENCHMARK_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${VPA_DIR}/.." && pwd)"

KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"

# Step 1: Create Kind cluster (skip if it already exists)
echo "=== Step 1: Kind cluster ==="
if kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER_NAME}$"; then
  echo "  Kind cluster '${KIND_CLUSTER_NAME}' already exists, skipping"
else
  KIND_CONFIG=$(mktemp)
  cp "${REPO_ROOT}/.github/kind-config.yaml" "${KIND_CONFIG}"
  cat >> "${KIND_CONFIG}" <<'EOF'
kubeadmConfigPatches:
- |
  kind: ClusterConfiguration
  apiServer:
    extraArgs:
      max-requests-inflight: "2000"
      max-mutating-requests-inflight: "1000"
  controllerManager:
    extraArgs:
      concurrent-replicaset-syncs: "500"
      kube-api-qps: "500"
      kube-api-burst: "1000"
EOF
  echo "  Creating Kind cluster '${KIND_CLUSTER_NAME}'..."
  kind create cluster --name "${KIND_CLUSTER_NAME}" --config="${KIND_CONFIG}"
fi

# Step 2: Deploy VPA
echo ""
echo "=== Step 2: Deploy VPA ==="
"${VPA_DIR}/hack/deploy-for-e2e-locally.sh" full-vpa

# Step 3: Install KWOK + create fake node
echo ""
echo "=== Step 3: Install KWOK ==="
"${SCRIPT_DIR}/install-kwok.sh"

# Step 4: Configure VPA deployments for benchmark
echo ""
echo "=== Step 4: Configure VPA ==="
"${SCRIPT_DIR}/configure-vpa.sh"

# Step 5: Build and run benchmark
echo ""
echo "=== Step 5: Build and run benchmark ==="
echo "  Building vpa-benchmark..."
go build -C "${BENCHMARK_DIR}" -o "${VPA_DIR}/bin/vpa-benchmark" .

echo "  Running benchmark..."
"${VPA_DIR}/bin/vpa-benchmark" "$@"
