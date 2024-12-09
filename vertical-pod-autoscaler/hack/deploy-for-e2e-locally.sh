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

rm -f ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
patch -c ${SCRIPT_ROOT}/deploy/vpa-rbac.yaml -i ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.diff -o ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/vpa-rbac.yaml
# Other-versioned CRDs are irrelevant as we're running a modern-ish cluster.
kubectl apply -f ${SCRIPT_ROOT}/deploy/vpa-v1-crd-gen.yaml
kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/k8s-metrics-server.yaml

for i in ${COMPONENTS}; do
  if [ $i == recommender-externalmetrics ] ; then
    i=recommender
  fi
  if [ $i == admission-controller ] ; then
    (cd ${SCRIPT_ROOT}/pkg/${i} && bash ./gencerts.sh e2e || true)
  fi
  ALL_ARCHITECTURES=${ARCH} make --directory ${SCRIPT_ROOT}/pkg/${i} docker-build REGISTRY=${REGISTRY} TAG=${TAG}
  docker tag ${REGISTRY}/vpa-${i}-${ARCH}:${TAG} ${REGISTRY}/vpa-${i}:${TAG}
  kind load docker-image ${REGISTRY}/vpa-${i}:${TAG}
done


for i in ${COMPONENTS}; do
  if [ $i == recommender-externalmetrics ] ; then
     kubectl delete namespace monitoring --ignore-not-found=true
     kubectl create namespace monitoring
     kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/prometheus.yaml
     kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/prometheus-adapter.yaml
     kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/metrics-pump.yaml
     kubectl apply -f ${SCRIPT_ROOT}/hack/e2e/${i}-deployment.yaml
  else
     REGISTRY=${REGISTRY} TAG=${TAG} ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh  ${SCRIPT_ROOT}/deploy/${i}-deployment.yaml | kubectl apply -f -
  fi
done
