REPO_ROOT:=$(shell git rev-parse --show-toplevel)
CAS_ROOT:=$(REPO_ROOT)/cluster-autoscaler
DEV_DIR:=$(CAS_ROOT)/cloudprovider/azure/examples/dev

BUILD_TAGS=azure

include $(CAS_ROOT)/Makefile

CLUSTER_AUTOSCALER_NAMESPACE?=default
CLUSTER_AUTOSCALER_SERVICEACCOUNT_NAME?=cluster-autoscaler

# TEST_SUITE selects a specific test suite directory (e.g., TEST_SUITE=scaleup).
# When unset, all suites under suites/ are run.
TEST_SUITE?=
TEST_TIMEOUT?=3h

# Compute ginkgo suite path: ./suites/... (all) or ./suites/<name>/... (single)
SUITE_PATH=$(if $(TEST_SUITE),./suites/$(shell echo $(TEST_SUITE) | tr A-Z a-z)/...,./suites/...)
FOCUS?=
LABEL_FILTER?=
ARTIFACTS?=_artifacts

# --------------------------------------------------------------------------
# CI target (Prow — builds CAS image, discovers cluster info from CAPZ)
# --------------------------------------------------------------------------

.PHONY: build-e2e
build-e2e:
	$(MAKE) -C $(CAS_ROOT) build-arch-$(GOARCH) make-image-arch-$(GOARCH) BUILD_TAGS=${BUILD_TAGS}
	docker push $(IMAGE)-$(GOARCH):$(TAG)

.PHONY: test-e2e
test-e2e: build-e2e ## CI: build image + run tests (Prow/CAPZ)
	go run github.com/onsi/ginkgo/v2/ginkgo --tags e2e -v --trace \
		--timeout $(TEST_TIMEOUT) \
		--output-dir "$(ARTIFACTS)" --junit-report="junit.e2e_suite.1.xml" \
		$(SUITE_PATH) -- \
		-resource-group="$$(KUBECONFIG= kubectl get managedclusters -n default -o jsonpath='{.items[0].status.nodeResourceGroup}')" \
		-cluster-name="$$(KUBECONFIG= kubectl get cluster -n default -o jsonpath='{.items[0].metadata.name}')" \
		-client-id="$$(KUBECONFIG= kubectl get userassignedidentities -n default -o jsonpath='{.items[0].status.clientId}')" \
		-cas-namespace="$(CLUSTER_AUTOSCALER_NAMESPACE)" \
		-cas-serviceaccount-name="$(CLUSTER_AUTOSCALER_SERVICEACCOUNT_NAME)" \
		-cas-image-repository="$(IMAGE)-$(GOARCH)" \
		-cas-image-tag="$(TAG)"

# --------------------------------------------------------------------------
# Local developer targets
# --------------------------------------------------------------------------
# Prerequisites:
#   1. az login
#   2. make setup-cluster   (creates AKS + ACR + identity, one-time)
#   3. make deploy-local    (builds + deploys CAS via skaffold)
#   4. make e2etests        (runs the tests)
#
# Required env var: AZURE_RESOURCE_GROUP (set automatically by setup-cluster)

.PHONY: help
help: ## Display help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Cluster Setup (one-time)

.PHONY: setup-cluster
setup-cluster: ## Create AKS cluster + ACR + workload identity for e2e testing
	cd $(DEV_DIR) && bash ./aks-dev-deploy.sh

##@ Build & Deploy

.PHONY: deploy-local
deploy-local: ## Build CAS and deploy to cluster via skaffold
	cd $(CAS_ROOT) && skaffold run --filename cloudprovider/azure/examples/dev/skaffold.yaml

.PHONY: deploy-local-dev
deploy-local-dev: ## Build + deploy CAS in watch mode (auto-redeploy on changes)
	cd $(CAS_ROOT) && skaffold dev --filename cloudprovider/azure/examples/dev/skaffold.yaml

##@ E2E Testing

.PHONY: e2etests
e2etests: ## Run e2e tests (CAS must already be deployed)
	go run github.com/onsi/ginkgo/v2/ginkgo \
		--tags e2e \
		-v --trace \
		--timeout $(TEST_TIMEOUT) \
		--output-dir "$(ARTIFACTS)" \
		--junit-report="junit.e2e_suite.1.xml" \
		$(if $(FOCUS),--focus="$(FOCUS)",) \
		$(if $(LABEL_FILTER),--label-filter="$(LABEL_FILTER)",) \
		$(SUITE_PATH) -- \
		-resource-group="$(AZURE_RESOURCE_GROUP)"

##@ Utilities

.PHONY: list-suites
list-suites: ## List available test suites
	@find suites -mindepth 1 -maxdepth 1 -type d -printf '%f\n' 2>/dev/null || echo "No suites found."

.PHONY: validate-env
validate-env: ## Check required environment variables
	@missing=""; \
	for var in AZURE_SUBSCRIPTION_ID AZURE_RESOURCE_GROUP; do \
		eval val=\$$$$var; \
		if [ -z "$$val" ]; then missing="$$missing $$var"; fi; \
	done; \
	if [ -n "$$missing" ]; then \
		echo "ERROR: Missing required environment variables:$$missing"; \
		exit 1; \
	fi; \
	echo "All required environment variables are set."
