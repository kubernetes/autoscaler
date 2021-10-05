all: build

TAG?=dev
REGISTRY?=staging-k8s.gcr.io
FLAGS=
TEST_ENVVAR=LD_FLAGS=-s GO111MODULE=on
ENVVAR=CGO_ENABLED=0 $(TEST_ENVVAR)
GOOS?=linux
COMPONENT=admission-controller
FULL_COMPONENT=vpa-${COMPONENT}

ALL_ARCHITECTURES?=amd64 arm arm64 ppc64le s390x
export DOCKER_CLI_EXPERIMENTAL=enabled

build: clean
	$(ENVVAR) GOOS=$(GOOS) go build ./...
	$(ENVVAR) GOOS=$(GOOS) go build -o ${COMPONENT}

build-binary: clean
	$(ENVVAR) GOOS=$(GOOS) go build -o ${COMPONENT}

.PHONY: build-binary-with-vendor
build-binary-with-vendor: $(addprefix build-binary-with-vendor-,$(ALL_ARCHITECTURES)) clean

.PHONY: build-binary-with-vendor-*
build-binary-with-vendor-%:
	$(ENVVAR) GOARCH=$* GOOS=$(GOOS) go build -mod vendor -o ${COMPONENT}-$*

test-unit: clean build
	$(TEST_ENVVAR) go test --test.short -race ./... $(FLAGS)

.PHONY: docker-build
docker-build: $(addprefix docker-build-,$(ALL_ARCHITECTURES))

.PHONY: docker-build-*
docker-build-%: 
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/${FULL_COMPONENT}-$*:${TAG} --build-arg ARCH=$* .

.PHONY: docker-push
docker-push: $(addprefix sub-push-,$(ALL_ARCHITECTURES)) push-multi-arch;

.PHONY: sub-push-*
sub-push-%: docker-build-% do-push-% ;

.PHONY: do-push-*
do-push-%:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker push ${REGISTRY}/${FULL_COMPONENT}-$*:${TAG}

.PHONY: push-multi-arch
push-multi-arch:
	docker manifest create --amend $(REGISTRY)/${FULL_COMPONENT}:$(TAG) $(shell echo $(ALL_ARCHITECTURES) | sed -e "s~[^ ]*~$(REGISTRY)/${FULL_COMPONENT}\-&:$(TAG)~g")
	@for arch in $(ALL_ARCHITECTURES); do docker manifest annotate --arch $${arch} $(REGISTRY)/${FULL_COMPONENT}:$(TAG) $(REGISTRY)/${FULL_COMPONENT}-$${arch}:${TAG}; done
	docker manifest push --purge $(REGISTRY)/${FULL_COMPONENT}:$(TAG)

docker-builder:
	docker build -t vpa-autoscaling-builder ../../builder

.PHONY: build-in-docker
build-in-docker: $(addprefix build-in-docker-,$(ALL_ARCHITECTURES))

.PHONY: build-in-docker-*
build-in-docker-%: clean docker-builder
	docker run -v `pwd`/../..:/gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler vpa-autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler && make build-binary-with-vendor-$* -C pkg/admission-controller'


.PHONY: release
release: build-in-docker docker-build docker-push
	@echo "Full in-docker release ${FULL_COMPONENT}:${TAG} completed"

clean: $(addprefix clean-,$(ALL_ARCHITECTURES))

clean-%:
	rm -f ${COMPONENT}-$*

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all build test-unit clean format release
