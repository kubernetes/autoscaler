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
cd "${KUBE_ROOT}"

GOLINT=${GOLINT:-"golint"}
excluded_packages=(
  '/vendor/'
  'vertical-pod-autoscaler/pkg/client'
  'cluster-autoscaler/cloudprovider/magnum/gophercloud'
  'cluster-autoscaler/cloudprovider/digitalocean/godo'
  'cluster-autoscaler/cloudprovider/exoscale/internal'
  'cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3'
  'cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go'
)

FIND_PACKAGES='go list ./... '
for package in "${excluded_packages[@]}"; do
     FIND_PACKAGES+="| grep -v ${package} "
done

PACKAGES=()
mapfile -t PACKAGES < <(eval ${FIND_PACKAGES})
bad_files=()
for package in "${PACKAGES[@]}"; do
  out=$("${GOLINT}" -min_confidence=0.9 "${package}")
  if [[ -n "${out}" ]]; then
    bad_files+=("${out}")
  fi
done
if [[ "${#bad_files[@]}" -ne 0 ]]; then
  echo "!!! '$GOLINT' problems: "
  echo "${bad_files[@]}"
  exit 1
fi

# ex: ts=2 sw=2 et filetype=sh
