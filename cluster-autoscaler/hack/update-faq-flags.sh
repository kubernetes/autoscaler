#!/usr/bin/env bash

# Copyright 2021 The Kubernetes Authors.
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

###
# This script get flags from cluster-autoscaler and generate new flags table
###

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(realpath $(dirname "${BASH_SOURCE[0]}"))/..
AUTOSCALER="${SCRIPT_ROOT}/cluster-autoscaler-$(go env GOARCH)"
TARGET_FILE="${SCRIPT_ROOT}/FAQ.md"

pushd "${SCRIPT_ROOT}"
[ -e "${AUTOSCALER}" ] || make build
popd

# Get flags from `cluster-autoscaler -h` stderr output
set +e
HELP_OUTPUT=$($AUTOSCALER --help 2>&1 | grep -Ev '(^$|^Usage|^pflag|--ginkgo)')
set -e
FLAGS=$(echo "${HELP_OUTPUT}" | awk '
    /^[[:space:]]*--/ { print; next }
    /^[[:space:]]*-/ { $1=""; print }  # remove the shorthand
')

TABLE_HEADER="| Parameter | Description | Default |
| --- | --- | --- |"
ARGS=("${TABLE_HEADER}")

# Generate new flag makrdown table
while read -r line; do
  param=$(echo "$line" | awk '{print $1}' | cut -c3-)
  desc=$(echo "$line" | cut -d' ' -f3- | sed -E 's/\(default .+\)//' | awk '{$1=$1; print}' )
  default=$(echo "$line" | grep -oP '\(default \K[^)]+' || echo "")
  ARGS+=("| \`$param\` | $desc | $default |")
done <<< "${FLAGS}"

ARGS+=("")
TABLE=$(printf "%s\n" "${ARGS[@]}")

# Search the flag table
TITLE="| Parameter | Description | Default |"
START_LINE=$(grep -n "${TITLE}" "${TARGET_FILE}" | cut -d: -f1)
# next empty line
END_LINE=$(awk -v start="${START_LINE}" 'NR > start && /^[[:space:]]*$/{print NR; exit}' "${TARGET_FILE}")
((END_LINE--))

# Replace the table with the generated one
TEMP=$(mktemp)
awk -v start="${START_LINE}" -v end="${END_LINE}" -v replacement="${TABLE}" '
  NR == start {print replacement; next}
  NR > start && NR <= end {next}
  {print}
' "${TARGET_FILE}" > "${TEMP}"
mv "${TEMP}" "${TARGET_FILE}"
