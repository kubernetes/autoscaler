REPO_ROOT:=$(shell git rev-parse --show-toplevel)
CAS_ROOT:=$(REPO_ROOT)/cluster-autoscaler

include $(CAS_ROOT)/Makefile

CLUSTER_AUTOSCALER_NAMESPACE?=default
CLUSTER_AUTOSCALER_SERVICEACCOUNT_NAME?=cluster-autoscaler

.PHONY: build-e2e
build-e2e:
	$(MAKE) -C $(CAS_ROOT) build-arch-$(GOARCH) make-image-arch-$(GOARCH)
	docker push $(IMAGE)-$(GOARCH):$(TAG)

ARTIFACTS?=_artifacts

.PHONY: test-e2e
test-e2e: build-e2e
	go run github.com/onsi/ginkgo/v2/ginkgo --tags e2e -v --trace --output-dir "$(ARTIFACTS)" --junit-report="junit.e2e_suite.1.xml" e2e -- \
		-resource-group="$$(KUBECONFIG= kubectl get managedclusters -n default -o jsonpath='{.items[0].status.nodeResourceGroup}')" \
		-cluster-name="$$(KUBECONFIG= kubectl get cluster -n default -o jsonpath='{.items[0].metadata.name}')" \
		-client-id="$$(KUBECONFIG= kubectl get userassignedidentities -n default -o jsonpath='{.items[0].status.clientId}')" \
		-cas-namespace="$(CLUSTER_AUTOSCALER_NAMESPACE)" \
		-cas-serviceaccount-name="$(CLUSTER_AUTOSCALER_SERVICEACCOUNT_NAME)" \
		-cas-image-repository="$(IMAGE)-$(GOARCH)" \
		-cas-image-tag="$(TAG)"
