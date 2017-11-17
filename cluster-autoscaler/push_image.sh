#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
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

IMAGE_TO_PUSH=$1
if [ -z $IMAGE_TO_PUSH ]; then
  echo No image passed
  exit 1
fi

echo "About to push image $IMAGE_TO_PUSH"
read -r -p "Are you sure? [y/N] " response
if [[ "$response" =~ ^([yY])+$ ]]; then
  gcloud docker -- pull $IMAGE_TO_PUSH
  if [ $? -eq 0 ]; then
    echo $IMAGE_TO_PUSH already exists
    exit 1
  fi
  gcloud docker -- push $IMAGE_TO_PUSH
else
  echo Aborted
  exit 1
fi