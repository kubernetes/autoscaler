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

export REGISTRY=gcr.io/`gcloud config get-value core/project`
export TAG=latest

COMPONENTS="recommender updater admission-controller"

if [ $# -ne 0 ]; then
  COMPONENTS=$*
fi

for i in ${COMPONENTS}; do
  make --directory ${SCRIPT_ROOT}/${i} release
done

kubectl create -f ${SCRIPT_ROOT}/api/vpa-crd.yaml
kubectl create -f ${SCRIPT_ROOT}/deploy/vpa-rbac.yaml

for i in ${COMPONENTS}; do
  ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh  ${SCRIPT_ROOT}/deploy/${i}-deployment.yaml | kubectl create -f -
done
