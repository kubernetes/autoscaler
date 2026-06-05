ALL_ARCH = amd64 arm64 s390x
all: $(addprefix build-arch-,$(ALL_ARCH))

# TODO: #8127 - Use default analyzers set by `go test` to include `printf` analyzer.
# Default analyzers that go test runs according to https://github.com/golang/go/blob/52624e533fe52329da5ba6ebb9c37712048168e0/src/cmd/go/internal/test/test.go#L649
# This doesn't include the `printf` analyzer until cluster-autoscaler libraries are updated.
GO_TEST_DEFAULT_ANALYZERS?=atomic,bool,buildtags,directive,errorsas,ifaceassert,nilfunc,slog,stringintconv,tests
TAG?=dev
FLAGS=
LDFLAGS?=-s
ENVVAR=CGO_ENABLED=0
GOOS?=linux
GOARCH?=$(shell go env GOARCH)
REGISTRY?=gcr.io/k8s-staging-autoscaling
DOCKER_NETWORK?=default
SUPPORTED_BUILD_TAGS=$(shell ls cloudprovider/router/ | grep -e '^router_.*\.go' | sed 's/router_\(.*\)\.go/\1/')

# VERSION is the cluster-autoscaler version baked into the binary via -ldflags.
#   1. Exact git tag pointing at HEAD, with the "cluster-autoscaler-" prefix stripped.
#   2. The current commit SHA.
#   3. The literal string "dev" (for builds outside a git checkout).
# A "-dirty" suffix is appended when the git working tree has uncommitted
# changes (cases 1 and 2 only).
# Override with `make VERSION=<value> ...` to force a specific value; an
# externally provided VERSION is used verbatim, with no "-dirty" suffix.
ifeq ($(origin VERSION),undefined)
  VERSION := $(shell git describe --exact-match --tags 2>/dev/null | sed -e 's|^cluster-autoscaler-||')
  ifeq ($(strip $(VERSION)),)
    VERSION := $(shell git rev-parse HEAD 2>/dev/null)
  endif
  ifeq ($(strip $(VERSION)),)
    VERSION := dev
  else
    VERSION := $(VERSION)$(shell git diff --no-ext-diff --quiet --exit-code 2>/dev/null || echo -dirty)
  endif
endif

VERSION_PKG := k8s.io/autoscaler/cluster-autoscaler/version
VERSION_LDFLAG := -X $(VERSION_PKG).ClusterAutoscalerVersion=$(VERSION)
LDFLAGS_VALUE := $(strip $(LDFLAGS) $(VERSION_LDFLAG))

ifdef BUILD_TAGS
  TAGS_FLAG=--tags ${BUILD_TAGS}
  PROVIDER=-${BUILD_TAGS}
  FOR_PROVIDER=" for ${BUILD_TAGS}"
else
  TAGS_FLAG=
  PROVIDER=
  FOR_PROVIDER=
endif
LDFLAGS_FLAG=--ldflags="${LDFLAGS_VALUE}"
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
	go test -race ./... -vet="${GO_TEST_DEFAULT_ANALYZERS}" ${TAGS_FLAG}

test-controllers: setup-envtest clean build
	ginkgo --race --vet="${GO_TEST_DEFAULT_ANALYZERS}" ${TAGS_FLAG} ./test/integration/controllers/...

test-ci: setup-envtest
	go test -race ./... -vet="${GO_TEST_DEFAULT_ANALYZERS}" ${TAGS_FLAG}

benchmark:
	go test ./... -bench=. -run='^$$' -vet="${GO_TEST_DEFAULT_ANALYZERS}"

dev-release: dev-release-arch-$(GOARCH)

dev-release-arch-%: build-arch-% make-image-arch-% push-image-arch-%
	@echo "Release ${TAG}${FOR_PROVIDER}-$* completed"

make-image: make-image-arch-$(GOARCH)

make-image-arch-%:
	GOOS=$(GOOS) docker buildx build --pull --platform linux/$* \
		--build-arg "GOARCH=$*" \
		--build-arg 'LDFLAGS=${LDFLAGS_VALUE}' \
		--build-arg 'BUILD_TAGS=${BUILD_TAGS}' \
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

release-extract-version = $(VERSION)

release-validate:
	@if [ -z "$(shell git tag --points-at HEAD | grep -Fx 'cluster-autoscaler-$(VERSION)')" ]; then \
		echo "Can't release: VERSION=$(VERSION) does not match a git tag at HEAD"; \
		exit 1; \
	fi

release: TAG=v$(call release-extract-version)
release: release-validate $(addprefix build-in-docker-arch-,$(ALL_ARCH)) execute-release
	@echo "Full in-docker release ${TAG}${FOR_PROVIDER} completed"

container: container-arch-$(GOARCH)

container-arch-%: build-in-docker-arch-% make-image-arch-%
	@echo "Full in-docker image ${TAG}${FOR_PROVIDER}-$* completed"

test-in-docker: clean docker-builder
	docker run ${RM_FLAG} -v `pwd`:/cluster-autoscaler/:Z autoscaling-builder:latest bash -c 'cd /cluster-autoscaler && go test -race ./... -vet="${GO_TEST_DEFAULT_ANALYZERS}" ${TAGS_FLAG}'

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ENVTEST ?= $(LOCALBIN)/setup-envtest
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')


.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

.PHONY: all build test-unit clean format execute-release dev-release docker-builder build-in-docker release generate push-image push-manifest
