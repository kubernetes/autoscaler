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
# $ K8S_TAG=<k8s_tag> ./start_godeps_update.sh
# K8S_TAG - k8s version to use for the dependencies update.
# Suggested format is K8S_TAG=v1.10.3
set -o errexit
set -o pipefail
set -o nounset

K8S_TAG=${K8S_TAG:-v1.16.2}
K8S_FORK="git@github.com:kubernetes/kubernetes.git"

export GO111MODULE=on


echo "Updating Vertical Pod Autoscaler dependencies based on k8s ${K8S_TAG}"

WORKDIR=${HOME}/k8s-vpa-update-deps
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})
cd $SCRIPT_ROOT

echo "About to delete ${WORKDIR} dir. Hit enter to continue or"
echo "Ctrl-C to abort."
read

rm -rf ${WORKDIR}
mkdir ${WORKDIR}

echo Getting k8s
K8S_REPO=${WORKDIR}/kubernetes
pushd ${WORKDIR}
git clone ${K8S_FORK} ${K8S_REPO}

pushd ${K8S_REPO}
echo "Syncing k8s to ${K8S_TAG}"
git checkout ${K8S_TAG}
echo "Gen k8s bindata"
hack/generate-bindata.sh
popd
popd

echo "Preparing VPA go.mod file"

# Deleting old stuff
rm -rf vendor
rm -f go.mod
rm -f go.sum

# Base VPA go.mod on one from k8s.io/kuberntes
cp ${K8S_REPO}/go.mod .

# Check go version
REQUIRED_GO_VERSION=$(cat go.mod  |grep '^go ' |tr -s ' ' |cut -d ' '  -f 2)
USED_GO_VERSION=$(go version |sed 's/.*go\([0-9]\+\.[0-9]\+\).*/\1/')

if [[ "${REQUIRED_GO_VERSION}" != "${USED_GO_VERSION}" ]];then
  err_rerun "Invalid go version ${USED_GO_VERSION}; required go version is ${REQUIRED_GO_VERSION}."
fi

# Fix module name and staging modules links
sed -i "s#module k8s.io/kubernetes#module k8s.io/autoscaler/vertical-pod-autoscaler#" go.mod
sed -i "s#\\./staging#${K8S_REPO}/staging#" go.mod

# Add k8s.io/kubernetes dependency
go mod edit -require k8s.io/kubernetes@v0.0.0
go mod edit -replace k8s.io/kubernetes=${K8S_REPO}

echo "Running go mod vendor"
go mod vendor

echo "Running go build -mod=vendor"
if ! go build -mod=vendor ./... ; then
    echo "Build failed" | exit 1
fi

echo "Running go test -mod=vendor"
if ! go test -mod=vendor `go list ./... | grep -v "e2e/"` ; then
    echo "Test run failed" | exit 1
fi

# Commit go.mod* and vendor
git restore --staged .
git add vendor go.mod go.sum
if ! git diff --quiet --cached; then
  echo "Committing vendor, go.mod and go.sum"
  git commit -m "Updating vendor against k8s ${K8S_TAG}."
else
  echo "No changes after vendor update; skipping commit"
fi

if ! git diff --quiet; then
  echo "Uncommitted changes (manual fixes?) still present in repository - please commit those"
fi

