#!/bin/bash -e

# This script will be run bazel when building process starts to
# generate key-value information that represents the status of the
# workspace. The output should be like
#
# KEY1 VALUE1
# KEY2 VALUE2
#
# If the script exits with non-zero code, it's considered as a failure
# and the output will be discarded.

if [[ -z "${DOCKER_REGISTRY}" ]]; then
  DOCKER_REGISTRY="gcr.io"
fi

if [[ -z "${DOCKER_IMAGE_PREFIX}" ]]; then
  DOCKER_IMAGE_PREFIX=`gcloud config get-value project`/
fi

if [[ -z "${DOCKER_TAG}" ]]; then
  DOCKER_TAG="latest"
fi

echo "STABLE_DOCKER_REGISTRY ${DOCKER_REGISTRY}"
echo "STABLE_DOCKER_IMAGE_PREFIX ${DOCKER_IMAGE_PREFIX}"
echo "STABLE_DOCKER_TAG ${DOCKER_TAG}"

if [[ -z "${K8S_CLUSTER}" ]]; then
  K8S_CLUSTER=`kubectl config current-context`
fi

echo "STABLE_K8S_CLUSTER ${K8S_CLUSTER}"
