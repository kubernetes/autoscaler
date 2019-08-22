#!/bin/bash

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
