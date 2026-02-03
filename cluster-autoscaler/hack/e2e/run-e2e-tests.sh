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

set -o nounset
set -o pipefail
set -o errexit

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
CA_ROOT="$(readlink -f "${SCRIPT_DIR}/../..")"

function print_help {
  echo "Usage: run-e2e-tests.sh [FOCUS]"
  echo "  FOCUS: Optional ginkgo focus string. Defaults to '\[sig-autoscaling\]'"
  echo "  Environment variables:"
  echo "    KUBECONFIG: Path to kubeconfig (default: ~/.kube/config)"
  echo "    ARTIFACTS: Directory for report artifacts (default: e2e/_artifacts)"
  echo "    SKIP: Optional ginkgo skip string"
  echo "    TIMEOUT: Timeout for ginkgo tests (default: 400m)"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    print_help
    exit 0
fi

FOCUS="${1:-"\\[sig-autoscaling\\]"}"
# 400m allocates ~20 mins per each test case.
TIMEOUT="${TIMEOUT:-400m}"

ABSOLUTE_PATH="$(cd "${CA_ROOT}"; pwd)"
export GOBIN="${ABSOLUTE_PATH}/e2e/_output/bin"
export ARTIFACTS="${ARTIFACTS:-"${ABSOLUTE_PATH}/e2e/_artifacts"}"

export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

pushd "${CA_ROOT}/e2e" >/dev/null

echo "Installing ginkgo..."
mkdir -p "${GOBIN}"
go install github.com/onsi/ginkgo/v2/ginkgo

echo "Building e2e tests..."
"${GOBIN}/ginkgo" build .

echo "Running e2e tests with focus: ${FOCUS}"
"${GOBIN}/ginkgo" -v --timeout="${TIMEOUT}" --focus="${FOCUS}" ./e2e.test -- --report-dir="${ARTIFACTS}" --disable-log-dump ${SKIP:-}
RESULT=$?

popd >/dev/null

echo "Test result: ${RESULT}"
if [ $RESULT -gt 0 ]; then
  echo "Tests failed"
  exit 1
fi
