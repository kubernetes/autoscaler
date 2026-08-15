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

# Shared Helm configuration for E2E deployments. Sourced by both
# deploy-for-e2e.sh (CI) and deploy-for-e2e-locally.sh.
#
# The caller MUST have set SCRIPT_ROOT to the repo's vertical-pod-autoscaler/
# directory before sourcing this file.

if [[ -z "${SCRIPT_ROOT:-}" ]]; then
  echo "ERROR: helm-settings.sh expects SCRIPT_ROOT to be set by the caller" >&2
  return 1 2>/dev/null || exit 1
fi

HELM_CHART_PATH="${SCRIPT_ROOT}/charts/vertical-pod-autoscaler"
VALUES_FILE="${SCRIPT_ROOT}/hack/e2e/values-e2e.yaml"
HELM_RELEASE_NAME="vpa"
HELM_NAMESPACE="kube-system"
