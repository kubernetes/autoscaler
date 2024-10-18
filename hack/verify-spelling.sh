#!/usr/bin/env bash
# Copyright 2018 The Kubernetes Authors.
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

DIR=$(dirname $0)

# Spell checking
git ls-files --full-name | grep -v -e vendor | grep -v cluster-autoscaler/cloudprovider/magnum/gophercloud| grep -v cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3 | grep -v cluster-autoscaler/cloudprovider/digitalocean/godo | grep -v cluster-autoscaler/cloudprovider/hetzner/hcloud-go | grep -v cluster-autoscaler/cloudprovider/bizflycloud/gobizfly | grep -v cluster-autoscaler/cloudprovider/oci/vendor-internal | grep -E -v 'cluster-autoscaler/cloudprovider/brightbox/(go-cache|gobrightbox|k8ssdk|linkheader)' | grep -v cluster-autoscaler/cloudprovider/aws/aws-sdk-go | grep -v cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go | xargs misspell -error -o stderr
