#!/bin/bash

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

function print_help {
  echo "ERROR! Usage: vpa-process-yaml.sh <YAML files>+"
  echo "Script will output content of YAML files separated with YAML document"
  echo "separator and substituting REGISTRY and TAG for pod images"
}

# Requires input from stdin, otherwise hangs. If the input is a Deployment manifest,
# apply kubectl patch to add feature gates specified in the FEATURE_GATES environment variable.
# e.g. cat file.yaml | apply_feature_gate
function apply_feature_gate() {
  local input=""
  while IFS= read -r line; do
      input+="$line"$'\n'
  done

  # matching precisely "kind: Deployment" to avoid matching "kind: DeploymentConfig" or a line with extra whitespace
  if echo "$input" | grep -qE '^kind: Deployment$'; then
    if [ -n "${FEATURE_GATES}" ]; then
      if ! echo "$input" | kubectl patch --type=json --local -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--feature-gates='"${FEATURE_GATES}"'"}]' -o yaml -f - 2>/dev/null; then
        # If it fails, there was no args field, so we need to add it
        echo "$input" | kubectl patch --type=json --local -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args", "value": ["--feature-gates='"${FEATURE_GATES}"'"]}]' -o yaml -f -
      fi
    else
      echo "$input"
    fi
  else
    echo "$input"
  fi
}

if [ $# -eq 0 ]; then
  print_help
  exit 1
fi

DEFAULT_REGISTRY="registry.k8s.io/autoscaling"
DEFAULT_TAG="1.5.0"

REGISTRY_TO_APPLY=${REGISTRY-$DEFAULT_REGISTRY}
TAG_TO_APPLY=${TAG-$DEFAULT_TAG}
FEATURE_GATES=${FEATURE_GATES:-""}

if [ "${REGISTRY_TO_APPLY}" != "${DEFAULT_REGISTRY}" ]; then
  (>&2 echo "WARNING! Using image repository from REGISTRY env variable (${REGISTRY_TO_APPLY}) instead of ${DEFAULT_REGISTRY}.")
fi

if [ "${TAG_TO_APPLY}" != "${DEFAULT_TAG}" ]; then
  (>&2 echo "WARNING! Using tag from TAG env variable (${TAG_TO_APPLY}) instead of the default (${DEFAULT_TAG}).")
fi

for i in $*; do
  sed -e "s,${DEFAULT_REGISTRY}/\([a-z-]*\):.*,${REGISTRY_TO_APPLY}/\1:${TAG_TO_APPLY}," $i | apply_feature_gate
  echo "---"
done
