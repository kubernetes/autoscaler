#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
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


# os::util::absolute_path returns the absolute path to the directory provided
function os::util::absolute_path() {
        local relative_path="$1"
        local absolute_path

        pushd "${relative_path}" >/dev/null
        relative_path="$( pwd )"
        if [[ -h "${relative_path}" ]]; then
                absolute_path="$( readlink "${relative_path}" )"
        else
                absolute_path="${relative_path}"
        fi
        popd >/dev/null

        echo "${absolute_path}"
}
readonly -f os::util::absolute_path

# find the absolute path to the root of the Origin source tree
init_source="$( dirname "${BASH_SOURCE}" )/../.."
OS_ROOT="$( os::util::absolute_path "${init_source}" )"
export OS_ROOT
cd "${OS_ROOT}"

PROJECT_PREFIX="k8s.io/autoscaler/cluster-autoscaler"
OS_OUTPUT_BINPATH="${OS_ROOT}/_output/bin"
