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
DEFAULT_TAG="1.7.1"

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

    # Print the row for the flag collected so far, extracting the default value
    # from either the pflag default location or usage text "(default: value)"
    function print_flag() {
        sub(/[[:space:]]+$/, "", desc)
        default_val = ""
        while (match(desc, / \(default:? [^(]*\)$/)) {
            if (default_val == "") {
                default_val = substr(desc, RSTART, RLENGTH)
                sub(/^ \(default:? /, "", default_val)
                sub(/\)$/, "", default_val)
            }
            desc = substr(desc, 1, RSTART-1)
        }

        # Boolean default is omitted by pflag only when it is false
        if (type == "bool" && default_val == "") {
            default_val = "false"
        }
        gsub(/\|/, "\\|", default_val)
        gsub(/\|/, "\\|", desc)
        print "| `" flag "` | " type " | " default_val " | " desc " |"
    }

    /^[[:space:]]*-{1,2}[a-zA-Z0-9_.-]+/ {
        if (collecting) {
            print_flag()
        }
        collecting = 1
        line = $0
        sub(/^[[:space:]]*/, "", line)

        # The flag together with the type is padded with spaces to align the
        # usage text so the first run of two or more spaces separates them
        if (match(line, /  +/)) {
            names = substr(line, 1, RSTART-1)
            desc = substr(line, RSTART+RLENGTH)
        } else {
            names = line
            desc = ""
        }

        # Extract the flag name, dropping any shorthand (e.g. "-v, --v Level")
        sub(/^-[a-zA-Z0-9], /, "", names)
        sub(/^--*/, "", names)

        # Whatever follows the flag name is its type (omitted for boolean flags)
        if (match(names, /[[:space:]]/)) {
            flag = substr(names, 1, RSTART-1)
            type = substr(names, RSTART+1)
        } else {
            flag = names
            type = "bool"
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
            print_flag()
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
    echo "<!-- toc -->"
    for component in "${COMPONENTS[@]}"; do
        echo "- [What are the parameters to VPA $component?](#what-are-the-parameters-to-vpa-$component)"
    done
    echo "<!-- /toc -->"
    echo

    for component in "${COMPONENTS[@]}"; do
        binary="${SCRIPT_ROOT}/pkg/${component}/${component}"
        if ! extract_flags "$binary" "$component" ; then
            echo "Error: Failed to extract flags for ${component}"
        fi
    done
} > "${TARGET_FILE}"

echo "VPA flags documentation has been generated in ${TARGET_FILE}"
