#!/bin/bash

# Copyright 2025 The Kubernetes Authors.
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

SCRIPT_ROOT=$(realpath $(dirname "${BASH_SOURCE[0]}"))/..
TARGET_FILE="${SCRIPT_ROOT}/docs/vpa-flags.md"
COMPONENTS=("admission-controller" "recommender" "updater")

# Function to extract flags from a binary
extract_flags() {
    local binary=$1
    local component=$2
    
    if [ ! -f "$binary" ]; then
        echo "Error: Binary not found for ${component} at ${binary}"
        return 1
    fi
    
    echo "# What are the parameters to VPA ${component}?"
    echo "This document is auto-generated from the flag definitions in the VPA ${component} code."
    echo "Last updated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    echo
    echo "| Flag | Type | Default | Description |"
    echo "|------|------|---------|-------------|"

    $binary --help 2>&1 | grep -E '^\s*-' | while read -r line; do
        flag=$(echo "$line" | awk '{print $1}' | sed 's/^-*//;s/=.*$//')
        type=$(echo "$line" | awk '{print $2}' | sed 's/[()]//g')
        default=$(echo "$line" | sed -n 's/.*default \([^)]*\).*/\1/p')
        description=$(echo "$line" | sed -E 's/^\s*-[^[:space:]]+ [^[:space:]]+ //;s/ \(default.*\)//')
        
        echo "| --${flag} | ${type} | ${default} | ${description} |"
    done
    echo
}

# Build components
pushd "${SCRIPT_ROOT}" >/dev/null
for component in "${COMPONENTS[@]}"; do
    echo "Building ${component}..."
    pushd "pkg/${component}" >/dev/null
    if ! go build -o ${component} ; then
        echo "Error: Failed to build ${component}"
        popd >/dev/null
        continue
    fi
    popd >/dev/null
done
popd >/dev/null

# Generate combined flags documentation
echo "Generating flags documentation..."
{
    echo "# Vertical Pod Autoscaler Flags"
    echo "This document contains the flags for all VPA components."
    echo

    for component in "${COMPONENTS[@]}"; do
        binary="${SCRIPT_ROOT}/pkg/${component}/${component}"
        if ! extract_flags "$binary" "$component" ; then
            echo "Error: Failed to extract flags for ${component}"
        fi
    done
} > "${TARGET_FILE}"

echo "VPA flags documentation has been generated in ${TARGET_FILE}"
