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

REGISTRY ?= us-central1-docker.pkg.dev/k8s-staging-images/autoscaling
HELM ?= helm
HELM_CHART ?= cluster-autoscaler
HELM_CHART_DIR ?= $(HELM_CHART)
HELM_CHART_REPO ?= $(REGISTRY)/charts
OUTPUT_DIR ?= _output
GOBIN ?= $(if $(shell go env GOBIN),$(shell go env GOBIN),$(shell go env GOPATH)/bin)
export PATH := $(GOBIN):$(PATH)

.PHONY: helm
helm:
	@command -v $(HELM) >/dev/null 2>&1 || GOBIN=$(GOBIN) go install helm.sh/helm/v3/cmd/helm@latest

.PHONY: package-helm
package-helm: helm
	mkdir -p $(OUTPUT_DIR)
	$(HELM) package $(HELM_CHART_DIR) -d $(OUTPUT_DIR)

.PHONY: push-helm
push-helm: package-helm
	chart_name=$$(grep -E '^name: ' $(HELM_CHART_DIR)/Chart.yaml | cut -d' ' -f2); \
	chart_version=$$(grep -E '^version: ' $(HELM_CHART_DIR)/Chart.yaml | cut -d' ' -f2); \
	$(HELM) push $(OUTPUT_DIR)/$${chart_name}-$${chart_version}.tgz oci://$(HELM_CHART_REPO)

.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)
