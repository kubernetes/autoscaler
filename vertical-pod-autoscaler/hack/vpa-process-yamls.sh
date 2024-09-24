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
COMPONENTS="vpa-v1-crd-gen vpa-rbac updater-deployment recommender-deployment admission-controller-deployment"

function script_path {
  # Regular components have deployment yaml files under /deploy/.  But some components only have
  # test deployment yaml files that are under hack/e2e. Check the main deploy directory before
  # using the e2e subdirectory.
  if test -f "${SCRIPT_ROOT}/deploy/${1}.yaml"; then
    echo "${SCRIPT_ROOT}/deploy/${1}.yaml"
  else
    echo "${SCRIPT_ROOT}/hack/e2e/${1}.yaml"
  fi
}

case ${ACTION} in
delete|diff) COMPONENTS+=" vpa-beta2-crd" ;;
esac

if [ $# -gt 1 ]; then
  COMPONENTS="$2-deployment"
fi

for i in $COMPONENTS; do
  if [ $i == admission-controller-deployment ] ; then
    if [[ ${ACTION} == create || ${ACTION} == update ]] ; then
      # Allow gencerts to fail silently if certs already exist
      (bash ${SCRIPT_ROOT}/pkg/admission-controller/gencerts.sh || true)
    elif [ ${ACTION} == delete ] ; then
      (bash ${SCRIPT_ROOT}/pkg/admission-controller/rmcerts.sh || true)
      (bash ${SCRIPT_ROOT}/pkg/admission-controller/delete-webhook.sh || true)
    fi
  fi
  if [[ ${ACTION} == print ]]; then
    ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh $(script_path $i)
  else
    ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh $(script_path $i) | kubectl ${ACTION} -f - || true
  fi
done
