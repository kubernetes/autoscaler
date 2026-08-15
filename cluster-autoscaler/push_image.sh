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

# See: https://cloud.google.com/container-registry/docs/support/deprecation-notices#gcloud-docker
MAX_DOCKER_MAJOR_VERSION_FOR_GCLOUD=18
MAX_DOCKER_MINOR_VERSION_FOR_GCLOUD=03

IMAGE_TO_PUSH=$1
if [ -z $IMAGE_TO_PUSH ]; then
  echo No image passed
  exit 1
fi

docker_client_version=$(docker version -f "{{.Client.Version}}")
docker_client_major_version=$(echo "$docker_client_version" | cut -d'.' -f 1)
docker_client_minor_version=$(echo "$docker_client_version" | cut -d'.' -f 2)

docker_push_cmd=("docker")
if [[ "${IMAGE_TO_PUSH}" == "gcr.io/"* ]] || [[ "${IMAGE_TO_PUSH}" == "staging-k8s.gcr.io/"* ]] ; then
    if [[ "$docker_client_major_version" -lt "$MAX_DOCKER_MAJOR_VERSION_FOR_GCLOUD" ]] || [[ "$docker_client_major_version" -eq "$MAX_DOCKER_MAJOR_VERSION_FOR_GCLOUD" && "$docker_client_minor_version" -le "$MAX_DOCKER_MINOR_VERSION_FOR_GCLOUD" ]]; then
      echo "Docker version  <= $MAX_DOCKER_MAJOR_VERSION_FOR_GCLOUD.$MAX_DOCKER_MINOR_VERSION_FOR_GCLOUD, using gcloud command"
      docker_push_cmd=("gcloud" "docker" "--")
    else
      echo "Docker version > $MAX_DOCKER_MAJOR_VERSION_FOR_GCLOUD.$MAX_DOCKER_MINOR_VERSION_FOR_GCLOUD, ensure you have run gcloud auth configure-docker"
    fi
fi

echo "About to push image $IMAGE_TO_PUSH"
read -r -p "Are you sure? [y/N] " response
if [[ "$response" =~ ^([yY])+$ ]]; then
  "${docker_push_cmd[@]}" pull $IMAGE_TO_PUSH
  if [ $? -eq 0 ]; then
    echo $IMAGE_TO_PUSH already exists
    exit 1
  fi
  "${docker_push_cmd[@]}" push $IMAGE_TO_PUSH
else
  echo Aborted
  exit 1
fi
