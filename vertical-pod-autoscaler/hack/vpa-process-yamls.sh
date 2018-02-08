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

REGISTRY_TO_APPLY=${REGISTRY-gcr.io/kubernetes-develop}
TAG_TO_APPLY=${TAG-0.0.1}

if [ "x$REGISTRY" != "xgcr.io/kubernetes-develop" ]; then
  echo "WARNING! Using image repository from REGISTRY env variable (${REGISTRY_TO_APPLY}) instead of gcr.io/kubernetes-develop."
fi


if [ "x$TAG" != "x0.0.1" ]; then
  echo "WARNING! Using tag from TAG env variable (${TAG_TO_APPLY}) instead of the default (0.0.1)."
fi

for i in $YAMLS; do
  sed -e "s,gcr.io/kubernetes-develop/\([a-z-]*\):.*,${REGISTRY_TO_APPLY}/\1:${TAG_TO_APPLY}," ${SCRIPT_ROOT}/$i | kubectl $1 -f -
done
