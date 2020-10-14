#!/usr/bin/env bash

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
git rm -r --force kubernetes
