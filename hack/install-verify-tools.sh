#!/bin/bash

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
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
GO_VERSION=($(go version))

# golint only works for golang 1.5+
if [[ -z $(echo "${GO_VERSION[2]}" | grep -E 'go1.1|go1.2|go1.3|go1.4') ]]; then
  go get -u github.com/golang/lint/golint
fi

go get -u github.com/tools/godep

# ex: ts=2 sw=2 et filetype=sh
