all: build

TAG?=dev
FLAGS=
ENVVAR=
GOOS?=linux

deps:
	go get github.com/tools/godep

build: clean deps
	$(ENVVAR) GOOS=$(GOOS) godep go build ./...
	$(ENVVAR) GOOS=$(GOOS) godep go build -o cluster-autoscaler

test-unit: clean deps build
	$(ENVVAR) godep go test --test.short -race ./... $(FLAGS)

dev-release: build execute-release
	echo "Release ${TAG} completed"

execute-release:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/cluster-autoscaler:${TAG} .
	gcloud docker -- push ${REGISTRY}/cluster-autoscaler:${TAG}

clean:
	rm -f cluster-autoscaler

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
    test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

docker-builder:
	docker build -t ca-builder ./builder

build-in-docker: clean docker-builder
	docker run -v `pwd`:/gopath/src/k8s.io/autoscaler/cluster-autoscaler/  ca-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/cluster-autoscaler && make'

release: build-in-docker execute-release
	echo "Full in-docker release ${TAG} completed"

.PHONY: all deps build test-unit clean format execute-release dev-release docker-builder build-in-docker release

