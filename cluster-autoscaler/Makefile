ALL_ARCH = amd64 arm64 s390x
all: $(addprefix build-arch-,$(ALL_ARCH))

TAG?=dev
FLAGS=
LDFLAGS?=-s
ENVVAR=CGO_ENABLED=0
GOOS?=linux
GOARCH?=$(shell go env GOARCH)
REGISTRY?=gcr.io/k8s-staging-autoscaling
DOCKER_NETWORK?=default
SUPPORTED_BUILD_TAGS=$(shell ls cloudprovider/builder/ | grep -e '^builder_.*\.go' | sed 's/builder_\(.*\)\.go/\1/')
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
  LDFLAGS_FLAG=--ldflags="${LDFLAGS}"
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

test-build-tags:
	@if [ -z "$(SUPPORTED_BUILD_TAGS)" ]; then \
		echo "No supported build tags found"; \
		exit 1; \
	fi
	@for tag in $(SUPPORTED_BUILD_TAGS); do \
		echo "Testing build with tag $$tag"; \
		BUILD_TAGS=$$tag $(MAKE) build || exit 1; \
	done

test-unit: clean build
	go test --test.short -race ./... ${TAGS_FLAG}

dev-release: dev-release-arch-$(GOARCH)

dev-release-arch-%: build-arch-% make-image-arch-% push-image-arch-%
	@echo "Release ${TAG}${FOR_PROVIDER}-$* completed"

make-image: make-image-arch-$(GOARCH)

make-image-arch-%:
	GOOS=$(GOOS) docker buildx build --pull --platform linux/$* \
		--build-arg "GOARCH=$*" \
		-t ${IMAGE}-$*:${TAG} \
		-f Dockerfile .
	@echo "Image ${TAG}${FOR_PROVIDER}-$* completed"

push-image: push-image-arch-$(GOARCH)

push-image-arch-%:
	./push_image.sh ${IMAGE}-$*:${TAG}

push-release-image-arch-%:
	docker push ${IMAGE}-$*:${TAG}

push-manifest:
	docker manifest create ${IMAGE}:${TAG} \
	    $(addprefix $(REGISTRY)/cluster-autoscaler$(PROVIDER)-, $(addsuffix :$(TAG), $(ALL_ARCH)))
	docker manifest push --purge ${IMAGE}:${TAG}

execute-release: $(addprefix make-image-arch-,$(ALL_ARCH)) $(addprefix push-release-image-arch-,$(ALL_ARCH)) push-manifest
	@echo "Release ${TAG}${FOR_PROVIDER} completed"

ROOT_DIR := $(shell git rev-parse --show-toplevel 2>/dev/null)
GIT_REMOTE_NAME ?= upstream
RELEASE_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
RELEASE_TAG ?= $(shell git describe --abbrev=0 2>/dev/null)
FULL_VERSION := $(RELEASE_TAG:cluster-autoscaler-%=%)
MAJOR_VERSION := $(shell echo $(FULL_VERSION) | sed -E 's/^([0-9]+)\.([0-9]+)\.([0-9]+)$$/\1/')
MINOR_VERSION := $(shell echo $(FULL_VERSION) | sed -E 's/^([0-9]+)\.([0-9]+)\.([0-9]+)$$/\2/')
PATCH_VERSION := $(shell echo $(FULL_VERSION) | sed -E 's/^([0-9]+)\.([0-9]+)\.([0-9]+)$$/\3/')
# if the release tag is a .0 version, use the main branch
ifeq ($(PATCH_VERSION),0)
	PREVIOUS_TAG ?= $(shell git tag -l | grep -E "^cluster-autoscaler-[0-9]+\.[0-9]+\.0$$" | sort -V | tail -n 2 | head -n 1 2>/dev/null)
else
	PREVIOUS_TAG ?= $(shell git tag -l | grep -E "^cluster-autoscaler-[0-9]+\.[0-9]+\.[0-9]+$$" | sort -V | tail -n 2 | head -n 1 2>/dev/null)
endif
GIT_REMOTE_NAME ?= upstream
GIT_ORG_NAME ?= kubernetes
GIT_REPO_NAME ?= autoscaler
START_SHA ?= $(shell git rev-list -n 1 $(PREVIOUS_TAG) 2>/dev/null)
END_SHA ?= $(shell git rev-parse $(GIT_REMOTE_NAME)/$(RELEASE_BRANCH) 2>/dev/null)

thing:
	@echo $(FULL_VERSION)
	@echo $(MAJOR_VERSION)
	@echo $(MINOR_VERSION)
	@echo $(PATCH_VERSION)
	@echo $(PREVIOUS_TAG)
	@echo $(START_SHA)
	@echo $(END_SHA)
	@echo $(ROOT_DIR)

$(RELEASE_NOTES): ## Install release notes to GOBIN
	go install k8s.io/release/cmd/release-notes@v0.18.0

.PHONY: release-notes
release-notes: $(RELEASE_NOTES)
	@echo "generating release notes from $(PREVIOUS_TAG) to $(RELEASE_TAG) with start sha $(START_SHA) and end sha $(END_SHA)"
	release-notes --org $(GIT_ORG_NAME) --repo $(GIT_REPO_NAME) --branch $(RELEASE_BRANCH) --repo-path $(ROOT_DIR) --start-sha $(START_SHA) --end-sha $(END_SHA) --markdown-links true --output $(RELEASE_TAG).md --list-v2

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

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: manifest
manifest: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./apis/..." output:crd:artifacts:config=apis/config/crd
