ALL_ARCH = amd64 arm64
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
IMAGE=$(REGISTRY)/ca-external-grpc-cloud-provider$(PROVIDER)

export DOCKER_CLI_EXPERIMENTAL := enabled

build: build-arch-$(GOARCH)

build-arch-%: clean-arch-%
	$(ENVVAR) GOOS=$(GOOS) GOARCH=$* go build -o ca-external-grpc-cloud-provider-$* ${LDFLAGS_FLAG} ${TAGS_FLAG}

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

clean: clean-arch-$(GOARCH)

clean-arch-%:
	rm -f ca-external-grpc-cloud-provider-$*

docker-builder:
	docker build --network=${DOCKER_NETWORK} -t autoscaling-builder ../../../../../builder

build-in-docker: build-in-docker-arch-$(GOARCH)

build-in-docker-arch-%: clean-arch-% docker-builder
	docker run ${RM_FLAG} -v `pwd`/../../../../:/gopath/src/k8s.io/autoscaler/cluster-autoscaler/:Z autoscaling-builder:latest \
		bash -c 'cd /gopath/src/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/examples/external-grpc-cloud-provider-service && BUILD_TAGS=${BUILD_TAGS} LDFLAGS="${LDFLAGS}" make build-arch-$*'

container: container-arch-$(GOARCH)

container-arch-%: build-in-docker-arch-% make-image-arch-%
	@echo "Full in-docker image ${TAG}${FOR_PROVIDER}-$* completed"

.PHONY: all build clean docker-builder build-in-docker
