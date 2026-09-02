
UPCLOUD_SDK_PACKAGE ?= github.com/upcloudltd/upcloud-go-api/v8
UPCLOUD_SDK_VERSION ?= v8.35.0

IMAGE_TAG ?= cluster-autoscaler-upcloud:dev

build:
	BUILD_TAGS=upcloud make -C ../../ build-in-docker

image:
	BUILD_TAGS=upcloud docker build -t $(IMAGE_TAG) -f ../../Dockerfile ../../

vendor:
	sh vendor.sh $(UPCLOUD_SDK_PACKAGE) $(UPCLOUD_SDK_VERSION) 

test:
	go test -v -trace k8s.io/autoscaler/cloudprovider/upcloud

lint:
	golangci-lint run
