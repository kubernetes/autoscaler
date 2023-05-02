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

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

function print_help {
  echo "ERROR! Usage: run-e2e.sh <suite>"
  echo "<suite> should be one of:"
  echo " - recommender"
  echo " - recommender-externalmetrics"
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

echo "Deleting KIND cluster 'kind'."
kind delete cluster -n kind -q

echo "Creating KIND cluster 'kind' with builtin registry."
${SCRIPT_ROOT}/hack/e2e/kind-with-registry.sh

echo "Building metrics-pump image"
docker build -t localhost:5001/write-metrics:dev -f ${SCRIPT_ROOT}/hack/e2e/Dockerfile.externalmetrics-writer ${SCRIPT_ROOT}/hack
echo "  pushing image to local registry"
docker push localhost:5001/write-metrics:dev

case ${SUITE} in
  recommender|recommender-externalmetrics)
    ${SCRIPT_ROOT}/hack/vpa-down.sh
    echo " ** Deploying for suite ${SUITE}"
    ${SCRIPT_ROOT}/hack/deploy-for-e2e-locally.sh ${SUITE}

    echo " ** Running suite ${SUITE}"
    if [ ${SUITE} == recommender-externalmetrics ]; then
       ${SCRIPT_ROOT}/hack/run-e2e-tests.sh recommender
    else
      ${SCRIPT_ROOT}/hack/run-e2e-tests.sh ${SUITE}
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac

