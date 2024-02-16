#!/usr/bin/env bash

# Copyright 2021 The Kubernetes Authors.
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

###
# This script is to be used when updating the generated clients of 
# the Provisioning Request CRD.
###

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(realpath $(dirname "${BASH_SOURCE[0]}"))/..
CODEGEN_PKG="../vendor/k8s.io/code-generator"
pushd "${SCRIPT_ROOT}/apis"

chmod +x "${CODEGEN_PKG}"/generate-groups.sh
chmod +x "${CODEGEN_PKG}"/generate-internal-groups.sh
 
bash "${CODEGEN_PKG}"/generate-groups.sh "applyconfiguration,client,deepcopy,informer,lister" \
  k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client \
  k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest \
  autoscaling.x-k8s.io:v1beta1 \
  --go-header-file "${SCRIPT_ROOT}"/../hack/boilerplate/boilerplate.generatego.txt

chmod -x "${CODEGEN_PKG}"/generate-groups.sh
chmod -x "${CODEGEN_PKG}"/generate-internal-groups.sh
popd
