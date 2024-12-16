#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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
  echo " - full-mpa"
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

export WORKSPACE=${WORKSPACE:-/workspace/_artifacts}

case ${SUITE} in
  recommender|updater|admission-controller|actuation|full-mpa)
    export KUBECONFIG=$HOME/.kube/config
    pushd ${SCRIPT_ROOT}/e2e
    go test ./v1alpha1/*go -v --test.timeout=90m --args --ginkgo.v=true --ginkgo.focus="\[MPA\] \[${SUITE}\]" --report-dir=${WORKSPACE} --disable-log-dump --ginkgo.timeout=90m
    V1ALPHA1_RESULT=$?
    popd
    echo v1alpha1 test result: ${V1ALPHA1_RESULT}
    if [ $V1ALPHA1_RESULT -gt 0 ]; then
      echo "Please check v1alpha1 \"go test\" logs!"
    fi
    if [ $V1ALPHA1_RESULT -gt 0 ]; then
      echo "Tests failed"
      exit 1
    fi
    ;;
  *)
    print_help
    exit 1
    ;;
esac
