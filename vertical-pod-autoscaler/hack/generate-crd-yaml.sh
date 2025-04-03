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
WORKSPACE=$(mktemp -d)

function cleanup() {
    rm -r ${WORKSPACE}
}
trap cleanup EXIT

if [[ -z $(which controller-gen) ]]; then
    (
        cd $WORKSPACE
	      go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.5
    )
    CONTROLLER_GEN=${GOBIN:-$(go env GOPATH)/bin}/controller-gen
else
    CONTROLLER_GEN=$(which controller-gen)
fi

# The following commands always returns an error because controller-gen does not accept keys other than strings.
${CONTROLLER_GEN} ${CRD_OPTS} paths="${APIS_PATH}/..." output:crd:dir="\"${WORKSPACE}\"" >& ${WORKSPACE}/errors.log ||:
grep -v -e 'map keys must be strings, not int' -e 'not all generators ran successfully' -e 'usage' ${WORKSPACE}/errors.log \
    && { echo "Failed to generate CRD YAMLs."; exit 1; }

cat "${WORKSPACE}/autoscaling.k8s.io_verticalpodautoscalercheckpoints.yaml" > ${OUTPUT}
cat "${WORKSPACE}/autoscaling.k8s.io_verticalpodautoscalers.yaml" >> ${OUTPUT}
