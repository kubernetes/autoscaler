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

set -o errexit
set -o nounset
set -o pipefail

BASE_NAME=$(basename $0)
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/../..
source "${SCRIPT_ROOT}/hack/lib/util.sh"

ARCH=$(kube::util::host_arch)

function print_help {
  echo "ERROR! Usage: $BASE_NAME <suite>"
  echo "<suite> should be one of:"
  echo " - recommender"
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

case ${SUITE} in
  recommender)
    COMPONENTS="${SUITE}"
    ;;
  *)
    print_help
    exit 1
    ;;
esac

echo "Deleting KIND cluster 'kind'."
kind delete cluster -n kind -q

echo "Creating KIND cluster 'kind'"
KIND_VERSION="kindest/node:v1.32.0"
if ! kind create cluster --image=${KIND_VERSION}; then
    echo "Failed to create KIND cluster. Exiting. Make sure kind version is updated."
    echo "Available versions: https://github.com/kubernetes-sigs/kind/releases"
    exit 1
fi

# Local KIND images
export REGISTRY=${REGISTRY:-localhost:5001}
export TAG=${TAG:-latest}

# Deploy common resources
rm -f ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
patch -c ${SCRIPT_ROOT}/deploy/vpa-rbac.yaml -i ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.diff -o ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
# Other-versioned CRDs are irrelevant as we're running a modern-ish cluster.
kubectl apply -f ${SCRIPT_ROOT}/deploy/vpa-v1-crd-gen.yaml
kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/k8s-metrics-server.yaml

for i in ${COMPONENTS}; do
  ALL_ARCHITECTURES=${ARCH} make --directory ${SCRIPT_ROOT}/pkg/${i} docker-build REGISTRY=${REGISTRY} TAG=${TAG}
  docker tag ${REGISTRY}/vpa-${i}-${ARCH}:${TAG} ${REGISTRY}/vpa-${i}:${TAG}
  kind load docker-image ${REGISTRY}/vpa-${i}:${TAG}
done


case ${SUITE} in
  recommender)

    echo " ** Running suite ${SUITE}"

    ASSERTION_SCRIPTS=$(find "${SCRIPT_ROOT}/hack/integration-tests/deploy/${SUITE}" -name "*.sh" -type f | sort)
    for assert_script in $ASSERTION_SCRIPTS; do
      echo "======================================================================================================"
      echo "Running assertion script $(basename "$assert_script")"
      echo "======================================================================================================"
      assert_start=$(date +%s)
      $assert_script
      # TODO: Cleanup on failure
      # TODO: Add option to skip cleanup for debugging
      assert_end=$(date +%s)
      echo "Took $((assert_end - assert_start))sec"
      echo "Assertion test $assert_script PASSED!"
      # TODO: Define a way to cleanup resources between tests
    done
    ;;
  *)
    print_help
    exit 1
    ;;
esac
