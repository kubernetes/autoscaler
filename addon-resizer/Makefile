# Copyright 2016 The Kubernetes Authors. All rights reserved#
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License..

# Usage:
# Build docker image for all the architecture - make all
# Build and push images for all the architecture - make all-push
# Build for specific architecture(ppc64le) - make container ARCH=ppc64le

ALL_ARCH = amd64 arm arm64 ppc64le s390x
ARCH ?= amd64

GOARM=7
GOLANG_VERSION = 1.12.1
REGISTRY = staging-k8s.gcr.io
IMGNAME = addon-resizer
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)
TAG = 2.3

BASEIMAGE?=gcr.io/distroless/static:latest

TEMP_DIR := $(shell mktemp -d)

all: all-container

sub-container-%:
	$(MAKE) ARCH=$* container

sub-push-%:
	$(MAKE) ARCH=$* push

all-container: test $(addprefix sub-container-,$(ALL_ARCH))

all-push: $(addprefix sub-push-,$(ALL_ARCH))
container: .container-$(ARCH)
.container-$(ARCH):
	cp -r * $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i 's|BASEIMAGE|$(BASEIMAGE)|g' Dockerfile

	docker run --rm -v $(TEMP_DIR):$(TEMP_DIR):Z -v `pwd`:/go/src/k8s.io/autoscaler/addon-resizer/:Z \
        golang:${GOLANG_VERSION} \
        /bin/bash -c "\
            go get github.com/tools/godep && \
            cd /go/src/k8s.io/autoscaler/addon-resizer/ && \
            CGO_ENABLED=0 GOARM=$(GOARM) GOARCH=$(ARCH) godep go build -a -installsuffix cgo --ldflags '-w -X k8s.io/autoscaler/addon-resizer/nanny.AddonResizerVersion=$(TAG)' -o $(TEMP_DIR)/pod_nanny main.go"

	docker build -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)

ifeq ($(ARCH), amd64)
	# This is for to maintain the backward compatibility
	docker tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

test:
	docker run --rm -v `pwd`:/go/src/k8s.io/autoscaler/addon-resizer/:Z \
	golang:${GOLANG_VERSION} \
        /bin/bash -c "\
            go get github.com/tools/godep && \
            cd /go/src/k8s.io/autoscaler/addon-resizer/ && \
            godep go test ./nanny -v"

push: .push-$(ARCH)
.push-$(ARCH): .container-$(ARCH)
	gcloud docker -- push $(MULTI_ARCH_IMG):$(TAG)
ifeq ($(ARCH), amd64)
	gcloud docker -- push $(IMAGE):$(TAG)
endif

clean: $(addprefix sub-clean-,$(ALL_ARCH))
sub-clean-%:
	docker rmi -f $(IMAGE)-$*:$(TAG) || true
ifeq ($(ARCH), amd64)
	docker rmi -f $(IMAGE):$(TAG)
endif
