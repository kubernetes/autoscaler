#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
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

set -e

which jq > /dev/null || { echo "jq is required" >&2; exit 1; }

VERSION=${1#"v"}
if [ -z "$VERSION" ]; then
    echo "Must specify version!"
    exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

curl -s -S -f "https://raw.githubusercontent.com/kubernetes/kubernetes/v${VERSION}/go.mod" > "$tmp/modules"
sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p' -i "$tmp/modules"

for MOD in $(cat "$tmp/modules"); do
  go mod download -json "${MOD}@kubernetes-${VERSION}" > "$tmp/mod-json"
  version="$(cat "$tmp/mod-json" | jq -r .Version)"
  echo "using ${MOD}@${version}"
  go mod edit -replace="${MOD}"="${MOD}@${version}"
done

#go get "k8s.io/kubernetes@v${VERSION}"
