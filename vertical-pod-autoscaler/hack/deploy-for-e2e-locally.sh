#!/bin/bash

# Copyright 2023 The Kubernetes Authors.
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
BASE_NAME=$(basename $0)
source "${SCRIPT_ROOT}/hack/lib/util.sh"

ARCH=$(kube::util::host_arch)

function print_help {
  echo "ERROR! Usage: $BASE_NAME [suite]*"
  echo "<suite> should be one of:"
  echo " - recommender"
  echo " - recommender-externalmetrics"
  echo " - updater"
  echo " - admission-controller"
  echo " - actuation"
  echo " - full-vpa"
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
  recommender|recommender-externalmetrics|updater|admission-controller)
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

# Local KIND images
export REGISTRY=${REGISTRY:-localhost:5001}
export TAG=${TAG:-latest}

# Deploy metrics server for E2E tests
kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/k8s-metrics-server.yaml

# Build and load Docker images for each component
for i in ${COMPONENTS}; do
  COMPONENT_NAME=$i
  if [ $i == recommender-externalmetrics ] ; then
    COMPONENT_NAME=recommender
  fi
  ALL_ARCHITECTURES=${ARCH} make --directory ${SCRIPT_ROOT}/pkg/${COMPONENT_NAME} docker-build REGISTRY=${REGISTRY} TAG=${TAG}
  docker tag ${REGISTRY}/vpa-${COMPONENT_NAME}-${ARCH}:${TAG} ${REGISTRY}/vpa-${COMPONENT_NAME}:${TAG}
  kind load docker-image ${REGISTRY}/vpa-${COMPONENT_NAME}:${TAG}
done

# Prepare Helm values
HELM_CHART_PATH="${SCRIPT_ROOT}/charts/vertical-pod-autoscaler"
VALUES_FILE="${SCRIPT_ROOT}/hack/e2e/values-e2e-local.yaml"
HELM_RELEASE_NAME="vpa"
HELM_NAMESPACE="kube-system"

# Build dynamic Helm set arguments based on components
HELM_SET_ARGS=""

# Enable components based on suite
for i in ${COMPONENTS}; do
  case $i in
    recommender|recommender-externalmetrics)
      HELM_SET_ARGS="${HELM_SET_ARGS} --set recommender.enabled=true"
      HELM_SET_ARGS="${HELM_SET_ARGS} --set recommender.image.repository=${REGISTRY}/vpa-recommender"
      HELM_SET_ARGS="${HELM_SET_ARGS} --set recommender.image.tag=${TAG}"
      ;;
    updater)
      HELM_SET_ARGS="${HELM_SET_ARGS} --set updater.enabled=true"
      HELM_SET_ARGS="${HELM_SET_ARGS} --set updater.image.repository=${REGISTRY}/vpa-updater"
      HELM_SET_ARGS="${HELM_SET_ARGS} --set updater.image.tag=${TAG}"
      ;;
    admission-controller)
      HELM_SET_ARGS="${HELM_SET_ARGS} --set admissionController.enabled=true"
      HELM_SET_ARGS="${HELM_SET_ARGS} --set admissionController.image.repository=${REGISTRY}/vpa-admission-controller"
      HELM_SET_ARGS="${HELM_SET_ARGS} --set admissionController.image.tag=${TAG}"
      ;;
  esac
done

# Add feature gates if specified
if [ -n "${FEATURE_GATES:-}" ]; then
  # Add feature gates to each enabled component
  for i in ${COMPONENTS}; do
    case $i in
      recommender|recommender-externalmetrics)
        HELM_SET_ARGS="${HELM_SET_ARGS} --set recommender.extraArgs[0]=--feature-gates=${FEATURE_GATES}"
        ;;
      updater)
        HELM_SET_ARGS="${HELM_SET_ARGS} --set updater.extraArgs[0]=--feature-gates=${FEATURE_GATES}"
        ;;
      admission-controller)
        HELM_SET_ARGS="${HELM_SET_ARGS} --set admissionController.extraArgs[0]=--feature-gates=${FEATURE_GATES}"
        ;;
    esac
  done
fi

# Uninstall any existing VPA Helm release
helm uninstall ${HELM_RELEASE_NAME} --namespace ${HELM_NAMESPACE} 2>/dev/null || true

# Install VPA using Helm chart
echo " ** Installing VPA via Helm chart"
helm install ${HELM_RELEASE_NAME} ${HELM_CHART_PATH} \
  --namespace ${HELM_NAMESPACE} \
  --values ${VALUES_FILE} \
  ${HELM_SET_ARGS} \
  --wait

# Handle external metrics special case
if [[ "${SUITE}" == "recommender-externalmetrics" ]]; then
  echo " ** Setting up external metrics infrastructure"

  # Apply extra RBAC for external metrics
  rm -f ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
  patch -c ${SCRIPT_ROOT}/deploy/vpa-rbac.yaml -i ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.diff -o ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
  kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml

  # Deploy Prometheus and adapter for external metrics
  kubectl delete namespace monitoring --ignore-not-found=true
  kubectl create namespace monitoring
  kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/prometheus.yaml
  kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/prometheus-adapter.yaml
  kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/metrics-pump.yaml

  # Upgrade Helm release with external metrics configuration
  echo " ** Updating recommender with external metrics args"
  helm upgrade ${HELM_RELEASE_NAME} ${HELM_CHART_PATH} \
    --namespace ${HELM_NAMESPACE} \
    --values ${VALUES_FILE} \
    ${HELM_SET_ARGS} \
    --set "recommender.extraArgs[0]=--use-external-metrics=true" \
    --set "recommender.extraArgs[1]=--external-metrics-cpu-metric=cpu" \
    --set "recommender.extraArgs[2]=--external-metrics-memory-metric=mem" \
    --wait
fi
