all: build

FLAGS=
OFFICIAL_NAME=balancer

build: clean
	go build .

build-linux-amd64: clean
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .

test-unit: clean
	go test --test.short -race ./... $(FLAGS)

docker-build:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/${OFFICIAL_NAME}:${TAG} .

docker-push:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker push ${REGISTRY}/${OFFICIAL_NAME}:${TAG}

docker-builder:
	docker build -t vpa-autoscaling-builder ../vertical-pod-autoscaler/builder

build-in-docker: clean docker-builder
	docker run -v `pwd`/..:/gopath/src/k8s.io/autoscaler vpa-autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/balancer && make build-linux-amd64'

build-image-in-docker: build-in-docker docker-build

test-in-docker: build-in-docker
	docker run -v `pwd`/..:/gopath/src/k8s.io/autoscaler vpa-autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/balancer && make test-unit'

release: build-image-in-docker docker-push
	@echo "Full in-docker release ${OFFICIAL_NAME}:${TAG} completed"

clean:
	rm -f balancer

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all build test-unit clean format release

