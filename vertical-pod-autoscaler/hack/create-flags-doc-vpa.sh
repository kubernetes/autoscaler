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

function generate_vpa_docs_components {
  components="recommender updater admission-controller"
  for component in $components; do
    echo "generate docs for $component"
    cd ../pkg/$component
    make document-flags
    cd -
  done
}

function move_and_merge_docs {
  files="vpa-admission-flags.md vpa-recommender-flags.md vpa-updater-flags.md"
  cd ../docs
  rm -f flags.md
  touch flags.md
  for file in $files; do
    echo "merge $file"
    cat $file >> flags.md
    echo "" >> flags.md
  done
  cd -
}

generate_vpa_docs_components
move_and_merge_docs