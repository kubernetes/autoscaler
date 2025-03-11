#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../code-generator)}

source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
  --input-pkg-root k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis \
  --output-base "$(dirname ${BASH_SOURCE})/../../../.." \
  --boilerplate "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

kube::codegen::gen_client \
  --input-pkg-root k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis \
  --output-pkg-root k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client \
  --output-base "$(dirname ${BASH_SOURCE})/../../../.." \
  --boilerplate "${SCRIPT_ROOT}"/hack/boilerplate.go.txt \
  --with-watch