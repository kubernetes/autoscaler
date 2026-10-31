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

set -o errexit
set -o nounset
set -o pipefail

echo "verify-kubelint"

cd "$(dirname "${BASH_SOURCE[0]}")/.."
TOOLS_DIR="${PWD}/hack/tools/kube-api-linter"

# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
GOLANGCI_LINT_VERSION=v2.15.0

echo "installing golangci-lint ${GOLANGCI_LINT_VERSION}"
GOBIN="${TOOLS_DIR}/bin" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"

echo "creating custom golangci linter"
(
  cd "${TOOLS_DIR}"
  ./bin/golangci-lint custom
)

"${TOOLS_DIR}/bin/golangci-kube-api-linter" run -v --config "${TOOLS_DIR}/.golangci-kal.yml"
