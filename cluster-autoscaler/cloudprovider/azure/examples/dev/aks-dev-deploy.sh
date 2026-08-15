#!/usr/bin/env bash

# Copyright 2024 The Kubernetes Authors.
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

set -euo pipefail

# This script deploys an AKS cluster with a user node pool and ACR,
# and workload identity to be used for cluster-autoscaler. It creates
# the necessary role assignments to allow the cluster-autoscaler to manage VMSS,
# configures skaffold to use the ACR for the cluster-autoscaler deployment,
# and updates the cluster-autoscaler deployment with the necessary values,
# ready to be used with skaffold dev/run/debug.

# assumed logged in (az login) and subscription set (az account set --subscription ...)

# set resource group and ACR name (preferably unique)
RG=${CODESPACE_NAME:-cluster-autoscaler-test}
ACR_NAME=$(echo "$RG" | tr -d -) # remove hyphens
az group create --name "${RG}" --location westus3 --output none

# deploy AKS & ACR
DEPLOYMENT_JSON=$(az deployment group create --name aks-dev --resource-group "${RG}" \
  --template-file ./aks-dev.bicep \
  --parameters acrName="${ACR_NAME}")

# get relevant information
RESOURCE_GROUP_MC=$(jq -r ".properties.outputs.nodeResourceGroup.value"                <<< "$DEPLOYMENT_JSON")
USER_POOL_NAME=$(   jq -r ".properties.outputs.userNodePoolName.value"                 <<< "$DEPLOYMENT_JSON")
AKS_NAME=$(         jq -r ".properties.outputs.aksName.value"                          <<< "$DEPLOYMENT_JSON")
CAS_UAI_PRINCIPAL=$(jq -r ".properties.outputs.casUserAssignedIdentityPrincipal.value" <<< "$DEPLOYMENT_JSON")
CAS_UAI_CLIENTID=$( jq -r ".properties.outputs.casUserAssignedIdentityClientId.value"  <<< "$DEPLOYMENT_JSON")

# confgure dev environment
az aks get-credentials --name "${AKS_NAME}" --resource-group "${RG}"
az acr login --name "${ACR_NAME}"
skaffold config set default-repo "${ACR_NAME}.azurecr.io/cluster-autoscaler"

# create role assignments to let CAS manage VMSS
az role assignment create \
  --assignee "${CAS_UAI_PRINCIPAL}" \
  --scope "$(az group show --name "${RESOURCE_GROUP_MC}" --query "id" --output tsv)" \
  --role "Virtual Machine Contributor" \
  --output none

# prep values for and update CAS deployment
VMSS_NAME=$(az resource list \
  --tag aks-managed-poolName="${USER_POOL_NAME}" \
  --query "[?resourceGroup=='${RESOURCE_GROUP_MC}'].name" \
  --output tsv)
TENANT_ID_B64=$(az account show --query tenantId --output tsv | base64)
RESOURCE_GROUP_MC_B64=$(base64 <<< "$RESOURCE_GROUP_MC")
SUBSCRIPTION_ID_B64=$(az account show --query id --output tsv | base64)

export TENANT_ID_B64 RESOURCE_GROUP_MC_B64 VMSS_NAME CAS_UAI_CLIENTID SUBSCRIPTION_ID_B64

yq '(.. | select(tag == "!!str")) |= envsubst(nu)' \
    cluster-autoscaler-vmss-wi-dynamic.yaml.tpl > \
    cluster-autoscaler-vmss-wi-dynamic.yaml

# skaffold dev/run/debug

exit

# To recover access after restarting codespace with existing AKS and ACR:
# az login & az account set -n ...
# az aks get-credentials -n cas-test -g $CODESPACE_NAME 
# ACR_NAME=$(echo "$CODESPACE_NAME" | tr -d -)
# az acr login -n $ACR_NAME
# skaffold config set default-repo "${ACR_NAME}.azurecr.io/cluster-autoscaler"
# skaffold dev/run/debug
