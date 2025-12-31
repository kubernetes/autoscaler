#!/usr/bin/env bash

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

# This script sets up and runs the VPA integration tests using controller-runtime's envtest.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VPA_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
INTEGRATION_DIR="${VPA_DIR}/integration"

# Kubernetes version to use for envtest binaries
# This should match the k8s.io/* dependency versions in go.mod
ENVTEST_K8S_VERSION="${ENVTEST_K8S_VERSION:-1.35.x}"

# Directory to store envtest binaries
ENVTEST_ASSETS_DIR="${ENVTEST_ASSETS_DIR:-${HOME}/.local/share/kubebuilder-envtest}"

echo "==> Setting up envtest environment..."

# Check if setup-envtest is installed
if ! command -v setup-envtest &> /dev/null; then
    echo "==> Installing setup-envtest..."
    go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
fi

# Setup envtest binaries and get the path
echo "==> Downloading envtest binaries for Kubernetes ${ENVTEST_K8S_VERSION}..."
KUBEBUILDER_ASSETS="$(setup-envtest use "${ENVTEST_K8S_VERSION}" --bin-dir "${ENVTEST_ASSETS_DIR}" -p path)"
export KUBEBUILDER_ASSETS

echo "==> Using envtest binaries from: ${KUBEBUILDER_ASSETS}"

# Change to integration test directory
cd "${INTEGRATION_DIR}"

# Run the tests
echo "==> Running integration tests..."
go test -tags=integration -v -timeout 300s -parallel 4 "$@" ./...
