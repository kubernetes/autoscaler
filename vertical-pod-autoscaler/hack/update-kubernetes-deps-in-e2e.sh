#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

set -eou errexit pipefail nounset

K8S_TAG=${K8S_TAG:-v1.16.2}
K8S_TAG=${K8S_TAG#v}

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
        go mod edit "-replace=${MOD}=${MOD}@${V}"
    done
    
    # update k8s.io/kubernetes to desired tag
    go get "k8s.io/kubernetes@v${K8S_TAG}"
}


# execute in subshell to keep CWD even in case of failures
(
    # find script directory invariantly of CWD
    DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

    # cd to e2e tests
    cd $DIR/../e2e

    # update k8s deps in e2e
    update_deps

    # tidy and vendor modules
    go mod tidy
    go mod vendor
)
