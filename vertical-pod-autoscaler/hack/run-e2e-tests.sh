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

case ${SUITE} in
  recommender|updater|admission-controller|actuation|full-vpa)
    export KUBECONFIG=$HOME/.kube/config
    go test ${SCRIPT_ROOT}/e2e/v1beta2/*go -v -test.timeout=60m  --args --ginkgo.v=true --ginkgo.focus="\[VPA\] \[${SUITE}\]" --report-dir=/workspace/_artifacts --disable-log-dump
    FIRST_RESULT=$?
    go test ${SCRIPT_ROOT}/e2e/v1/*go -v -test.timeout=60m  --args --ginkgo.v=true --ginkgo.focus="\[VPA\] \[${SUITE}\]" --report-dir=/workspace/_artifacts --disable-log-dump
    echo First test result: ${FIRST_RESULT}
    if [ $FIRST_RESULT -gt 0 ]; then
      echo "Please check first \"go test\" logs!"
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac
