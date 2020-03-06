all: build

TAG?=dev
REGISTRY?=staging-k8s.gcr.io
FLAGS=
TEST_ENVVAR=LD_FLAGS=-s GO111MODULE=on
ENVVAR=CGO_ENABLED=0 $(TEST_ENVVAR)
GOOS?=linux
COMPONENT=admission-controller
FULL_COMPONENT=vpa-${COMPONENT}

build: clean
	$(ENVVAR) GOOS=$(GOOS) go build ./...
	$(ENVVAR) GOOS=$(GOOS) go build -o ${COMPONENT}

build-binary: clean
	$(ENVVAR) GOOS=$(GOOS) go build -o ${COMPONENT}

build-binary-with-vendor: clean
	$(ENVVAR) GOOS=$(GOOS) go build -mod vendor -o ${COMPONENT}

test-unit: clean build
	$(TEST_ENVVAR) go test --test.short -race ./... $(FLAGS)

docker-build:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/${FULL_COMPONENT}:${TAG} .

docker-push:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker push ${REGISTRY}/${FULL_COMPONENT}:${TAG}

docker-builder:
	docker build -t vpa-autoscaling-builder ../../builder

build-in-docker: clean docker-builder
	docker run -v `pwd`/../..:/gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler vpa-autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler && make build-binary-with-vendor -C pkg/admission-controller'

release: build-in-docker docker-build docker-push
	@echo "Full in-docker release ${FULL_COMPONENT}:${TAG} completed"

clean:
	rm -f ${COMPONENT}

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all build test-unit clean format release
