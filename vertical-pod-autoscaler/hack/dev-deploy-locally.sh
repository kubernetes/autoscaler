#!/bin/bash

# Copyright 2025 The Kubernetes Authors.
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

set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

SUITE=full-vpa
REQUIRED_COMMANDS="
docker
git
go
kubectl
make
"

for i in $REQUIRED_COMMANDS; do
  if ! command -v $i > /dev/null 2>&1
  then
    echo "$i could not be found, please ensure it is installed"
    echo
    echo "The following commands are required to run these tests:"
    echo $REQUIRED_COMMANDS
    exit 1;
  fi
done

if ! docker ps >/dev/null 2>&1
then
  echo "docker isn't running"
  echo
  echo "Please ensure that docker is running"
  exit 1
fi

if ! kubectl version >/dev/null 2>&1
then
  echo "Kubernetes isn't running"
  echo
  echo "Please ensure that Kubernetes is running"
  exit 1
fi

COMMIT=$(git rev-parse HEAD 2>/dev/null)
COMMIT=${COMMIT:0:14}
export BUILD_LD_FLAGS="-s -X=k8s.io/autoscaler/vertical-pod-autoscaler/common.gitCommit=$COMMIT"
export TAG=$COMMIT


echo " ** Deploying building and deploying all VPA components"
${SCRIPT_ROOT}/hack/deploy-for-e2e-locally.sh full-vpa

echo " ** Restarting all VPA components"
for i in admission-controller updater recommender; do
  kubectl -n kube-system rollout restart deployment/vpa-$i
done
