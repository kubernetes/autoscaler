#!/usr/bin/env bash
set -o pipefail
set -o nounset
set -o errexit

UPSTREAM_REPO=${UPSTREAM_REPO:-https://github.com/hetznercloud/hcloud-go.git}
UPSTREAM_REF=${UPSTREAM_REF:-main}

vendor_path=hcloud-go

original_module_path="github.com/hetznercloud/hcloud-go/v2/"
vendor_module_path=k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/


echo "# Removing existing directory."
rm -rf hcloud-go

echo "# Cloning repo"
git clone --depth=1 --branch "$UPSTREAM_REF" "$UPSTREAM_REPO" "$vendor_path"

echo "# Removing unnecessary files"
find "$vendor_path" -type f ! -name "*.go" ! -name "LICENSE" -delete
find "$vendor_path" -type f -name "*_test.go" -delete
find "$vendor_path" -type d -empty -delete

echo "# Rewriting module path"
find "$vendor_path" -type f -exec sed -i "s@${original_module_path}@${vendor_module_path}@g" {} +

