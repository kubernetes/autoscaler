all: build

TAG?=dev
REGISTRY?=staging-k8s.gcr.io
FLAGS=
ENVVAR=
GOOS?=linux
COMPONENT=recommender
FULL_COMPONENT=vpa-${COMPONENT}

deps:
	go get github.com/tools/godep

build: clean deps
	$(ENVVAR) GOOS=$(GOOS) godep go build ./...
	$(ENVVAR) GOOS=$(GOOS) godep go build -o ${COMPONENT}

build-binary: clean deps
	$(ENVVAR) GOOS=$(GOOS) godep go build -o ${COMPONENT}

test-unit: clean deps build
	$(ENVVAR) godep go test --test.short -race ./... $(FLAGS)

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
	docker run -v `pwd`/../..:/gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler vpa-autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler && make build-binary -C pkg/recommender'

release: build-in-docker docker-build docker-push
	@echo "Full in-docker release ${FULL_COMPONENT}:${TAG} completed"

clean:
	rm -f ${COMPONENT}

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all deps build test-unit clean format release
