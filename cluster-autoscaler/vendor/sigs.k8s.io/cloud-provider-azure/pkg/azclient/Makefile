# Copyright 2022 The Kubernetes Authors.
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


# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: generate

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: fmt
fmt: goimports ## Run go fmt against code.
	$(GOIMPORTS) -w -local sigs.k8s.io/cloud-provider-azure/pkg/azclient .

.PHONY: vet
vet: golangci-lint ## Run go vet against code.
	pushd client-gen; $(LOCALBIN)/golangci-lint run --timeout 10m ./... ;popd

##@ Build
.PHONY: build
TYPESCAFFOLD = $(LOCALBIN)/typescaffold
CLIENTGEN = $(LOCALBIN)/client-gen
build: ## Build manager binary.
	pushd client-gen; CGO_ENABLED=0 go build -o ../bin/client-gen ./cmd/client-gen/ ;popd
	pushd client-gen; CGO_ENABLED=0 go build -o ../bin/typescaffold ./cmd/typescaffold/;popd

.PHONY: generate
generate: install-dependencies build generatecode generateimpl fmt vet-all

.PHONY: generatecode
generatecode: build ## Generate client
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4 --package-alias armcontainerservice --resource ManagedCluster --client-name ManagedClustersClient  --ratelimitkey containerServiceRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry --package-alias armcontainerregistry --resource Registry --client-name RegistriesClient --verbs get,delete,listbyrg
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources --package-alias resources --resource Deployment --client-name DeploymentsClient --verbs delete --ratelimitkey deploymentRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources --package-alias resources --resource ResourceGroup --client-name ResourceGroupsClient 
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource Disk --client-name DisksClient --ratelimitkey diskRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource AvailabilitySet --client-name AvailabilitySetsClient --verbs get,list --ratelimitkey availabilitySetRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource VirtualMachine --client-name VirtualMachinesClient --verbs createorupdate,delete,list --expand --ratelimitkey virtualMachineRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource VirtualMachineScaleSet --client-name VirtualMachineScaleSetsClient --verbs get,createorupdate,delete,list --ratelimitkey virtualMachineSizesRateLimit --expand
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource VirtualMachineScaleSet --subresource VirtualMachineScaleSetVM --client-name VirtualMachineScaleSetVMsClient --verbs get,delete,list 
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource Snapshot --client-name SnapshotsClient --verbs get,createorupdate,delete --ratelimitkey snapshotRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 --package-alias armcompute --resource SSHPublicKeyResource --client-name SSHPublicKeysClient --verbs get,listbyrg
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource VirtualNetwork --subresource Subnet --client-name SubnetsClient --verbs get,createorupdate,delete,list --expand --ratelimitkey subnetsRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource VirtualNetwork --client-name VirtualNetworksClient --verbs get,createorupdate,delete,list --expand 
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource Interface --client-name InterfacesClient --verbs get,createorupdate,delete,list --expand --ratelimitkey interfaceRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource LoadBalancer --client-name LoadBalancersClient --verbs get,createorupdate,delete,list --expand --ratelimitkey loadBalancerRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource PrivateEndpoint --client-name PrivateEndpointsClient --verbs get,createorupdate --expand --ratelimitkey privateEndpointRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource PublicIPAddress --client-name PublicIPAddressesClient --verbs get,createorupdate,delete,list --expand --ratelimitkey publicIPAddressRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource PublicIPPrefix --client-name PublicIPPrefixesClient --verbs get,createorupdate,delete,list --expand 
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource RouteTable --client-name RouteTablesClient --verbs get,createorupdate,delete,list  --ratelimitkey routeTableRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource SecurityGroup --client-name SecurityGroupsClient --verbs get,createorupdate,delete,list --ratelimitkey securityGroupRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource PrivateLinkService --client-name PrivateLinkServicesClient --verbs get,createorupdate,delete,list --expand --ratelimitkey privateLinkServiceRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 --package-alias armnetwork --resource IPGroup --client-name IPGroupsClient --verbs get,createorupdate,delete,listbyrg --expand --ratelimitkey ipGroupRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage --package-alias armstorage --resource Account --client-name AccountsClient --verbs listbyrg --expand
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns --package-alias armprivatedns --resource PrivateZone --client-name PrivateZonesClient  --verbs get,createorupdate --ratelimitkey privateDNSRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns --package-alias armprivatedns --resource PrivateZone --subresource VirtualNetworkLink --client-name VirtualNetworkLinksClient --verbs get,createorupdate --ratelimitkey virtualNetworkRateLimit
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage --package-alias armstorage --resource Account --subresource FileShare --client-name FileSharesClient --verbs get,createorupdate,delete,list
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage --package-alias armstorage --resource Account --subresource BlobContainer --client-name BlobContainersClient --verbs get,list
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage --package-alias armstorage --resource BlobServiceProperties --client-name BlobServicesClient
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault --package-alias armkeyvault --resource Vault --subresource Secret --client-name SecretsClient --verbs get,createorupdate,delete,list
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault --package-alias armkeyvault --resource Vault --client-name VaultsClient --verbs get,createorupdate,delete,list
	$(TYPESCAFFOLD) --package github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources --package-alias armresources --resource Provider --client-name ProvidersClient 


.PHONY: generateimpl
generateimpl: build ## Generate client
	PATH=$(LOCALBIN):$$PATH $(CLIENTGEN) clientgen:headerFile=../../hack/boilerplate/boilerplate.gomock.txt paths=./...

.PHONY: vet-all
vet-all: golangci-lint ## Run go vet against code.
	$(LOCALBIN)/golangci-lint run --timeout 10m ./...


ifndef ignore-not-found
  ignore-not-found = false
endif
##@ Build Dependencies

.PHONY: install-dependencies
install-dependencies: golangci-lint goimports mockgen ginkgo## Install all build dependencies.

GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) latest

GOIMPORTS ?= $(LOCALBIN)/goimports
.PHONY: goimports
goimports: $(GOIMPORTS) ## Download goimports locally if necessary.
$(GOIMPORTS): $(LOCALBIN)
	test -s $(LOCALBIN)/goimports || GOBIN=$(LOCALBIN)  go install golang.org/x/tools/cmd/goimports@latest

MOCKGEN ?= $(LOCALBIN)/mockgen
.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/mockgen || GOBIN=$(LOCALBIN)  go install go.uber.org/mock/mockgen@latest

GINKGO ?= $(LOCALBIN)/ginkgo
.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.
$(GINKGO): $(LOCALBIN)
	test -s $(LOCALBIN)/ginkgo || GOBIN=$(LOCALBIN)  go install github.com/onsi/ginkgo/v2/ginkgo@latest
