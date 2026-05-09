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

export REGISTRY=gcr.io/`gcloud config get-value core/project`
export TAG=latest

echo "Configuring registry authentication"
mkdir -p "${HOME}/.docker"
gcloud auth configure-docker -q

# Standardize configuration across environments (Addresses Review #1)
source "${SCRIPT_ROOT}/hack/e2e/helm-settings.sh"

for i in ${COMPONENTS}; do
  ALL_ARCHITECTURES=amd64 make --directory ${SCRIPT_ROOT}/pkg/${i} release
done

echo "Deploying VPA components via Helm for E2E CI..."

# Generate TLS certificates natively to prevent webhook conflicts in CI
for COMPONENT in ${COMPONENTS}; do
  if [[ "${COMPONENT}" == "admission-controller" ]]; then
    echo " ** Generating TLS certificates for admission-controller"
    TMP_DIR=$(mktemp -d)
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
    openssl req -x509 -new -nodes -key "${TMP_DIR}/caKey.pem" -days 10000 \
      -out "${TMP_DIR}/caCert.pem" -subj "/CN=${CN_BASE}_ca" \
      -addext "subjectAltName = DNS:${CN_BASE}_ca"
    openssl genrsa -out "${TMP_DIR}/serverKey.pem" 2048
    openssl req -new -key "${TMP_DIR}/serverKey.pem" -out "${TMP_DIR}/server.csr" \
      -subj "/CN=vpa-webhook.${HELM_NAMESPACE}.svc" -config "${TMP_DIR}/server.conf"
    openssl x509 -req -in "${TMP_DIR}/server.csr" -CA "${TMP_DIR}/caCert.pem" \
      -CAkey "${TMP_DIR}/caKey.pem" -CAcreateserial -out "${TMP_DIR}/serverCert.pem" \
      -days 10000 -extensions SAN -extensions v3_req -extfile "${TMP_DIR}/server.conf"

    kubectl delete secret -n ${HELM_NAMESPACE} vpa-tls-certs --ignore-not-found=true
    kubectl create secret -n ${HELM_NAMESPACE} generic vpa-tls-certs \
      --from-file="${TMP_DIR}/caKey.pem" \
      --from-file="${TMP_DIR}/caCert.pem" \
      --from-file="${TMP_DIR}/serverKey.pem" \
      --from-file="${TMP_DIR}/serverCert.pem"

    echo " ** Generating E2E rotation test certificates"
    openssl genrsa -out "${TMP_DIR}/e2eCaKey.pem" 2048
    openssl req -x509 -new -nodes -key "${TMP_DIR}/e2eCaKey.pem" -days 10000 \
      -out "${TMP_DIR}/e2eCaCert.pem" -subj "/CN=${CN_BASE}_e2e_ca" \
      -addext "subjectAltName = DNS:${CN_BASE}_e2e_ca"
    openssl genrsa -out "${TMP_DIR}/e2eKey.pem" 2048
    openssl req -new -key "${TMP_DIR}/e2eKey.pem" -out "${TMP_DIR}/e2e.csr" \
      -subj "/CN=vpa-webhook.${HELM_NAMESPACE}.svc" -config "${TMP