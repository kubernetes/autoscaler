# Copyright The Kubernetes Authors.
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

export GOBIN ?= $(CURDIR)/bin
export PATH := $(GOBIN):$(PATH)

# NOTE: To work with this Makefile, ensure you run 'make install-dependencies' first.
# This will install all required tools into the local bin/ directory.

.PHONY: install-dependencies
install-dependencies: HELM_VERSION ?= v3.21.1
install-dependencies: KIND_VERSION ?= v0.32.0
install-dependencies: ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}') # Pin the version to controller-runtime's one.
install-dependencies:
	@rm -Rf bin && mkdir -p $(GOBIN)
	go install helm.sh/helm/v3/cmd/helm@$(HELM_VERSION)
	go install sigs.k8s.io/kind@$(KIND_VERSION)
	go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION)
	@$(MAKE) setup-envtest

.PHONY: setup-envtest
setup-envtest: ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
setup-envtest:
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@setup-envtest use $(ENVTEST_K8S_VERSION) --bin-dir $(GOBIN) -p path || { \
		echo "Warning: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		echo "Attempting fallback to latest available envtest binaries version."; \
		setup-envtest use latest --bin-dir $(GOBIN) -p path || { \
			echo "Error: Failed to set up envtest binaries."; \
			exit 1; \
		}; \
	}
	@chmod -R +w $(GOBIN)/k8s

.PHONY: build-kwok
build-kwok:
	@CGO_ENABLED=0 GOOS=linux go build -o cluster-autoscaler-kwok ./kwok

.PHONY: image-kwok
image-kwok: TAG?=dev
image-kwok: build-kwok
	@docker build -t cluster-autoscaler-kwok:${TAG} -f kwok/Dockerfile .

.PHONY: test
test:
	@go test -race ./...

test-controllers:
	@ginkgo --race ./pkg/test/integration/controllers/...

.PHONY: clean
clean:
	@rm -f cluster-autoscaler-kwok

.PHONY: format
format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"


.PHONY: run-e2e
run-e2e: e2e-kwok-cluster e2e-install-ca
	@go test -tags e2e -v ./test/e2e/... -args -v=4
	@$(MAKE) e2e-teardown

E2E_CLUSTER_NAME ?= ca-e2e-kwok

.PHONY: e2e-kwok-cluster
e2e-kwok-cluster: E2E_KIND_CONFIG ?= kind-config.yaml
e2e-kwok-cluster: KWOK_REPO_URL ?= https://kwok.sigs.k8s.io/charts/
e2e-kwok-cluster:
	@kind get clusters 2>/dev/null | grep -q "^$(E2E_CLUSTER_NAME)$$" || kind create cluster --name $(E2E_CLUSTER_NAME) --config $(E2E_KIND_CONFIG)
	helm repo add kwok-charts $(KWOK_REPO_URL)
	helm upgrade --install kwok --namespace kube-system kwok-charts/kwok --set hostNetwork=true --wait
	helm upgrade --install kwok-stage-fast --namespace kube-system kwok-charts/stage-fast --wait
	kubectl apply -f pkg/apis/config/crd/
	kubectl -n kube-system patch ds kindnet -p '{"spec":{"template":{"spec":{"nodeSelector":{"node-role.kubernetes.io/control-plane":""}}}}}'
	kubectl -n kube-system patch ds kube-proxy -p '{"spec":{"template":{"spec":{"nodeSelector":{"node-role.kubernetes.io/control-plane":""}}}}}'

.PHONY: e2e-install-ca
e2e-install-ca: image-kwok
	kind load docker-image cluster-autoscaler-kwok:dev --name $(E2E_CLUSTER_NAME)
	helm upgrade --install cluster-autoscaler --namespace kube-system ./kwok/charts/ --wait \
		--set tolerations[0].key=node-role.kubernetes.io/control-plane \
		--set tolerations[0].operator=Exists \
		--set tolerations[0].effect=NoSchedule \
		--set extraArgs.scan-interval=5s \
		--set extraArgs.scale-down-unneeded-time=10s \
		--set extraArgs.scale-down-unready-time=10s \
		--set extraArgs.scale-down-unready-enabled=true \
		--set extraArgs.scale-down-delay-after-add=10s \
		--set extraArgs.scale-down-delay-after-delete=10s \
		--set extraArgs.scale-down-delay-after-failure=10s \
		--set extraArgs.scale-down-delay-type-local=true \
		--set extraArgs.unremovable-node-recheck-timeout=5s \
		--set extraArgs.max-node-provision-time=15s \
		--set extraArgs.max-node-startup-time=15s \
		--set extraArgs.scale-down-enabled=true \
		--set extraArgs.enable-csi-node-aware-scheduling=false \
		--set extraArgs.expendable-pods-priority-cutoff=-10 \
		--set extraArgs.v=5
	kubectl rollout restart deployment/cluster-autoscaler -n kube-system
	kubectl rollout status deployment/cluster-autoscaler -n kube-system --timeout=60s

.PHONY: e2e-teardown
e2e-teardown:
	kind delete cluster --name $(E2E_CLUSTER_NAME)
