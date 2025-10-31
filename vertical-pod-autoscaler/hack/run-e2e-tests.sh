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
export GOBIN="${ABSOLUTE_PATH}/e2e/_output/bin"

export WORKSPACE=${WORKSPACE:-/workspace/_artifacts}

SKIP="--ginkgo.skip=\[Feature\:OffByDefault\]"

if [ "${TEST_WITH_FEATURE_GATES_ENABLED:-}" == "true" ]; then
  SKIP=""
fi

NUMPROC=${NUMPROC:-4}
if [ ${SUITE} == "actuation" ] ;
	NUMPROC=8
fi

case ${SUITE} in
  recommender|updater|admission-controller|actuation|full-vpa)
    export KUBECONFIG=$HOME/.kube/config
    pushd ${SCRIPT_ROOT}/e2e
    go install github.com/onsi/ginkgo/v2/ginkgo
    ${GOBIN}/ginkgo build v1/ && ${GOBIN}/ginkgo --nodes=$NUMPROC --focus="\[VPA\] \[${SUITE}\]" v1/v1.test -- --report-dir=${WORKSPACE} --disable-log-dump ${SKIP}
    V1_RESULT=$?
    popd
    echo v1 test result: ${V1_RESULT}
    if [ $V1_RESULT -gt 0 ]; then
      echo "Please check v1 \"go test\" logs!"
    fi
    if [ $V1_RESULT -gt 0 ]; then
      echo "Tests failed"
      exit 1
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac
