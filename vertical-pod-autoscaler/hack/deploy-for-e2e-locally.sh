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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
BASE_NAME=$(basename "$0")
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
  echo ""
  echo "Environment variables:"
  echo "  REGISTRY           - Container image registry (default: localhost:5001)"
  echo "  TAG                - Container image tag (default: latest)"
  echo "  FEATURE_GATES      - Comma-separated list of feature gates to enable"
  echo "  EXTRA_HELM_VALUES  - Path to an additional Helm values file to use during installation"
}

if [ $# -eq 0 ] || [ $# -gt 1 ]; then
  print_help
  exit 1
fi

SUITE=$1

case ${SUITE} in
  recommender-externalmetrics)
    COMPONENTS="recommender"
    ;;
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

export REGISTRY=${REGISTRY:-localhost:5001}
export TAG=${TAG:-latest}

# Diagnostic dump on any failure — same signal the CI script provides.
function dump_diagnostics() {
  local rc=$?
  if [[ ${rc} -ne 0 ]]; then
    echo "============================================================"
    echo "deploy-for-e2e-locally.sh FAILED (exit ${rc}) - diagnostic dump"
    echo "============================================================"
    helm list --namespace "${HELM_NAMESPACE:-kube-system}" --all 2>&1 || true
    helm status "${HELM_RELEASE_NAME:-vpa}" --namespace "${HELM_NAMESPACE:-kube-system}" 2>&1 || true
    kubectl get crd 2>&1 | grep -E 'autoscaling\.k8s\.io' || echo "(no VPA CRDs)"
    kubectl get pods --namespace "${HELM_NAMESPACE:-kube-system}" -o wide 2>&1 | grep -E 'NAME|vpa-' || true
    kubectl get events --namespace "${HELM_NAMESPACE:-kube-system}" --sort-by='.lastTimestamp' 2>&1 | tail -30 || true
    for pod in $(kubectl get pods --namespace "${HELM_NAMESPACE:-kube-system}" -o name 2>/dev/null | grep -E 'vpa-' || true); do
      echo "--- ${pod} ---"
      kubectl logs --namespace "${HELM_NAMESPACE:-kube-system}" "${pod}" --all-containers --tail=80 2>&1 || true
    done
    echo "============================================================"
  fi
}
trap dump_diagnostics EXIT

# Deploy metrics server for E2E tests via Helm chart.
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm repo update metrics-server
helm upgrade --install --set args={--kubelet-insecure-tls} metrics-server metrics-server/metrics-server \
  --namespace kube-system --version 3.13.0 --wait

# Build and load Docker images for each component.
for COMPONENT in ${COMPONENTS}; do
  ALL_ARCHITECTURES=${ARCH} make --directory "${SCRIPT_ROOT}/pkg/${COMPONENT}" docker-build REGISTRY="${REGISTRY}" TAG="${TAG}"
  docker tag "${REGISTRY}/vpa-${COMPONENT}-${ARCH}:${TAG}" "${REGISTRY}/vpa-${COMPONENT}:${TAG}"
  kind load docker-image "${REGISTRY}/vpa-${COMPONENT}:${TAG}"
done

# helm-settings.sh defines HELM_CHART_PATH, VALUES_FILE, HELM_RELEASE_NAME,
# HELM_NAMESPACE.
source "${SCRIPT_ROOT}/hack/e2e/helm-settings.sh"

# Sanity checks before we touch the cluster.
if [[ ! -d "${HELM_CHART_PATH}" ]]; then
  echo "ERROR: Helm chart not found at ${HELM_CHART_PATH}"; exit 1
fi
if [[ ! -f "${VALUES_FILE}" ]]; then
  echo "ERROR: Values file not found at ${VALUES_FILE}"; exit 1
fi

# Build per-suite Helm value overrides.
HELM_SET_ARGS=()
for COMPONENT in ${COMPONENTS}; do
  case ${COMPONENT} in
    recommender)
      HELM_SET_ARGS+=(
        "--set" "recommender.enabled=true"
        "--set" "recommender.image.repository=${REGISTRY}/vpa-recommender"
        "--set" "recommender.image.tag=${TAG}"
      )
      ;;
    updater)
      HELM_SET_ARGS+=(
        "--set" "updater.enabled=true"
        "--set" "updater.image.repository=${REGISTRY}/vpa-updater"
        "--set" "updater.image.tag=${TAG}"
      )
      ;;
    admission-controller)
      HELM_SET_ARGS+=(
        "--set" "admissionController.enabled=true"
        "--set" "admissionController.image.repository=${REGISTRY}/vpa-admission-controller"
        "--set" "admissionController.image.tag=${TAG}"
      )
      ;;
  esac
done

# Optional feature gates per component.
RECOM_ARGS_IDX=0
if [ -n "${FEATURE_GATES:-}" ]; then
  ESCAPED_FEATURE_GATES="${FEATURE_GATES//,/\\,}"
  for COMPONENT in ${COMPONENTS}; do
    case ${COMPONENT} in
      recommender)
        HELM_SET_ARGS+=("--set" "recommender.extraArgs[${RECOM_ARGS_IDX}]=--feature-gates=${ESCAPED_FEATURE_GATES}")
        RECOM_ARGS_IDX=$((RECOM_ARGS_IDX + 1))
        ;;
      updater)
        HELM_SET_ARGS+=("--set" "updater.extraArgs[0]=--feature-gates=${ESCAPED_FEATURE_GATES}")
        ;;
      admission-controller)
        # IMPORTANT: extraArgs is a list. Setting only one element via --set
        # would replace the entire list (including --reload-cert from
        # values-e2e.yaml). Append --reload-cert as the next index.
        HELM_SET_ARGS+=(
          "--set" "admissionController.extraArgs[0]=--feature-gates=${ESCAPED_FEATURE_GATES}"
          "--set" "admissionController.extraArgs[1]=--reload-cert"
        )
        ;;
    esac
  done
fi

echo "===> Cleaning up any previous VPA installation in ${HELM_NAMESPACE}"
helm uninstall "${HELM_RELEASE_NAME}" --namespace "${HELM_NAMESPACE}" 2>/dev/null || true
kubectl delete crd verticalpodautoscalers.autoscaling.k8s.io --ignore-not-found=true
kubectl delete crd verticalpodautoscalercheckpoints.autoscaling.k8s.io --ignore-not-found=true
kubectl delete mutatingwebhookconfiguration vpa-webhook-config --ignore-not-found=true

# Apply CRDs explicitly before helm install (idempotent).
echo "===> Applying CRDs from ${HELM_CHART_PATH}/crds/"
kubectl apply -f "${HELM_CHART_PATH}/crds/"

# Handle external metrics special case.
if [[ "${SUITE}" == "recommender-externalmetrics" ]]; then
  echo "===> Setting up external metrics infrastructure"
  kubectl delete namespace monitoring --ignore-not-found=true
  kubectl create namespace monitoring
  kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/prometheus.yaml"
  kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/prometheus-adapter.yaml"
  kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/metrics-pump.yaml"

  echo "===> Configuring recommender with external metrics args"
  HELM_SET_ARGS+=(
    "--values" "${SCRIPT_ROOT}/hack/e2e/values-external-metrics.yaml"
    "--set" "recommender.extraArgs[${RECOM_ARGS_IDX}]=--use-external-metrics=true"
    "--set" "recommender.extraArgs[$((RECOM_ARGS_IDX + 1))]=--external-metrics-cpu-metric=cpu"
    "--set" "recommender.extraArgs[$((RECOM_ARGS_IDX + 2))]=--external-metrics-memory-metric=mem"
  )
fi

# Generate TLS certs for admission-controller AFTER cleanup, BEFORE helm
# install. Idempotent secret creation: delete first, then recreate.
for COMPONENT in ${COMPONENTS}; do
  if [[ "${COMPONENT}" == "admission-controller" ]]; then
    echo "===> Generating TLS certificates for admission-controller"
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "${TMP_DIR}"' RETURN
    CN_BASE="vpa_webhook"

    cat > "${TMP_DIR}/server.conf" << EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = DNS:vpa-webhook.${HELM_NAMESPACE}.svc
EOF

    openssl genrsa -out "${TMP_DIR}/caKey.pem" 2048
    openssl req -x509 -new -nodes -key "${TMP_DIR}/caKey.pem" -days 100000 \
      -out "${TMP_DIR}/caCert.pem" -subj "/CN=${CN_BASE}_ca" \
      -addext "subjectAltName = DNS:${CN_BASE}_ca"
    openssl genrsa -out "${TMP_DIR}/serverKey.pem" 2048
    openssl req -new -key "${TMP_DIR}/serverKey.pem" -out "${TMP_DIR}/server.csr" \
      -subj "/CN=vpa-webhook.${HELM_NAMESPACE}.svc" -config "${TMP_DIR}/server.conf"
    openssl x509 -req -in "${TMP_DIR}/server.csr" -CA "${TMP_DIR}/caCert.pem" \
      -CAkey "${TMP_DIR}/caKey.pem" -CAcreateserial -out "${TMP_DIR}/serverCert.pem" \
      -days 100000 -extensions SAN -extensions v3_req -extfile "${TMP_DIR}/server.conf"

    kubectl delete secret -n "${HELM_NAMESPACE}" vpa-tls-certs --ignore-not-found=true
    kubectl create secret -n "${HELM_NAMESPACE}" generic vpa-tls-certs \
      --from-file="${TMP_DIR}/caKey.pem" \
      --from-file="${TMP_DIR}/caCert.pem" \
      --from-file="${TMP_DIR}/serverKey.pem" \
      --from-file="${TMP_DIR}/serverCert.pem"

    echo "===> Generating E2E rotation test certificates"
    openssl genrsa -out "${TMP_DIR}/e2eCaKey.pem" 2048
    openssl req -x509 -new -nodes -key "${TMP_DIR}/e2eCaKey.pem" -days 100000 \
      -out "${TMP_DIR}/e2eCaCert.pem" -subj "/CN=${CN_BASE}_e2e_ca" \
      -addext "subjectAltName = DNS:${CN_BASE}_e2e_ca"
    openssl genrsa -out "${TMP_DIR}/e2eKey.pem" 2048
    openssl req -new -key "${TMP_DIR}/e2eKey.pem" -out "${TMP_DIR}/e2e.csr" \
      -subj "/CN=vpa-webhook.${HELM_NAMESPACE}.svc" -config "${TMP_DIR}/server.conf"
    openssl x509 -req -in "${TMP_DIR}/e2e.csr" -CA "${TMP_DIR}/e2eCaCert.pem" \
      -CAkey "${TMP_DIR}/e2eCaKey.pem" -CAcreateserial -out "${TMP_DIR}/e2eCert.pem" \
      -days 100000 -extensions SAN -extensions v3_req -extfile "${TMP_DIR}/server.conf"

    kubectl delete secret -n "${HELM_NAMESPACE}" vpa-e2e-certs --ignore-not-found=true
    kubectl create secret -n "${HELM_NAMESPACE}" generic vpa-e2e-certs \
      --from-file="${TMP_DIR}/e2eCaKey.pem" \
      --from-file="${TMP_DIR}/e2eCaCert.pem" \
      --from-file="${TMP_DIR}/e2eKey.pem" \
      --from-file="${TMP_DIR}/e2eCert.pem"

    rm -rf "${TMP_DIR}"
    echo "===> TLS certificates created"
    break
  fi
done

# Optional extra Helm values file (must be PREPENDED so values-e2e.yaml stays
# the base layer; helm precedence is later -f overrides earlier).
if [[ -n "${EXTRA_HELM_VALUES:-}" ]]; then
  echo "===> Using extra Helm values from: ${EXTRA_HELM_VALUES}"
  HELM_SET_ARGS=("--values" "${EXTRA_HELM_VALUES}" "${HELM_SET_ARGS[@]}")
fi

echo "===> Installing/Upgrading VPA via Helm chart"
helm upgrade --install "${HELM_RELEASE_NAME}" "${HELM_CHART_PATH}" \
  --namespace "${HELM_NAMESPACE}" \
  --values "${VALUES_FILE}" \
  "${HELM_SET_ARGS[@]}" \
  --atomic \
  --wait \
  --timeout 15m

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
