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

pushd ${CONTRIB_ROOT}/cluster-autoscaler/
GO_VET_ANALYZER_NAME_FLAGS="-appends=false -asmdecl=false -assign=false -cgocall=false -composites=false -copylocks=false -defers=false -framepointer=false -hostport=false -httpresponse=false -loopclosure=false -lostcancel=false -shift=false -sigchanyzer=false -stdmethods=false -stdversion=false -structtag=false -testinggoroutine=false -timeformat=false -unmarshal=false -unreachable=false -unsafeptr=false -unusedresult=false -waitgroup=false"
go test -race -count=1 ./...
go vet "${GO_VET_ANALYZER_NAME_FLAGS}" "$(go list ./... | grep -v vendor | grep -v sdk-go | grep -v go-sdk)"
popd
