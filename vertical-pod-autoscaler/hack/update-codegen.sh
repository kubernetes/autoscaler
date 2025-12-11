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

GO_CMD=${1:-go}
CURRENT_DIR=$(dirname "${BASH_SOURCE[0]}")
cd "${CURRENT_DIR}/.."
go mod download
CODEGEN_PKG=$($GO_CMD list -m -mod=readonly -f "{{.Dir}}" k8s.io/code-generator)

REPO_ROOT="$(git rev-parse --show-toplevel)"
VPA_ROOT="${REPO_ROOT}/vertical-pod-autoscaler/"

# shellcheck source=/dev/null
source "${CODEGEN_PKG}/kube_codegen.sh"

# Ensure openapi-gen is installed
OPENAPI_PKG=$($GO_CMD list -m -mod=readonly -f "{{.Dir}}" k8s.io/kube-openapi)
$GO_CMD install "${OPENAPI_PKG}/cmd/openapi-gen"

kube::codegen::gen_helpers \
    "$(dirname ${BASH_SOURCE})/../pkg/apis" \
    --boilerplate "${REPO_ROOT}/hack/boilerplate/boilerplate.generatego.txt"

echo "Ran gen helpers, moving on to generating openapi definitions..."

# Generate OpenAPI definitions
OPENAPI_OUTPUT_DIR="${VPA_ROOT}/pkg/generated/openapi"
mkdir -p "${OPENAPI_OUTPUT_DIR}"

openapi-gen \
    --go-header-file "${REPO_ROOT}/hack/boilerplate/boilerplate.generatego.txt" \
    --output-dir "${OPENAPI_OUTPUT_DIR}" \
    --output-pkg "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/generated/openapi" \
    --output-file "zz_generated.openapi.go" \
    --report-filename "${OPENAPI_OUTPUT_DIR}/violation_exceptions.list" \
    "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1" \
    "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1" \
    "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2" \
    "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1" \
    "k8s.io/api/core/v1" \
    "k8s.io/api/autoscaling/v1" \
    "k8s.io/apimachinery/pkg/api/resource" \
    "k8s.io/apimachinery/pkg/apis/meta/v1" \
    "k8s.io/apimachinery/pkg/runtime" \
    "k8s.io/apimachinery/pkg/version"

echo "Generated openapi definitions, moving on to generating client code..."

# Generate OpenAPI schema JSON for applyconfiguration-gen
OPENAPI_SCHEMA_FILE=$(mktemp)
trap "rm -f ${OPENAPI_SCHEMA_FILE}" EXIT
echo  "${VPA_ROOT}/pkg/generated/openapi/cmd/models-schema"
$GO_CMD run "${VPA_ROOT}/pkg/generated/openapi/cmd/models-schema" > "${OPENAPI_SCHEMA_FILE}"

kube::codegen::gen_client \
  "$(dirname ${BASH_SOURCE})/../pkg/apis" \
  --output-pkg k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client \
  --output-dir "$(dirname ${BASH_SOURCE})/../pkg/client" \
  --boilerplate "${REPO_ROOT}/hack/boilerplate/boilerplate.generatego.txt" \
  --with-watch \
  --with-applyconfig \
  --applyconfig-openapi-schema "${OPENAPI_SCHEMA_FILE}"

echo "Generated client code, running \`go mod tidy\`..."

# We need to clean up the go.mod file since code-generator adds temporary library to the go.mod file.
"${GO_CMD}" mod tidy
