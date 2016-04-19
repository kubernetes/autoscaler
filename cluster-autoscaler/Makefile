all: build

TAG = v0.0.1
PREFIX = gcr.io/google_containers
FLAGS = 
ENVVAR=

deps:
	go get github.com/tools/godep

build: clean deps
	$(ENVVAR) godep go build ./...
	$(ENVVAR) godep go build -o cluster-autoscaler 

test-unit: clean deps build
	$(ENVVAR) godep go test --test.short -race ./... $(FLAGS)

clean:
	rm -f clusterautoscaler

.PHONY: all deps build test-unit clean
