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
# Prerequisites: kubectl, yq
#
# Usage: ./configure-vpa.sh

set -euo pipefail

VPA_NAMESPACE="${VPA_NAMESPACE:-kube-system}"

echo "=== Configuring VPA deployments for benchmark ==="

echo "  Configuring vpa-recommender (QPS=100, burst=200, memory-saver=true)..."
kubectl get deployment vpa-recommender -n "${VPA_NAMESPACE}" -o yaml | \
  yq '.spec.template.spec.containers[0].args = ["--kube-api-qps=100", "--kube-api-burst=200", "--memory-saver=true"]' | \
  kubectl apply -f -

echo "  Configuring vpa-updater (QPS=100, burst=200, updater-interval=2m)..."
kubectl get deployment vpa-updater -n "${VPA_NAMESPACE}" -o yaml | \
  yq '.spec.template.spec.containers[0].args = ["--kube-api-qps=100", "--kube-api-burst=200", "--updater-interval=2m"]' | \
  kubectl apply -f -

echo "  Configuring vpa-admission-controller (QPS=100, burst=200)..."
kubectl get deployment vpa-admission-controller -n "${VPA_NAMESPACE}" -o yaml | \
  yq '.spec.template.spec.containers[0].args = ["--kube-api-qps=100", "--kube-api-burst=200"]' | \
  kubectl apply -f -

echo "=== VPA configuration complete ==="
