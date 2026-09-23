REPO_ROOT = $(shell git rev-parse --show-toplevel)

## Tool Binaries
CONTROLLER_GEN = go tool sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: manifests
manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=./config/crd

.PHONY: generate
generate:
	$(CONTROLLER_GEN) object:headerFile="$(REPO_ROOT)/hack/boilerplate/boilerplate.generatego.txt" paths="./..."

.PHONY: clients
clients:
	./hack/update-codegen.sh

.PHONY: all
all: manifests generate clients
