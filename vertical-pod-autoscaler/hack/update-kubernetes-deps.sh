#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
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

# Usage:
# $ K8S_TAG=<k8s_tag> ./hack/update-kubernetes-deps-in-e2e.sh
# K8S_TAG - k8s version to use for the dependencies update.
# Suggested format is K8S_TAG=v1.10.3

set -euo pipefail

K8S_TAG=${K8S_TAG:-v1.33.0}
K8S_TAG=${K8S_TAG#v}
K8S_FORK="git@github.com:kubernetes/kubernetes.git"

export GO111MODULE=on

function update_deps() {
    # list staged k8s.io repos
    MODS=($(
        curl -sS https://raw.githubusercontent.com/kubernetes/kubernetes/v${K8S_TAG}/go.mod |
        sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
    ))

    # get matching tag for each staged k8s.io repo
    for MOD in "${MODS[@]}"; do
        V=$(
            go mod download -json "${MOD}@kubernetes-${K8S_TAG}" |
            sed -n 's|.*"Version": "\(.*\)".*|\1|p'
        )
        echo "Replacing ${MOD} with version ${V}"
        go mod edit "-replace=${MOD}=${MOD}@${V}"
    done
}

# execute in subshell to keep CWD even in case of failures
(
    # find script directory invariantly of CWD
    DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

    # cd to e2e tests
    cd ${DIR}/..

    echo "Updating VPA dependencies to k8s ${K8S_TAG}"
    update_deps

    echo "Updating k8s to ${K8S_TAG}"
    go get "k8s.io/kubernetes@v${K8S_TAG}"

    echo "Running go mod tidy and vendoring deps"
    # tidy and vendor modules
    go mod tidy
    go mod vendor
)
