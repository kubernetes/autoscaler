#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

function print_help {
  echo "ERROR! Usage: deploy-for-e2e.sh [suite]*"
  echo "<suite> should be one of:"
  echo " - recommender"
  echo " - updater"
  echo " - admission-controller"
  echo " - actuation"
  echo " - full-vpa"
  echo "If component is not specified all above will be started."
}

if [ $# -eq 0 ]; then
  print_help
  exit 1
fi

if [ $# -gt 1 ]; then
  print_help
  exit 1
fi

SUITE=$1

case ${SUITE} in
  recommender|updater|admission-controller)
    COMPONENTS="${SUITE}"
    ;;
  full-vpa)
    COMPONENTS="recommender updater admission-controller"
    ;;
  actuation)
    COMPONENTS="updater admission-controller"
    ;;
  *)
    print_help
    exit 1
    ;;
esac

# CI-specific registry logic
export REGISTRY=gcr.io/`gcloud config get-value core/project`
export TAG=latest

echo "Configuring registry authentication"
mkdir -p "${HOME}/.docker"
gcloud auth configure-docker -q

for i in ${COMPONENTS}; do
  if [ $i == admission-controller ] ; then
    (cd ${SCRIPT_ROOT}/pkg/${i} && bash ./gencerts.sh e2e || true)
  fi
  ALL_ARCHITECTURES=amd64 make --directory ${SCRIPT_ROOT}/pkg/${i} release
done

echo "Deploying VPA components via Helm for E2E CI..."

HELM_CHART_PATH="${SCRIPT_ROOT}/charts/vertical-pod-autoscaler"
VALUES_FILE="${SCRIPT_ROOT}/hack/e2e/values-e2e.yaml"
HELM_RELEASE_NAME="vpa"
HELM_NAMESPACE="kube-system"

# Build dynamic Helm set arguments based on components to isolate test suites
HELM_SET_ARGS=()
HELM_SET_ARGS+=("--set" "recommender.enabled=false")
HELM_SET_ARGS+=("--set" "updater.enabled=false")
HELM_SET_ARGS+=("--set" "admissionController.enabled=false")

for COMPONENT in ${COMPONENTS}; do
  case ${COMPONENT} in
    recommender)
      HELM_SET_ARGS+=("--set" "recommender.enabled=true")
      HELM_SET_ARGS+=("--set" "recommender.image.repository=${REGISTRY}/vpa-recommender")
      HELM_SET_ARGS+=("--set" "recommender.image.tag=${TAG}")
      ;;
    updater)
      HELM_SET_ARGS+=("--set" "updater.enabled=true")
      HELM_SET_ARGS+=("--set" "updater.image.repository=${REGISTRY}/vpa-updater")
      HELM_SET_ARGS+=("--set" "updater.image.tag=${TAG}")
      ;;
    admission-controller)
      HELM_SET_ARGS+=("--set" "admissionController.enabled=true")
      HELM_SET_ARGS+=("--set" "admissionController.image.repository=${REGISTRY}/vpa-admission-controller")
      HELM_SET_ARGS+=("--set" "admissionController.image.tag=${TAG}")
      ;;
  esac
done

# Helm does not delete or upgrade CRDs under crds/, so remove them here
kubectl delete crd verticalpodautoscalers.autoscaling.k8s.io --ignore-not-found=true
kubectl delete crd verticalpodautoscalercheckpoints.autoscaling.k8s.io --ignore-not-found=true

helm upgrade --install ${HELM_RELEASE_NAME} "${HELM_CHART_PATH}" \
  --namespace ${HELM_NAMESPACE} \
  --values "${VALUES_FILE}" \
  "${HELM_SET_ARGS[@]}" \
  --wait