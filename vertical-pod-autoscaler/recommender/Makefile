all: build

TAG?=0.0.1
REGISTRY?=gcr.io/kubernetes-develop
FLAGS=
ENVVAR=
GOOS?=linux

deps:
	go get github.com/tools/godep

build: clean deps
	$(ENVVAR) GOOS=$(GOOS) godep go build ./...
	$(ENVVAR) GOOS=$(GOOS) godep go build -o recommender

test-unit: clean deps build
	$(ENVVAR) godep go test --test.short -race ./... $(FLAGS)

docker:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/recommender:${TAG} .
	gcloud docker -- push ${REGISTRY}/recommender:${TAG}

release: build docker

clean:
	rm -f recommender

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all deps build test-unit clean format release
