#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
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

REPOSITORY_ROOT=$(realpath $(dirname ${BASH_SOURCE})/..)
CRD_OPTS=crd:allowDangerousTypes=true
APIS_PATH=${REPOSITORY_ROOT}/pkg/apis
OUTPUT=${REPOSITORY_ROOT}/deploy/vpa-v1-crd-gen.yaml
CHARTS_CRD_DIR=${REPOSITORY_ROOT}/charts/vertical-pod-autoscaler/crds
CONTROLLER_GEN_VERSION=v0.16.5
WORKSPACE=$(mktemp -d)

function cleanup() {
    rm -r ${WORKSPACE}
}
trap cleanup EXIT

CONTROLLER_GEN=$(command -v controller-gen || true)
INSTALLED_VERSION=""
if [[ -n "${CONTROLLER_GEN}" ]]; then
    # "controller-gen --version" prints "Version: vX.Y.Z"
    INSTALLED_VERSION=$("${CONTROLLER_GEN}" --version 2>/dev/null | awk '{print $2}' || true)
fi

if [[ "${INSTALLED_VERSION}" != "${CONTROLLER_GEN_VERSION}" ]]; then
    if [[ -n "${CONTROLLER_GEN}" ]]; then
        echo "Found controller-gen ${INSTALLED_VERSION:-(unknown version)} at ${CONTROLLER_GEN}, but ${CONTROLLER_GEN_VERSION} is required."
    fi
    echo "Installing controller-gen ${CONTROLLER_GEN_VERSION} into a temporary directory."
    (
        cd $WORKSPACE
        GOBIN="${WORKSPACE}/bin" go install sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION}
    )
    CONTROLLER_GEN=${WORKSPACE}/bin/controller-gen
fi

# The following commands always returns an error because controller-gen does not accept keys other than strings.
${CONTROLLER_GEN} ${CRD_OPTS} paths="${APIS_PATH}/..." output:crd:dir="\"${WORKSPACE}\"" >& ${WORKSPACE}/errors.log ||:
grep -v -e 'map keys must be strings, not int' -e 'not all generators ran successfully' -e 'usage' ${WORKSPACE}/errors.log \
    && { echo "Failed to generate CRD YAMLs."; exit 1; }

cat "${WORKSPACE}/autoscaling.k8s.io_verticalpodautoscalercheckpoints.yaml" > ${OUTPUT}
cat "${WORKSPACE}/autoscaling.k8s.io_verticalpodautoscalers.yaml" >> ${OUTPUT}

# Copy the generated CRD to the charts directory
cp ${OUTPUT} ${CHARTS_CRD_DIR}/vpa-v1-crd-gen.yaml
echo "CRD copied to ${CHARTS_CRD_DIR}/vpa-v1-crd-gen.yaml"
