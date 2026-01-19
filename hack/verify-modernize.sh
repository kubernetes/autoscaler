#!/bin/bash

# Copyright 2026 The Kubernetes Authors.
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

CA_ROOT=$(dirname "${BASH_SOURCE}")/../cluster-autoscaler
cd "${CA_ROOT}"

GOLINT=${GOLINT:-"golint"}
excluded_packages=(
  'cloudprovider/alicloud/alibaba-cloud-sdk-go'
  'cloudprovider/baiducloud/baiducloud-sdk-go'
  'cloudprovider/civo/civo-cloud-sdk-go'
  'cloudprovider/ovhcloud/sdk'
  'cloudprovider/magnum/gophercloud'
  'cloudprovider/digitalocean/godo'
  'cloudprovider/bizflycloud/gobizfly'
  'cloudprovider/brightbox/gobrightbox'
  'cloudprovider/brightbox/k8ssdk'
  'cloudprovider/brightbox/go-cache'
  'cloudprovider/exoscale/internal'
  'cloudprovider/huaweicloud/huaweicloud-sdk-go-v3'
  'cloudprovider/ionoscloud/ionos-cloud-sdk-go'
  'cloudprovider/hetzner/hcloud-go'
  'cloudprovider/oci/vendor-internal'
  'cloudprovider/tencentcloud/tencentcloud-sdk-go'
  'cloudprovider/volcengine/volc-sdk-golang'
  'cloudprovider/volcengine/volcengine-go-sdk'
  'cloudprovider/externalgrpc'
  'expander/grpcplugin'
)

FIND_PACKAGES='go list ./... 2> /dev/null'
for package in "${excluded_packages[@]}"; do
  FIND_PACKAGES+="| grep -v ${package} "
done

PACKAGES=()
mapfile -t PACKAGES < <(eval ${FIND_PACKAGES})
bad_files=()
for package in ${PACKAGES[@]+"${PACKAGES[@]}"}; do
  out=$(go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test "${package}")
  if [[ -n "${out}" ]]; then
    bad_files+=("${out}")
  fi
done
if [[ "${#bad_files[@]}" -ne 0 ]]; then
  echo "!!! modernize problems: "
  echo "${bad_files[@]}"
  exit 1
fi

# ex: ts=2 sw=2 et filetype=sh
