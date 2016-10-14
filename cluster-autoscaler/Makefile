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

release: build
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
	docker build -t ${REGISTRY}/cluster-autoscaler:${TAG} .
	gcloud docker push ${REGISTRY}/cluster-autoscaler:${TAG}

clean:
	rm -f cluster-autoscaler

.PHONY: all deps build test-unit clean release
