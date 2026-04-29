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

BASE_NAME=$(basename "$0")
SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
KIND_CONFIG="${SCRIPT_ROOT}/../.github/kind-config.yaml"

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
helm
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

# Clean up before exit
function cleanup {
  echo " ** Cleaning up..."
  helm uninstall vpa --namespace kube-system 2>/dev/null || true
  kubectl delete namespace monitoring --ignore-not-found=true 2>/dev/null || true
}
trap cleanup EXIT

echo "Deleting KIND cluster 'kind'."
kind delete cluster -n kind -q

if [ ! -f "${KIND_CONFIG}" ]; then
  echo "Missing KIND config file: ${KIND_CONFIG}"
  exit 1
fi

echo "Creating KIND cluster 'kind'"
if ! kind create cluster --config "${KIND_CONFIG}"; then
    echo "Failed to create KIND cluster using ${KIND_CONFIG}. Exiting."
    exit 1
fi

# Build and deploy external metrics writer if needed
if [[ "${SUITE}" == "recommender-externalmetrics" ]]; then
  echo " ** Building external metrics writer image"
  docker build -t localhost:5001/write-metrics:dev -f "${SCRIPT_ROOT}"/hack/e2e/Dockerfile.externalmetrics-writer "${SCRIPT_ROOT}"/hack
  kind load docker-image localhost:5001/write-metrics:dev
fi

export FEATURE_GATES=""
export TEST_WITH_FEATURE_GATES_ENABLED=""

if [ "${ENABLE_ALL_FEATURE_GATES:-}" == "true" ] ; then
  export FEATURE_GATES='AllAlpha=true,AllBeta=true'
  export TEST_WITH_FEATURE_GATES_ENABLED="true"
fi

case ${SUITE} in
  recommender|recommender-externalmetrics|updater|admission-controller|actuation|full-vpa)
    # Checking if user specified artifact directory to dump logs
    if [[ -z "${ARTIFACTS:-}" ]]; then
      # Create temp dir for artifacts
      ARTIFACTS=$(mktemp -d)
      echo " ** Log artifacts will be stored in ${ARTIFACTS}"
    fi

    echo " ** Deploying VPA components..."
    "${SCRIPT_ROOT}"/hack/deploy-for-e2e-locally.sh "${SUITE}"

    echo " ** Running E2E tests..."
    if [ "${SUITE}" == recommender-externalmetrics ]; then
       ARTIFACTS="${ARTIFACTS}" "${SCRIPT_ROOT}"/hack/run-e2e-tests.sh recommender
    else 
       ARTIFACTS="${ARTIFACTS}" "${SCRIPT_ROOT}"/hack/run-e2e-tests.sh "${SUITE}"
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac
