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
# This script is to be used as a break-glass solution if there is a breaking
# change in a release of Kubernetes. This allows us to switch to an unreleased
# commit by submoduling the whole k/k repository.
###

set -o errexit
set -o pipefail

VERSION=${1}
FORK=${2:-git@github.com:kubernetes/kubernetes.git}
if [ -z "$VERSION" ]; then
    echo "Usage: hack/submodule-k8s.sh <k8s sha> <k8s fork:-git@github.com:kubernetes/kubernetes.git>"
    exit 1
fi

set -x

WORKDIR=$(mktemp -d)
REPO="${WORKDIR}/kubernetes"
git clone --depth 1 ${FORK} ${REPO}

pushd ${REPO}
git fetch --depth 1 origin v${VERSION}
git checkout FETCH_HEAD

MODS=($(
    cat go.mod | sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))

popd
rm -rf ${WORKDIR}

git submodule add --force https://github.com/kubernetes/kubernetes
git submodule update --init --recursive --remote
cd kubernetes
git checkout $VERSION
cd ..

go mod edit "-replace=k8s.io/kubernetes=./kubernetes"

for MOD in "${MODS[@]}"; do
    go mod edit "-replace=${MOD}=./kubernetes/staging/src/${MOD}"
done
go mod vendor
go mod tidy
