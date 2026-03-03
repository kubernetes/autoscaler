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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
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

# Local KIND images
export REGISTRY=${REGISTRY:-localhost:5001}
export TAG=${TAG:-latest}


# Deploy metrics server for E2E tests
kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/k8s-metrics-server.yaml"

# Build and load Docker images for each component
for COMPONENT in ${COMPONENTS}; do
  ALL_ARCHITECTURES=${ARCH} make --directory "${SCRIPT_ROOT}/pkg/${COMPONENT}" docker-build REGISTRY="${REGISTRY}" TAG="${TAG}"
  docker tag "${REGISTRY}/vpa-${COMPONENT}-${ARCH}:${TAG}" "${REGISTRY}/vpa-${COMPONENT}:${TAG}"
  kind load docker-image "${REGISTRY}/vpa-${COMPONENT}:${TAG}"
done



# Prepare Helm values
HELM_CHART_PATH="${SCRIPT_ROOT}/charts/vertical-pod-autoscaler"
VALUES_FILE="${SCRIPT_ROOT}/hack/e2e/values-e2e-local.yaml"
HELM_RELEASE_NAME="vpa"
HELM_NAMESPACE="kube-system"

# Build dynamic Helm set arguments based on components
HELM_SET_ARGS=()

# Enable components
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

# Add feature gates if specified
RECOM_ARGS_IDX=0
if [ -n "${FEATURE_GATES:-}" ]; then
  # Add feature gates to each enabled component
  for COMPONENT in ${COMPONENTS}; do
    case ${COMPONENT} in
      recommender)
        HELM_SET_ARGS+=("--set" "recommender.extraArgs[${RECOM_ARGS_IDX}]=--feature-gates=${FEATURE_GATES}")
        RECOM_ARGS_IDX=$((RECOM_ARGS_IDX + 1))
        ;;
      updater)
        HELM_SET_ARGS+=("--set" "updater.extraArgs[0]=--feature-gates=${FEATURE_GATES}")
        ;;
      admission-controller)
        HELM_SET_ARGS+=("--set" "admissionController.extraArgs[0]=--feature-gates=${FEATURE_GATES}")
        ;;
    esac
  done
fi

# Uninstall any existing VPA Helm release
helm uninstall ${HELM_RELEASE_NAME} --namespace ${HELM_NAMESPACE} 2>/dev/null || true

# Handle external metrics special case
if [[ "${SUITE}" == "recommender-externalmetrics" ]]; then
  echo " ** Setting up external metrics infrastructure"

  # Deploy Prometheus and adapter for external metrics
  kubectl delete namespace monitoring --ignore-not-found=true
  kubectl create namespace monitoring
  kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/prometheus.yaml"
  kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/prometheus-adapter.yaml"
  kubectl apply -f "${SCRIPT_ROOT}/hack/e2e/metrics-pump.yaml"

  # Initialize external metrics args

  echo " ** Configuring recommender with external metrics args"
  HELM_SET_ARGS+=("--values" "${SCRIPT_ROOT}/hack/e2e/values-external-metrics.yaml")
  HELM_SET_ARGS+=("--set" "recommender.extraArgs[${RECOM_ARGS_IDX}]=--use-external-metrics=true")
  HELM_SET_ARGS+=("--set" "recommender.extraArgs[$((RECOM_ARGS_IDX + 1))]=--external-metrics-cpu-metric=cpu")
  HELM_SET_ARGS+=("--set" "recommender.extraArgs[$((RECOM_ARGS_IDX + 2))]=--external-metrics-memory-metric=mem")
fi

# Generate TLS certificates for admission-controller before Helm install
# This is needed when using registerWebhook: true (application-managed webhook)
for COMPONENT in ${COMPONENTS}; do
  if [[ "${COMPONENT}" == "admission-controller" ]]; then
    echo " ** Generating TLS certificates for admission-controller"
    TMP_DIR=$(mktemp -d)
    CN_BASE="vpa_webhook"

    # Create server config for certificate generation
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

    # Generate main CA and server certificates
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

    # Create the main TLS certs secret (with legacy key names for compatibility)
    kubectl delete secret -n ${HELM_NAMESPACE} vpa-tls-certs --ignore-not-found=true
    kubectl create secret -n ${HELM_NAMESPACE} generic vpa-tls-certs \
      --from-file="${TMP_DIR}/caKey.pem" \
      --from-file="${TMP_DIR}/caCert.pem" \
      --from-file="${TMP_DIR}/serverKey.pem" \
      --from-file="${TMP_DIR}/serverCert.pem"

    # Generate E2E rotation test certificates (different CA for testing cert rotation)
    echo " ** Generating E2E rotation test certificates"
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

    # Create the E2E certs secret
    kubectl delete secret -n ${HELM_NAMESPACE} vpa-e2e-certs --ignore-not-found=true
    kubectl create secret -n ${HELM_NAMESPACE} generic vpa-e2e-certs \
      --from-file="${TMP_DIR}/e2eCaKey.pem" \
      --from-file="${TMP_DIR}/e2eCaCert.pem" \
      --from-file="${TMP_DIR}/e2eKey.pem" \
      --from-file="${TMP_DIR}/e2eCert.pem"

    # Clean up temp directory
    rm -rf "${TMP_DIR}"
    echo " ** TLS certificates created"
    break
  fi
done

# Install/Upgrade VPA using Helm chart
echo " ** Installing/Upgrading VPA via Helm chart"
helm upgrade --install ${HELM_RELEASE_NAME} "${HELM_CHART_PATH}" \
  --namespace ${HELM_NAMESPACE} \
  --values "${VALUES_FILE}" \
  "${HELM_SET_ARGS[@]}" \
  --wait \
  --timeout 15m
