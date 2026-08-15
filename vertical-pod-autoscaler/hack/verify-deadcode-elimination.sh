#!/usr/bin/env bash

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

# This script verifies if the golang linker is eliminating dead code in
# various components we care about, such as kube-apiserver, kubelet and others
# Usage: `hack/verify-deadcode-elimination.sh`.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${SCRIPT_ROOT}"
COMPONENTS=("admission-controller" "recommender" "updater")
FAILED=false
FAILURE_COUNT=0

# Arrays to collect per-binary results for JUnit output
declare -a BINARY_SIZES=()
declare -a BINARY_OUTPUTS=()
declare -a BINARY_FAILED=()

# Install whydeadcode
go install github.com/aarzilli/whydeadcode@latest

# Prefer full path for running zeitgeist
WHYDEADCODE_BIN="$(which whydeadcode)"

for binary in "${COMPONENTS[@]}"; do
  echo "Processing ${binary} ..."
  pushd "pkg/${binary}"
  output=$(GOLDFLAGS=-dumpdep go build -ldflags=-dumpdep -o "${binary}" 2>&1 | grep "\->" | ${WHYDEADCODE_BIN} 2>&1)
  if [[ -n "$output" ]]; then
    echo "golang linker is not eliminating dead code in ${binary}, please check the trace output below:"
    echo "(NOTE: that there may be false positives, but the first trace should be a real issue)"
    echo "$output"
    FAILED=true
    FAILED_BINARIES+=("${binary}")
    BINARY_FAILED+=("true")
    BINARY_OUTPUTS+=("$output")
    FAILURE_COUNT=$((FAILURE_COUNT + 1))
  else
    BINARY_FAILED+=("false")
    BINARY_OUTPUTS+=("")
  fi

  # Build a stripped binary (-s) to measure the size we actually ship.
  # The dead code build above uses -dumpdep which produces a non-stripped
  # binary, so we rebuild here with -s to report a representative size.
  echo "Building stripped ${binary} binary and checking its size:"
  go build -ldflags=-s -o "${binary}"

  # Capture size in bytes (wc -c is portable across macOS and Linux)
  size_bytes=$(wc -c < "${binary}" | tr -d ' ')
  BINARY_SIZES+=("$size_bytes")

  echo "$size_bytes $binary"

  echo ""
  popd
done

# Generate JUnit XML report
if [[ -z "${ARTIFACTS:-}" ]]; then
  ARTIFACTS=$(mktemp -d)
fi
JUNIT_OUTPUT_DIR="${ARTIFACTS}"
JUNIT_OUTPUT_FILE="${JUNIT_OUTPUT_DIR}/junit_deadcode.xml"
mkdir -p "${JUNIT_OUTPUT_DIR}"

cat > "${JUNIT_OUTPUT_FILE}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="deadcode-elimination" tests="${#COMPONENTS[@]}" failures="${FAILURE_COUNT}">
EOF

for i in "${!COMPONENTS[@]}"; do
  binary="${COMPONENTS[$i]}"
  size="${BINARY_SIZES[$i]}"
  failed="${BINARY_FAILED[$i]}"

  cat >> "${JUNIT_OUTPUT_FILE}" <<EOF
  <testcase name="${binary}" classname="deadcode">
    <properties>
      <property name="size_bytes" value="${size}"/>
    </properties>
EOF

  if [[ "$failed" == "true" ]]; then
    # Escape XML special characters in output
    escaped_output=$(echo "${BINARY_OUTPUTS[$i]}" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g')
    cat >> "${JUNIT_OUTPUT_FILE}" <<EOF
    <failure message="dead code detected">${escaped_output}</failure>
EOF
  fi

  cat >> "${JUNIT_OUTPUT_FILE}" <<EOF
  </testcase>
EOF
done

cat >> "${JUNIT_OUTPUT_FILE}" <<EOF
</testsuite>
EOF

echo "JUnit report written to ${JUNIT_OUTPUT_FILE}"

if [[ "$FAILED" == "true" ]]; then
  echo "Dead code elimination check failed for the following binaries:"
  for failed_binary in "${FAILED_BINARIES[@]}"; do
    echo "  - ${failed_binary}"
  done
  exit 1
fi
