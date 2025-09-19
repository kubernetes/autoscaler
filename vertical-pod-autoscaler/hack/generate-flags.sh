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
TARGET_FILE="${SCRIPT_ROOT}/docs/flags.md"
COMPONENTS=("admission-controller" "recommender" "updater")
DEFAULT_TAG="1.5.0"

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
    echo
    echo "| Flag | Type | Default | Description |"
    echo "|------|------|---------|-------------|"

    # Parse help output from binary to extract flag details
    $binary --help 2>&1 | awk '
    BEGIN {
        collecting = 0
        flag = ""
        desc = ""
        default_val = ""
        type = ""
    }

    /^[[:space:]]*-{1,2}[a-zA-Z0-9_.-]+/ {
        if (collecting) {
            gsub(/\|/, "\\|", desc)
            print "| `" flag "` | " type " | " default_val " | " desc " |"
        }
        collecting = 1
        line = $0

        # Extract flag name
        sub(/^[[:space:]]*/, "", line)
        split(line, parts, " ")
        flag = parts[1]
        sub(/^--*/, "", flag)

        # Extract default value (if any)
        default_val = ""
        match(line, /\(default[=: ]*[^)]*\)/)
        if (RSTART > 0) {
            default_val = substr(line, RSTART+8, RLENGTH-9)
        }

        # Try to guess type manually from known keywords in the line
        type = ""
        if (line ~ /string[[:space:]]/) type = "string"
        else if (line ~ /int[[:space:]]/) type = "int"
        else if (line ~ /uint[[:space:]]/) type = "uint"
        else if (line ~ /float[[:space:]]/) type = "float"
        else if (line ~ /bool[[:space:]]/) type = "bool"
        else if (line ~ /mapStringBool[[:space:]]/) type = "mapStringBool"
        else if (line ~ /traceLocation[[:space:]]/) type = "traceLocation"
        else if (line ~ /severity[[:space:]]/) type = "severity"
        else if (line ~ /moduleSpec[[:space:]]/) type = "moduleSpec"

        # Clean up description from line
        desc = line
        sub(/^[[:space:]]*-{1,2}[a-zA-Z0-9_.-]+[[:space:]]*/, "", desc)
        sub(/\(default[=: ]*[^)]*\)/, "", desc)
        gsub(/^[[:space:]]+/, "", desc)
        if (type != "") {
            sub(type "[[:space:]]+", "", desc)
        }
        next
    }

    /^[[:space:]]+/ {
        if (collecting) {
            line = $0
            gsub(/^[[:space:]]+/, "", line)
            desc = desc "<br>" line
        }
    }

    END {
        if (collecting) {
            gsub(/\|/, "\\|", desc)
            print "| `" flag "` | " type " | " default_val " | " desc " |"
        }
    }'
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
    echo "To view the most recent _release_ of flags for all VPA components, consult the release tag [flags($DEFAULT_TAG)](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-$DEFAULT_TAG/vertical-pod-autoscaler/docs/flags.md) documentation."
    echo
    echo "> **Note:** This document is auto-generated from the default branch (master) of the VPA repository."
    echo

    for component in "${COMPONENTS[@]}"; do
        binary="${SCRIPT_ROOT}/pkg/${component}/${component}"
        if ! extract_flags "$binary" "$component" ; then
            echo "Error: Failed to extract flags for ${component}"
        fi
    done
} > "${TARGET_FILE}"

echo "VPA flags documentation has been generated in ${TARGET_FILE}"
