#!/bin/bash

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

set -o nounset
set -o pipefail

BASE_NAME=$(basename $0)
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

function print_help {
  echo "ERROR! Usage: $BASE_NAME <suite>"
  echo "<suite> should be one of:"
  echo " - recommender"
  echo " - recommender-externalmetrics"
  echo " - updater"
  echo " - admission-controller"
  echo " - actuation"
  echo " - full-vpa"
}

if [ $# -eq 0 ]; then
  print_help
  exit 1
fi

if [ $# -gt 1 ]; then
  print_help
  exit 1
fi

SUITE=$1
REQUIRED_COMMANDS="
docker
go
kind
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


echo "Deleting KIND cluster 'kind'."
kind delete cluster -n kind -q

echo "Creating KIND cluster 'kind'"
KIND_VERSION="kindest/node:v1.32.0"
if ! kind create cluster --image=${KIND_VERSION}; then
    echo "Failed to create KIND cluster. Exiting. Make sure kind version is updated."
    echo "Available versions: https://github.com/kubernetes-sigs/kind/releases"
    exit 1
fi

echo "Building metrics-pump image"
docker build -t localhost:5001/write-metrics:dev -f ${SCRIPT_ROOT}/hack/e2e/Dockerfile.externalmetrics-writer ${SCRIPT_ROOT}/hack
echo "  loading image into kind"
kind load docker-image localhost:5001/write-metrics:dev


case ${SUITE} in
  recommender|recommender-externalmetrics|updater|admission-controller|actuation|full-vpa)
    ${SCRIPT_ROOT}/hack/vpa-down.sh
    echo " ** Deploying for suite ${SUITE}"
    ${SCRIPT_ROOT}/hack/deploy-for-e2e-locally.sh ${SUITE}

    echo " ** Running suite ${SUITE}"
    if [ ${SUITE} == recommender-externalmetrics ]; then
       WORKSPACE=./workspace/_artifacts ${SCRIPT_ROOT}/hack/run-e2e-tests.sh recommender
    else
      WORKSPACE=./workspace/_artifacts ${SCRIPT_ROOT}/hack/run-e2e-tests.sh ${SUITE}
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac

