#!/bin/bash

# Copyright The Kubernetes Authors.
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

go install github.com/elastic/crd-ref-docs@v0.3.0
export CONTROLLER_GEN=${GOBIN:-$(go env GOPATH)/bin}/crd-ref-docs

cd "$(dirname "${BASH_SOURCE[0]}")/.."
$CONTROLLER_GEN \
    --source-path=pkg/apis/ \
    --config=./hack/api-docs/config.yaml \
    --renderer=markdown \
    --output-path=${WORKSPACE}


ret=0

diff -Naupr ${WORKSPACE}/out.md ${OUTPUT} || ret=$?
if [ $ret -eq 0 ]
then
  echo "api-docs are up to date."
else
  echo "api-docs are out of date. Please run hack/generate-api-docs.sh"
  exit 1
fi
