ALL_ARCH = amd64 arm64 s390x
all: $(addprefix build-arch-,$(ALL_ARCH))

TAG?=dev
FLAGS=
LDFLAGS?=-s
ENVVAR=CGO_ENABLED=0
GOOS?=linux
GOARCH?=$(shell go env GOARCH)
REGISTRY?=staging-k8s.gcr.io
DOCKER_NETWORK?=default
ifdef BUILD_TAGS
  TAGS_FLAG=--tags ${BUILD_TAGS}
  PROVIDER=-${BUILD_TAGS}
  FOR_PROVIDER=" for ${BUILD_TAGS}"
else
  TAGS_FLAG=
  PROVIDER=
  FOR_PROVIDER=
endif
ifdef LDFLAGS
  LDFLAGS_FLAG=--ldflags "${LDFLAGS}"
else
  LDFLAGS_FLAG=
endif
ifdef DOCKER_RM
  RM_FLAG=--rm
else
  RM_FLAG=
endif
ifndef AWS_REGION
  AWS_REGION=$(shell aws configure get region)
endif

IMAGE=$(REGISTRY)/cluster-autoscaler$(PROVIDER)

export DOCKER_CLI_EXPERIMENTAL := enabled

build:
	@$(MAKE) build-arch-$(GOARCH)

build-arch-%: clean-arch-%
	$(ENVVAR) GOOS=$(GOOS) GOARCH=$* go build -o cluster-autoscaler-$* ${LDFLAGS_FLAG} ${TAGS_FLAG}

test-unit: clean build
	go test --test.short -race ./... ${TAGS_FLAG}

dev-release: dev-release-arch-$(GOARCH)

dev-release-arch-%: build-arch-% make-image-arch-% push-image-arch-%
	@echo "Release ${TAG}${FOR_PROVIDER}-$* completed"

make-image: make-image-arch-$(GOARCH)

make-image-arch-%:
ifdef BASEIMAGE
	docker build --pull --build-arg BASEIMAGE=${BASEIMAGE} \
		-t ${IMAGE}-$*:${TAG} \
		-f Dockerfile.$* .
else
	docker build --pull \
		-t ${IMAGE}-$*:${TAG} \
		-f Dockerfile.$* .
endif
	@echo "Image ${TAG}${FOR_PROVIDER}-$* completed"

push-image: push-image-arch-$(GOARCH)

push-image-arch-%:
	./push_image.sh ${IMAGE}-$*:${TAG}

push-manifest:
	docker manifest create ${IMAGE}:${TAG} \
	    $(addprefix $(REGISTRY)/cluster-autoscaler$(PROVIDER)-, $(addsuffix :$(TAG), $(ALL_ARCH)))
	docker manifest push --purge ${IMAGE}:${TAG}

execute-release: $(addprefix make-image-arch-,$(ALL_ARCH)) $(addprefix push-image-arch-,$(ALL_ARCH)) push-manifest
	@echo "Release ${TAG}${FOR_PROVIDER} completed"

clean: clean-arch-$(GOARCH)

clean-arch-%:
	rm -f cluster-autoscaler-$*

generate:
	AWS_REGION=$(AWS_REGION) go generate ./cloudprovider/aws

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

docker-builder:
	docker build --network=${DOCKER_NETWORK} -t autoscaling-builder ../builder

build-in-docker: build-in-docker-arch-$(GOARCH)

build-in-docker-arch-%: clean-arch-% docker-builder
	docker run ${RM_FLAG} -v `pwd`:/gopath/src/k8s.io/autoscaler/cluster-autoscaler/:Z autoscaling-builder:latest \
		bash -c 'cd /gopath/src/k8s.io/autoscaler/cluster-autoscaler && BUILD_TAGS=${BUILD_TAGS} LDFLAGS="${LDFLAGS}" make build-arch-$*'

release: $(addprefix build-in-docker-arch-,$(ALL_ARCH)) execute-release
	@echo "Full in-docker release ${TAG}${FOR_PROVIDER} completed"

container: container-arch-$(GOARCH)

container-arch-%: build-in-docker-arch-% make-image-arch-%
	@echo "Full in-docker image ${TAG}${FOR_PROVIDER}-$* completed"

test-in-docker: clean docker-builder
	docker run ${RM_FLAG} -v `pwd`:/cluster-autoscaler/:Z autoscaling-builder:latest bash -c 'cd /cluster-autoscaler && go test -race ./... ${TAGS_FLAG}'

.PHONY: all build test-unit clean format execute-release dev-release docker-builder build-in-docker release generate push-image push-manifest

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.14.0

include hack/tools.mk

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: manifest
manifest: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./apis/..." output:crd:artifacts:config=apis/config/crd

.PHONY: start
start:
	@GO111MODULE=on go run main.go \
			--kubeconfig=$(TARGET_KUBECONFIG) \
			--cloud-provider=mcm \
			--nodes=0:2:$(MACHINE_DEPLOYMENT_1_ZONE_1) \
			--nodes=1:2:$(MACHINE_DEPLOYMENT_2_ZONE_1) \
			--nodes=0:1:$(MACHINE_DEPLOYMENT_2_ZONE_2) \
			--nodes=0:1:$(MACHINE_DEPLOYMENT_2_ZONE_3) \
			--skip-nodes-with-system-pods=false \
			--skip-nodes-with-local-storage=false \
			--scale-down-delay-after-add=10s \
			--scale-down-delay-after-failure=2s \
			--v=4 \
			--ignore-taint=node.gardener.cloud/critical-components-not-ready \
			--ignore-taint=testing.node.gardener.cloud/initial-node-blocked \
			--expander=least-waste \
			--scale-down-unneeded-time=5s \
			--balance-similar-node-groups=true \
			--max-node-provision-time=4m \
			--leader-elect=$(LEADER_ELECT) \
			--leader-elect-retry-period="20s" \
			--leader-elect-renew-deadline="30s" \
			--leader-elect-lease-duration="40s"


.PHONY: download-kubeconfigs
download-kubeconfigs:
	@chmod +x ./hack/local_setup.sh
	@echo "enter project name"; \
	read PROJECT; \
	echo "enter seed name"; \
	read SEED; \
	echo "enter shoot name"; \
	read SHOOT; \
	echo "enter cluster provider(gcp|aws|azure|vsphere|openstack|alicloud|metal|equinix-metal)"; \
	read PROVIDER; \
	echo "enter zone with zero nodes in worker pool \"three-zones\" where the PV needs to be created to perform test"; \
	read ZONE; \
	. ./hack/local_setup.sh --SEED $$SEED --SHOOT $$SHOOT --PROJECT $$PROJECT; \
	echo "###############################################" ; \
	echo "To run the integration test run the following commands"; \
	echo " ";\
	echo "export CONTROL_NAMESPACE=$$CONTROL_NAMESPACE"; \
	echo "export CONTROL_KUBECONFIG=$$CONTROL_KUBECONFIG"; \
	echo "export TARGET_KUBECONFIG=$$TARGET_KUBECONFIG"; \
	echo "export CLUSTER_PROVIDER=$$PROVIDER"; \
	echo "export VOLUME_ZONE=$$ZONE"; \
	echo " "; \
	echo "make test-integration"

.PHONY: test-integration
test-integration:
	../.ci/local_integration_test

.PHONY: sast
sast: $(GOSEC)
	@./hack/sast.sh

.PHONY: sast-report
sast-report: $(GOSEC)
	@./hack/sast.sh --gosec-report true
