all: build

TAG?=dev
FLAGS=
ENVVAR=
GOOS?=linux
ifdef BUILD_TAGS
	PROVIDER=-${BUILD_TAGS}
	FOR_PROVIDERS=" for ${BUILD_TAGS}"
endif


deps:
	go get github.com/tools/godep

build: clean deps
	$(ENVVAR) GOOS=$(GOOS) godep go build ./...
	$(ENVVAR) GOOS=$(GOOS) godep go build -o cluster-autoscaler

build-binary: clean deps
	$(ENVVAR) GOOS=$(GOOS) godep go build -tags '${BUILD_TAGS}' -o cluster-autoscaler

test-unit: clean deps build
	$(ENVVAR) godep go test --test.short -race ./... $(FLAGS)

dev-release: build-binary execute-release
	@echo "Release ${TAG}${FOR_PROVIDERS} completed"

execute-release:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/cluster-autoscaler${PROVIDER}:${TAG} .
	./push_image.sh ${REGISTRY}/cluster-autoscaler${PROVIDER}:${TAG}

clean:
	rm -f cluster-autoscaler

generate:
	go generate ./cloudprovider/aws

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
    test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

docker-builder:
	docker build -t autoscaling-builder ../builder

build-in-docker: clean docker-builder
	docker run -v `pwd`:/gopath/src/k8s.io/autoscaler/cluster-autoscaler/ autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/cluster-autoscaler && make build-binary'

release: build-in-docker execute-release
	@echo "Full in-docker release ${TAG} completed"

test-in-docker: clean docker-builder
	docker run -v `pwd`:/gopath/src/k8s.io/autoscaler/cluster-autoscaler/ autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/cluster-autoscaler && godep go test ./... '

.PHONY: all deps build test-unit clean format execute-release dev-release docker-builder build-in-docker release generate

