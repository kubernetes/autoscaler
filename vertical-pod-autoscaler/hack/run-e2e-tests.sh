#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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
  echo "ERROR! Usage: run-e2e-tests.sh <suite>"
  echo "<suite> should be one of:"
  echo " - recommender"
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

export GO111MODULE=on
# todo(adrianmoisey): Make the setting of GOBIN nicer
ABSOLUTE_PATH=$(realpath "${SCRIPT_ROOT}")
export GOBIN="${ABSOLUTE_PATH}/test/e2e/_output/bin"

export ARTIFACTS=${ARTIFACTS:-/workspace/_artifacts}

SKIP="--ginkgo.skip=\[Feature\:OffByDefault\]"

if [ "${TEST_WITH_FEATURE_GATES_ENABLED:-}" == "true" ]; then
  SKIP=""
fi

NUMPROC=${NUMPROC:-10}

# Dump logs of all VPA component pods into ARTIFACTS so CI uploads them.
function dump_vpa_logs {
  local ns="kube-system"
  mkdir -p "${ARTIFACTS}"
  local component pod name
  for component in admission-controller recommender updater; do
    for pod in $(kubectl get pods -n "${ns}" -l "app.kubernetes.io/component=${component}" -o name 2>/dev/null); do
      name="${pod#pod/}"
      echo "Dumping $component logs for pod $pod ..."
      kubectl logs -n "${ns}" "${name}" --request-timeout=5s --tail=-1 > "${ARTIFACTS}/${name}.log" 2>/dev/null || true
      kubectl logs -n "${ns}" "${name}" --request-timeout=5s --tail=-1 --previous > "${ARTIFACTS}/${name}-previous.log" 2>/dev/null \
        || rm -f "${ARTIFACTS}/${name}-previous.log"
    done
  done
}

case ${SUITE} in
  recommender|updater|admission-controller|actuation|full-vpa)
    export KUBECONFIG=$HOME/.kube/config
    pushd ${SCRIPT_ROOT}/test/e2e
    go install github.com/onsi/ginkgo/v2/ginkgo
    ${GOBIN}/ginkgo build v1/ && ${GOBIN}/ginkgo --nodes=$NUMPROC --focus="\[VPA\] \[${SUITE}\]" v1/v1.test -- --report-dir=${ARTIFACTS} --disable-log-dump ${SKIP}
    V1_RESULT=$?
    popd
    echo "Copying VPA logs to ${ARTIFACTS}"
    dump_vpa_logs
    echo v1 test result: ${V1_RESULT}
    if [ $V1_RESULT -gt 0 ]; then
      echo "Please check v1 \"go test\" logs!"
      echo "Tests failed"
      exit 1
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac
