#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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

# Get the absolute path to the script's directory
SCRIPT_DIR=$(realpath $(dirname ${BASH_SOURCE}))
# Get the repository root (parent of hack directory)
REPOSITORY_ROOT=$(realpath ${SCRIPT_DIR}/..)

function generate_vpa_docs_components {
  components="recommender updater admission-controller"
  for component in $components; do
    echo "generate docs for $component"
    # Use absolute paths based on REPOSITORY_ROOT
    (
      cd "${REPOSITORY_ROOT}/pkg/${component}"
      make document-flags
    )
  done
}

function move_and_merge_docs {
  files="vpa-admission-flags.md vpa-recommender-flags.md vpa-updater-flags.md"
  # Use absolute path for docs directory
  (
    cd "${REPOSITORY_ROOT}/docs"
    rm -f flags.md
    touch flags.md
    for file in $files; do
      echo "merge $file"
      cat "$file" >> flags.md
      echo "" >> flags.md
      # Remove the temporary file after merging
      rm -f "$file"
    done
  )
  echo "updated docs/flags.md"
}

generate_vpa_docs_components
move_and_merge_docs
