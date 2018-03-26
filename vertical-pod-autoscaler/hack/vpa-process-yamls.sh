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
  echo "<action> should be either 'create' or 'delete'."
  echo "<component> might be on of 'admission-controller', 'updater', 'recommender'."
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

YAMLS="api/vpa-crd.yaml deploy/vpa-rbac.yaml deploy/updater-deployment.yaml deploy/recommender-deployment.yaml deploy/admission-controller-deployment.yaml"

if [ $# -gt 1 ]; then
  YAMLS="deploy/$2-deployment.yaml"
fi

for i in $YAMLS; do
  ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh ${SCRIPT_ROOT}/$i | kubectl $1 -f - || true
done

