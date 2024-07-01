REPO_ROOT:=$(shell git rev-parse --show-toplevel)
CAS_ROOT:=$(REPO_ROOT)/cluster-autoscaler
CAS_CHART:=$(REPO_ROOT)/charts/cluster-autoscaler

include $(CAS_ROOT)/Makefile

TOOLS_BIN_DIR := $(abspath bin)

export PATH := $(TOOLS_BIN_DIR):$(PATH)

HELM_VER := v3.15.2
HELM_BIN := helm
HELM := $(TOOLS_BIN_DIR)/$(HELM_BIN)-$(HELM_VER)

helm: $(HELM)

$(HELM):
	mkdir -p $(TOOLS_BIN_DIR)
	rm -f "$(TOOLS_BIN_DIR)/$(HELM_BIN)*"
	curl --retry 3 -fsSL -o $(TOOLS_BIN_DIR)/get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
	chmod 700 $(TOOLS_BIN_DIR)/get_helm.sh
	USE_SUDO=false HELM_INSTALL_DIR=$(TOOLS_BIN_DIR) DESIRED_VERSION=$(HELM_VER) BINARY_NAME=$(HELM_BIN)-$(HELM_VER) $(TOOLS_BIN_DIR)/get_helm.sh
	ln -sf $(HELM) $(TOOLS_BIN_DIR)/$(HELM_BIN)
	rm -f $(TOOLS_BIN_DIR)/get_helm.sh

CLUSTER_AUTOSCALER_NAMESPACE?=default
CLUSTER_AUTOSCALER_SERVICEACCOUNT_NAME?=cluster-autoscaler

.PHONY: install-e2e
install-e2e: $(HELM)
	$(MAKE) -C $(CAS_ROOT) build-arch-$(GOARCH) make-image-arch-$(GOARCH)
	docker push $(IMAGE)-$(GOARCH):$(TAG)
	$(HELM) upgrade --install cluster-autoscaler $(CAS_CHART) \
		--namespace $(CLUSTER_AUTOSCALER_NAMESPACE) --create-namespace \
		--set cloudProvider=azure \
		--set azureTenantID=$(AZURE_TENANT_ID) \
		--set azureSubscriptionID=$(AZURE_SUBSCRIPTION_ID) \
		--set azureUseWorkloadIdentityExtension=true \
		--set-string podLabels."azure\.workload\.identity/use"=true \
		--set rbac.serviceAccount.name=$(CLUSTER_AUTOSCALER_SERVICEACCOUNT_NAME) \
		--set rbac.serviceAccount.annotations."azure\.workload\.identity/tenant-id"=$(AZURE_TENANT_ID) \
		--set rbac.serviceAccount.annotations."azure\.workload\.identity/client-id"="$$(KUBECONFIG= kubectl get userassignedidentities -o jsonpath='{.items[0].status.clientId}')" \
		--set autoDiscovery.clusterName="$$(KUBECONFIG= kubectl get cluster -o jsonpath='{.items[0].metadata.name}')" \
		--set azureResourceGroup="$$(KUBECONFIG= kubectl get managedclusters -o jsonpath='{.items[0].status.nodeResourceGroup}')" \
		--set nodeSelector."kubernetes\.io/os"=linux \
		--set image.repository=$(IMAGE)-$(GOARCH) \
		--set image.tag=$(TAG) \
		--set image.pullPolicy=Always \
		--set extraArgs.scale-down-delay-after-add=10s \
		--set extraArgs.scale-down-unneeded-time=10s \
		--set extraArgs.scale-down-candidates-pool-ratio=1.0 \
		--set extraArgs.unremovable-node-recheck-timeout=10s \
		--set extraArgs.skip-nodes-with-system-pods=false \
		--set extraArgs.skip-nodes-with-local-storage=false \
		--wait

ARTIFACTS?=_artifacts

.PHONY: test-e2e
test-e2e: install-e2e
	go run github.com/onsi/ginkgo/v2/ginkgo -v --trace --output-dir "$(ARTIFACTS)" --junit-report="junit.e2e_suite.1.xml" e2e -- \
		-resource-group="$$(KUBECONFIG= kubectl get managedclusters -o jsonpath='{.items[0].status.nodeResourceGroup}')"
