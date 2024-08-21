#!/usr/bin/env bash

# Copyright 2021 The Kubernetes Authors.
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

###
# This script is to be used when updating the generated clients of 
# the Provisioning Request CRD.
###

set -o errexit
set -o nounset
set -o pipefail

GO_CMD=${1:-go}
CURRENT_DIR=$(dirname "${BASH_SOURCE[0]}")
REPO_ROOT="$(git rev-parse --show-toplevel)"
CODEGEN_PKG=$($GO_CMD list -m -mod=readonly -f "{{.Dir}}" k8s.io/code-generator)
cd "${CURRENT_DIR}/.."

# shellcheck source=/dev/null
source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate "${REPO_ROOT}/hack/boilerplate/boilerplate.generatego.txt" \
    "${REPO_ROOT}/cluster-autoscaler/apis/provisioningrequest"

echo "Ran gen helpers, moving on to generating client code..."

kube::codegen::gen_client \
    --output-pkg k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client \
    --output-dir "${REPO_ROOT}/cluster-autoscaler/apis/provisioningrequest/client" \
    --boilerplate "${REPO_ROOT}/hack/boilerplate/boilerplate.generatego.txt" \
    --with-watch \
    --with-applyconfig \
    "${REPO_ROOT}/cluster-autoscaler/apis/provisioningrequest"

echo "Generated client code, running `go mod tidy`..."

# We need to clean up the go.mod file since code-generator adds temporary library to the go.mod file.
"${GO_CMD}" mod tidy
