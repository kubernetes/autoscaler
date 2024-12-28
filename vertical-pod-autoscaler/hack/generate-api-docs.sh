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


set -o errexit
set -o nounset
set -o pipefail

REPOSITORY_ROOT=$(realpath $(dirname ${BASH_SOURCE})/..)
OUTPUT=${REPOSITORY_ROOT}/docs/api.md
WORKSPACE=$(mktemp -d)

function cleanup() {
    rm -r ${WORKSPACE}
}
trap cleanup EXIT

if [[ -z $(which crd-ref-docs) ]]; then
    (
        cd $WORKSPACE
	      go install github.com/elastic/crd-ref-docs@latest
    )
    CONTROLLER_GEN=${GOBIN:-$(go env GOPATH)/bin}/crd-ref-docs
else
    CONTROLLER_GEN=$(which crd-ref-docs)
fi

cd "$(dirname "${BASH_SOURCE[0]}")/.."
crd-ref-docs \
    --source-path=pkg/apis/ \
    --config=./hack/api-docs/config.yaml \
    --renderer=markdown \
    --output-path=${WORKSPACE}

mv ${WORKSPACE}/out.md ${OUTPUT}
