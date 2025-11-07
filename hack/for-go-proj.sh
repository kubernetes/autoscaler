#! /bin/bash

# Copyright 2014 The Kubernetes Authors.
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
set -o pipefail
set -o nounset

CONTRIB_ROOT="$(dirname ${BASH_SOURCE})/.."

if [[ $# -ne 1 ]]; then
  echo "missing subcommand: [cluster-autoscaler|vertical-pod-autoscaler|addon-resizer]"
  exit 1
fi

PROJECT="${1}"

case "${PROJECT}" in
  "cluster-autoscaler")
    pushd ${CONTRIB_ROOT}/cluster-autoscaler/
    # TODO: #8127 - Use default analyzers set by `go test` to include `printf` analyzer.
    # Default analyzers that go test runs according to https://github.com/golang/go/blob/52624e533fe52329da5ba6ebb9c37712048168e0/src/cmd/go/internal/test/test.go#L649
    # This doesn't include the `printf` analyzer until cluster-autoscaler libraries are updated.
    ANALYZERS="atomic,bool,buildtags,directive,errorsas,ifaceassert,nilfunc,slog,stringintconv,tests"
    go test -count=1 ./... -vet="${ANALYZERS}"
    popd
    ;;
  "vertical-pod-autoscaler")
    pushd ${CONTRIB_ROOT}/vertical-pod-autoscaler
    go test -count=1  -race $(go list ./... | grep -v /vendor/ | grep -v vertical-pod-autoscaler/e2e | grep -v cluster-autoscaler/apis)
    popd
    pushd ${CONTRIB_ROOT}/vertical-pod-autoscaler/e2e
    go test -run=None ./...
    popd
    ;;
  "addon-resizer")
    pushd ${CONTRIB_ROOT}/addon-resizer
    godep go test -race $(go list ./... | grep -v /vendor/ | grep -v vertical-pod-autoscaler/e2e)
    popd
    ;;
  *)
    echo "invalid subcommand: ${CMD}"
    exit 1
    ;;
esac