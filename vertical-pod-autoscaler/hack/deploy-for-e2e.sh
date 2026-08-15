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
HELM_VERSION="${HELM_VERSION:-v4.1.4}"

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

# The kubekins-e2e prow image does not ship Helm; install on demand.
function ensure_helm_installed() {
  if command -v helm >/dev/null 2>&1; then
    return 0
  fi
  echo "Installing Helm ${HELM_VERSION}"
  local tmp; tmp=$(mktemp -d)
  curl -fsSL "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" -o "${tmp}/helm.tgz"
  tar -xzf "${tmp}/helm.tgz" -C "${tmp}"
  install -m 0755 "${tmp}/linux-amd64/helm" /usr/local/bin/helm
  rm -rf "${tmp}"
}
ensure_helm_installed

source "${SCRIPT_ROOT}/hack/e2e/helm-settings.sh"

export REGISTRY=gcr.io/`gcloud config get-value core/project`
export TAG=latest

echo "Configuring registry authentication"
mkdir -p "${HOME}/.docker"
gcloud auth configure-docker -q

# Build images and (for admission-controller) generate TLS certs before Helm install.
for COMPONENT in ${COMPONENTS}; do
  if [[ "${COMPONENT}" == "admission-controller" ]]; then
    (cd "${SCRIPT_ROOT}/pkg/${COMPONENT}" && bash ./gencerts.sh e2e)
  fi
  ALL_ARCHITECTURES=amd64 make --directory "${SCRIPT_ROOT}/pkg/${COMPONENT}" release
done

# Build per-suite Helm value overrides. values-e2e.yaml disables every
# component by default; we enable only the ones the suite needs and point
# them at the PR-built images in our project's GCR.
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

# Propagate FEATURE_GATES (set on alpha/beta CI lanes) to component args.
# Helm --set replaces lists; for admissionController we re-add --reload-cert
# (set in values-e2e.yaml for the cert-rotation test) so it isn't dropped.
# Commas in the value are escaped so Helm --set treats it as one assignment.
if [[ -n "${FEATURE_GATES:-}" ]]; then
  ESCAPED_FEATURE_GATES="${FEATURE_GATES//,/\\,}"
  for COMPONENT in ${COMPONENTS}; do
    case ${COMPONENT} in
      recommender)
        HELM_SET_ARGS+=("--set" "recommender.extraArgs[0]=--feature-gates=${ESCAPED_FEATURE_GATES}")
        ;;
      updater)
        HELM_SET_ARGS+=("--set" "updater.extraArgs[0]=--feature-gates=${ESCAPED_FEATURE_GATES}")
        ;;
      admission-controller)
        HELM_SET_ARGS+=(
          "--set" "admissionController.extraArgs[0]=--reload-cert"
          "--set" "admissionController.extraArgs[1]=--feature-gates=${ESCAPED_FEATURE_GATES}"
        )
        ;;
    esac
  done
fi

# Helm 3+ installs CRDs from chart's crds/ directory on first install of a
# release. CI runs on fresh VMs, so this is always a first install.
helm upgrade --install "${HELM_RELEASE_NAME}" "${HELM_CHART_PATH}" \
  --namespace "${HELM_NAMESPACE}" \
  --values "${VALUES_FILE}" \
  "${HELM_SET_ARGS[@]}" \
  --debug \
  --wait \
  --timeout 10m
