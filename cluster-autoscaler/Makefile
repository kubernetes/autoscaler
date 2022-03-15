ALL_ARCH = amd64 arm64
all: $(addprefix build-arch-,$(ALL_ARCH))

TAG?=dev
FLAGS=
LDFLAGS?=-s
ENVVAR=CGO_ENABLED=0
GOOS?=linux
GOARCH?=$(shell go env GOARCH)
REGISTRY?=registry.hub.docker.com
DOCKER_NETWORK?=default
ifdef BUILD_TAGS
  TAGS_FLAG=--tags ${BUILD_TAGS}
else
  TAGS_FLAG=
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
IMAGE=$(REGISTRY)/phongvu0403/cluster-autoscaler

export DOCKER_CLI_EXPERIMENTAL := enabled

build: build-arch-$(GOARCH)

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
	@echo "Image ${TAG}-$* completed"

push-image: push-image-arch-$(GOARCH)

push-image-arch-%:
	./push_image.sh ${IMAGE}-$*:${TAG}

push-manifest:
	docker manifest create ${IMAGE}:${TAG} \
	    $(addprefix $(REGISTRY)/cluster-autoscaler$(PROVIDER)-, $(addsuffix :$(TAG), $(ALL_ARCH)))
	docker manifest push --purge ${IMAGE}:${TAG}

execute-release: $(addprefix make-image-arch-,$(ALL_ARCH)) $(addprefix push-image-arch-,$(ALL_ARCH))
	@echo "Release ${TAG} completed"

clean: clean-arch-$(GOARCH)

clean-arch-%:
	rm -f cluster-autoscaler-$*

generate:
	go generate ./cloudprovider/aws

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
