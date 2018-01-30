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

YAMLS="api/vpa-crd.yaml deploy/vpa-rbac.yaml deploy/admission-controller-rbac.yaml deploy/updater-deployment.yaml deploy/recommender-deployment.yaml deploy/admission-controller-deployment.yaml"

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
