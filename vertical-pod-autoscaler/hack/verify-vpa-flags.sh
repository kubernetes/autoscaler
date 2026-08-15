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
FLAG_FILE="${SCRIPT_ROOT}/docs/flags.md"
GENERATE_FLAGS_SCRIPT="${SCRIPT_ROOT}/hack/generate-flags.sh"

existing_flags=$(<"$FLAG_FILE")
"$GENERATE_FLAGS_SCRIPT"
updated_flags=$(<"$FLAG_FILE")

if [[ "$existing_flags" != "$updated_flags" ]]; then
    echo "VPA flags are not up to date. Please run ${GENERATE_FLAGS_SCRIPT}"
    exit 1
fi
