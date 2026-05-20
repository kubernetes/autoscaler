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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

verify_go_mod() {
  local module_dir="$1"

  if [[ ! -f "${module_dir}/go.mod" ]]; then
    echo "Error: ${module_dir}/go.mod does not exist"
    return 1
  fi
  if [[ ! -f "${module_dir}/go.sum" ]]; then
    echo "Error: ${module_dir}/go.sum does not exist"
    return 1
  fi

  echo "Verifying ${module_dir}/go.mod and ${module_dir}/go.sum"
  (
    cd "${module_dir}"
    go mod tidy -diff
  )
}

ret=0

verify_go_mod "${SCRIPT_ROOT}" || ret=$?
verify_go_mod "${SCRIPT_ROOT}/test" || ret=$?

if [[ $ret -eq 0 ]]
then
  echo "go.mod and go.sum are up to date."
else
  echo "go.mod and go.sum are out of date. Please run 'go mod tidy' in the affected directories"
  exit 1
fi
