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

# Helpers for the Go tools tracked in hack/tools/go.mod.
#
# Usage:
#   source "${ROOT}/hack/lib/tools.sh"
#   CONTROLLER_GEN=$(vpa::tool controller-gen)

VPA_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

# Builds the named tool if needed and prints the path to its binary.
# Build/download progress goes to stderr, so callers can safely capture stdout.
vpa::tool() {
  (
    # -modfile needs a go.mod in the working directory.
    cd "${VPA_ROOT}"
    GOWORK=off go tool -modfile="${VPA_ROOT}/hack/tools/go.mod" -n "$1"
  )
}
