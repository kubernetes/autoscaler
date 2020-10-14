#!/usr/bin/env bash

###

set -o errexit
set -o pipefail

###
# This script is to be used as a break-glass solution if there is a breaking
# change in a release of Kubernetes. This allows us to switch to an unreleased
# commit by submoduling the whole k/k repository.
###

VERSION=${1}
if [ -z "$VERSION" ]; then
    echo "Usage: hack/submodule-k8s.sh <k8s sha>"
    exit 1
fi

set -x

MODS=($(
    curl -sS https://raw.githubusercontent.com/kubernetes/kubernetes/${VERSION}/go.mod |
    sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))

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
