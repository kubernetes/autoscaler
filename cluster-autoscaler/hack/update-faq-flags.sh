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

function get_grep_command() {
    if grep -oP '\w' <<< "test" >/dev/null 2>&1; then
        echo "grep"
    elif command -v ggrep >/dev/null 2>&1; then
        echo "ggrep"
    elif command -v egrep >/dev/null 2>&1; then
        echo "egrep"
    else
        echo "Error: No suitable grep command found (requires GNU grep with -P support)" >&2
        exit 1
    fi
}

GREP_CMD="$(get_grep_command)"
SCRIPT_ROOT=$(realpath $(dirname "${BASH_SOURCE[0]}"))/..
AUTOSCALER="${SCRIPT_ROOT}/cluster-autoscaler-$(go env GOARCH)"
TARGET_FILE="${SCRIPT_ROOT}/FAQ.md"

pushd "${SCRIPT_ROOT}" >/dev/null
[ -e "${AUTOSCALER}" ] || make build
popd >/dev/null

# Get flags from `cluster-autoscaler -h` stderr output
set +e
HELP_OUTPUT=$($AUTOSCALER --help 2>&1 | $GREP_CMD -Ev '(^$|^Usage|^pflag|--ginkgo)')
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
  default=$(echo "$line" | $GREP_CMD -oP '\(default \K[^)]+' || echo "")
  ARGS+=("| \`$param\` | $desc | $default |")
done <<< "${FLAGS}"

ARGS+=("")
TABLE=$(printf "%s\n" "${ARGS[@]}")

# Search the flag table
TITLE_PATTERN="\|[[:space:]]*Parameter[[:space:]]*\|[[:space:]]*Description[[:space:]]*\|[[:space:]]*Default[[:space:]]*\|"
START_LINE=$($GREP_CMD -n -E "${TITLE_PATTERN}" "${TARGET_FILE}" | cut -d: -f1)
# next empty line
END_LINE=$(awk -v start="${START_LINE}" 'NR > start && /^[[:space:]]*$/{print NR; exit}' "${TARGET_FILE}")
((END_LINE--))

# Replace the table with the generated one
TEMP=$(mktemp)
TABLE_TEMP=$(mktemp)
echo "${TABLE}" > "${TABLE_TEMP}"

awk -v start="${START_LINE}" -v end="${END_LINE}" -v table_temp="${TABLE_TEMP}" '
  BEGIN {
    while ((getline l < table_temp) > 0) rep=rep l"\n";
    close(table_temp)
  }
  NR == start {print rep; next}
  NR > start && NR <= end {next}
  {print}
' "${TARGET_FILE}" > "${TEMP}"
rm -f "${TABLE_TEMP}"
mv "${TEMP}" "${TARGET_FILE}"

echo "FAQ.md has been automatically updated, please check for changes and submit"
