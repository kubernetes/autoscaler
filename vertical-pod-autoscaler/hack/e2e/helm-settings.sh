#!/bin/bash

# Shared Helm configuration for E2E deployments
HELM_CHART_PATH="${SCRIPT_ROOT}/charts/vertical-pod-autoscaler"
VALUES_FILE="${SCRIPT_ROOT}/hack/e2e/values-e2e.yaml"
HELM_RELEASE_NAME="vpa"
HELM_NAMESPACE="kube-system"