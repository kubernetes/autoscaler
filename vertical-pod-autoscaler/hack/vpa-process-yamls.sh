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

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

function print_help {
  echo "ERROR! Usage: vpa-process-yamls.sh <action> [<component>]"
  echo "<action> should be either 'create', 'diff', 'print' or 'delete'."
  echo "The 'print' action will print all resources that would be used by, e.g., 'kubectl diff'."
  echo "<component> might be one of 'admission-controller', 'updater', 'recommender'."
  echo "If <component> is set, only the deployment of that component will be processed,"
  echo "otherwise all components and configs will be processed."
}

if [ $# -eq 0 ]; then
  print_help
  exit 1
fi

if [ $# -gt 2 ]; then
  print_help
  exit 1
fi

ACTION=$1
COMPONENTS="vpa-v1-crd vpa-rbac updater-deployment recommender-deployment admission-controller-deployment"
case ${ACTION} in
delete|diff|print) COMPONENTS+=" vpa-beta2-crd" ;;
esac

if [ $# -gt 1 ]; then
  COMPONENTS="$2-deployment"
fi

for i in $COMPONENTS; do
  if [ $i == admission-controller-deployment ] ; then
    if [ ${ACTION} == create ] ; then
      (bash ${SCRIPT_ROOT}/pkg/admission-controller/gencerts.sh || true)
    elif [ ${ACTION} == delete ] ; then
      (bash ${SCRIPT_ROOT}/pkg/admission-controller/rmcerts.sh || true)
      (bash ${SCRIPT_ROOT}/pkg/admission-controller/delete-webhook.sh || true)
    fi
  fi
  if [[ ${ACTION} == print ]]; then
    ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh ${SCRIPT_ROOT}/deploy/$i.yaml
  else
    ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh ${SCRIPT_ROOT}/deploy/$i.yaml | kubectl ${ACTION} -f - || true
  fi
done
