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
# This script is to be used when updating Kubernetes and its staging
# repositories to *tagged* releases. This is the ideal case, but another
# script, submodule-k8s.sh, is available as a break-glass solution if we must
# switch to an unreleased commit.
###

set -o errexit
set -o pipefail

VERSION=${1#"v"}
if [ -z "$VERSION" ]; then
    echo "Usage: hack/update-vendor.sh <k8s version>"
    exit 1
fi

set -x

MODS=($(
    curl -sS https://raw.githubusercontent.com/kubernetes/kubernetes/v${VERSION}/go.mod |
    sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))

for MOD in "${MODS[@]}"; do
    V=$(
        go mod download -json "${MOD}@kubernetes-${VERSION}" |
        sed -n 's|.*"Version": "\(.*\)".*|\1|p'
    )
    go mod edit "-replace=${MOD}=${MOD}@${V}"
done
go get "k8s.io/kubernetes@v${VERSION}"
go mod vendor
go mod tidy
git rm -r --force --ignore-unmatch kubernetes
