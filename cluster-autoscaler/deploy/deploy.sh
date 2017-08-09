#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
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

# Usage:
#   REGISTRY=<my_reg> MIG_LINK=<my_mig> [MIN=1] [MAX=4] [VERSION=v0.6.0] deploy.sh

set -e

ROOT=$(dirname "${BASH_SOURCE}")/..

if [ -z "$REGISTRY" ]; then
	echo "Required variable REGISTRY unset"
	exit 1
fi

# TODO(filipg): Detect MIG link automaticaly.
if [ -z "$MIG_LINK" ]; then
	echo "Required variable MIG_LINK unset"
	exit 1
fi

echo "==Building and pushing image=="
make release

# Create configuration file based on template
tmp_config=$(mktemp ca-controller-XXXX.yaml)
cp ${ROOT}/deploy/ca-controller.yaml ${tmp_config}

sed -i "s|{{VERSION}}|${VERSION:-dev}|g" ${tmp_config}
sed -i "s|{{REGISTRY}}|${REGISTRY}|g" ${tmp_config}
sed -i "s|{{MIN}}|${MIN_NODES:-1}|g" ${tmp_config}
sed -i "s|{{MAX}}|${MAX_NODES:-4}|g" ${tmp_config}
sed -i "s|{{MIG_LINK}}|${MIG_LINK}|g" ${tmp_config}

echo ""
echo "==Start cluster autoscaler=="
kubectl create -f ${tmp_config}

rm ${tmp_config}
