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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

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

if [ $# -eq 0 ] || [ $# -gt 1 ]; then
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

# Diagnostic dump on any failure. run-e2e.sh does NOT use errexit, so a silent
# failure here would let ginkgo run against an empty cluster and fail with
# misleading 404 errors. Loud structured output makes the real cause visible
# in the prow build log.
function dump_diagnostics() {
  local rc=$?
  if [[ ${rc} -ne 0 ]]; then
    echo "============================================================"
    echo "deploy-for-e2e.sh FAILED (exit ${rc}) - diagnostic dump"
    echo "============================================================"
    echo "--- Helm releases in ${HELM_NAMESPACE:-kube-system} ---"
    helm list --namespace "${HELM_NAMESPACE:-kube-system}" --all 2>&1 || true
    echo "--- Helm release status ---"
    helm status "${HELM_RELEASE_NAME:-vpa}" --namespace "${HELM_NAMESPACE:-kube-system}" 2>&1 || true
    echo "--- VPA-related CRDs ---"
    kubectl get crd 2>&1 | grep -E 'autoscaling\.k8s\.io' || echo "(no VPA CRDs found)"
    echo "--- Pods in ${HELM_NAMESPACE:-kube-system} ---"
    kubectl get pods --namespace "${HELM_NAMESPACE:-kube-system}" -o wide 2>&1 | grep -E 'NAME|vpa-' || true
    echo "--- Recent events in ${HELM_NAMESPACE:-kube-system} ---"
    kubectl get events --namespace "${HELM_NAMESPACE:-kube-system}" --sort-by='.lastTimestamp' 2>&1 | tail -30 || true
    echo "--- MutatingWebhookConfiguration vpa-webhook-config ---"
    kubectl get mutatingwebhookconfiguration vpa-webhook-config -o yaml 2>&1 | head -40 || echo "(not found)"
    echo "--- Pod logs (vpa-*) ---"
    for pod in $(kubectl get pods --namespace "${HELM_NAMESPACE:-kube-system}" -o name 2>/dev/null | grep -E 'vpa-' || true); do
      echo "--- ${pod} ---"
      kubectl logs --namespace "${HELM_NAMESPACE:-kube-system}" "${pod}" --all-containers --tail=80 2>&1 || true
    done
    echo "============================================================"
  fi
}
trap dump_diagnostics EXIT

# helm-settings.sh defines HELM_CHART_PATH, VALUES_FILE, HELM_RELEASE_NAME,
# HELM_NAMESPACE.
source "${SCRIPT_ROOT}/hack/e2e/helm-settings.sh"

# Sanity: chart and values must exist before we touch the cluster.
if [[ ! -d "${HELM_CHART_PATH}" ]]; then
  echo "ERROR: Helm chart not found at ${HELM_CHART_PATH}"
  exit 1
fi
if [[ ! -f "${VALUES_FILE}" ]]; then
  echo "ERROR: Values file not found at ${VALUES_FILE}"
  exit 1
fi
if [[ ! -f "${HELM_CHART_PATH}/crds/vpa-v1-crd-gen.yaml" ]]; then
  echo "ERROR: CRDs not found at ${HELM_CHART_PATH}/crds/"
  exit 1
fi

export REGISTRY="gcr.io/$(gcloud config get-value core/project)"
export TAG=latest

echo "===> Configuring registry authentication"
mkdir -p "${HOME}/.docker"
gcloud auth configure-docker -q

echo "===> Building and pushing component images: ${COMPONENTS}"
for COMPONENT in ${COMPONENTS}; do
  ALL_ARCHITECTURES=amd64 make --directory "${SCRIPT_ROOT}/pkg/${COMPONENT}" release
done

echo "===> Cleaning up any previous VPA installation in ${HELM_NAMESPACE}"
# Drop --wait on uninstall: not needed for fresh prow clusters and adds flakes.
helm uninstall "${HELM_RELEASE_NAME}" --namespace "${HELM_NAMESPACE}" 2>/dev/null || true
kubectl delete crd verticalpodautoscalers.autoscaling.k8s.io --ignore-not-found=true
kubectl delete crd verticalpodautoscalercheckpoints.autoscaling.k8s.io --ignore-not-found=true
kubectl delete mutatingwebhookconfiguration vpa-webhook-config --ignore-not-found=true

# Apply CRDs BEFORE helm install. Helm 3 only installs CRDs from crds/ on the
# very first install of a release; pre-applying with kubectl apply guarantees
# they exist regardless of helm release history. apply (not create) is
# idempotent.
echo "===> Applying CRDs from ${HELM_CHART_PATH}/crds/"
kubectl apply -f "${HELM_CHART_PATH}/crds/"

# Generate TLS certs AFTER cleanup but BEFORE helm install so the admission
# controller pod can mount the secret on first start. The official gencerts.sh
# uses kubectl create which fails if the secret already exists; explicitly
# delete first to make this idempotent on re-runs.
for COMPONENT in ${COMPONENTS}; do
  if [[ "${COMPONENT}" == "admission-controller" ]]; then
    echo "===> Generating TLS certificates for admission-controller"
    # gencerts.sh hard-codes --namespace=kube-system. HELM_NAMESPACE must match.
    if [[ "${HELM_NAMESPACE}" != "kube-system" ]]; then
      echo "ERROR: gencerts.sh hard-codes kube-system but HELM_NAMESPACE=${HELM_NAMESPACE}"
      exit 1
    fi
    kubectl delete secret -n "${HELM_NAMESPACE}" vpa-tls-certs --ignore-not-found=true
    kubectl delete secret -n "${HELM_NAMESPACE}" vpa-e2e-certs  --ignore-not-found=true
    (cd "${SCRIPT_ROOT}/pkg/${COMPONENT}" && bash ./gencerts.sh e2e)
    break
  fi
done

# Build per-suite Helm value overrides. values-e2e.yaml already disables every
# component by default; we enable only the ones this suite needs.
HELM_SET_ARGS=()
for COMPONENT in ${COMPONENTS}; do
  case ${COMPONENT} in
    recommender)
      HELM_SET_ARGS+=(
        "--set" "recommender.enabled=true"
        "--set" "recommender.image.repository=${REGISTRY}/vpa-recommender"
        "--set" "recommender.image.tag=${TAG}"
        "--set" "recommender.image.pullPolicy=Always"
      )
      ;;
    updater)
      HELM_SET_ARGS+=(
        "--set" "updater.enabled=true"
        "--set" "updater.image.repository=${REGISTRY}/vpa-updater"
        "--set" "updater.image.tag=${TAG}"
        "--set" "updater.image.pullPolicy=Always"
      )
      ;;
    admission-controller)
      HELM_SET_ARGS+=(
        "--set" "admissionController.enabled=true"
        "--set" "admissionController.image.repository=${REGISTRY}/vpa-admission-controller"
        "--set" "admissionController.image.tag=${TAG}"
        "--set" "admissionController.image.pullPolicy=Always"
      )
      ;;
  esac
done

echo "===> Installing VPA via Helm (suite: ${SUITE})"
# --atomic   : roll back on failure so we never leave a half-installed release
#              for the next prow job to trip over.
# --timeout  : explicit > default 5m, in case GCR pulls are slow on first
#              install of the day.
helm upgrade --install "${HELM_RELEASE_NAME}" "${HELM_CHART_PATH}" \
  --namespace "${HELM_NAMESPACE}" \
  --values "${VALUES_FILE}" \
  "${HELM_SET_ARGS[@]}" \
  --atomic \
  --wait \
  --timeout 10m

# Post-install verification. We want to fail fast HERE if something required
# by the e2e tests is missing, rather than have ginkgo report misleading 404s.
echo "===> Verifying VPA install"
kubectl get crd verticalpodautoscalers.autoscaling.k8s.io >/dev/null
kubectl get crd verticalpodautoscalercheckpoints.autoscaling.k8s.io >/dev/null

for COMPONENT in ${COMPONENTS}; do
  case ${COMPONENT} in
    recommender|updater)
      kubectl rollout status -n "${HELM_NAMESPACE}" "deployment/vpa-${COMPONENT}" --timeout=5m
      ;;
    admission-controller)
      kubectl rollout status -n "${HELM_NAMESPACE}" "deployment/vpa-admission-controller" --timeout=5m
      # The admission controller registers vpa-webhook-config itself when
      # registerWebhook=true. Wait for it explicitly so we either succeed here
      # or fail with a clear message instead of a silent 180s timeout in
      # ginkgo's BeforeEach.
      echo "===> Waiting for admission controller to register vpa-webhook-config"
      for i in $(seq 1 30); do
        if kubectl get mutatingwebhookconfiguration vpa-webhook-config >/dev/null 2>&1; then
          echo "vpa-webhook-config registered"
          break
        fi
        if [[ ${i} -eq 30 ]]; then
          echo "ERROR: vpa-webhook-config was not registered within 5 minutes"
          exit 1
        fi
        sleep 10
      done
      ;;
  esac
done

echo "===> Deploy complete for suite '${SUITE}'"
