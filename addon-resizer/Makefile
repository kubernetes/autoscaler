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
export DOCKER_CLI_EXPERIMENTAL=enabled

GOARM=7
GOLANG_VERSION = 1.24
REGISTRY = gcr.io/k8s-staging-autoscaling
IMGNAME = addon-resizer
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)
TAG = 1.8.23
# The output type could either be docker (local), or registry.
OUTPUT_TYPE ?= docker

TEMP_DIR := $(shell mktemp -d)

all: all-container

sub-container-%:
	$(MAKE) ARCH=$* container

sub-push-%:
	$(MAKE) ARCH=$* push

all-container: test $(addprefix sub-container-,$(ALL_ARCH))

all-push: $(addprefix sub-push-,$(ALL_ARCH)) push-multi-arch

buildx-setup:
	docker buildx inspect img-builder > /dev/null || docker buildx create --name img-builder --use

container: .container-$(ARCH)
.container-$(ARCH): buildx-setup
	cp -r * $(TEMP_DIR)
	cd $(TEMP_DIR)

	docker buildx build \
		--pull \
		--platform linux/$(ARCH) \
		--build-arg ARCH=$(ARCH) \
		--build-arg GOARM=$(GOARM) \
		--build-arg TAG=$(TAG) \
		--output=type=$(OUTPUT_TYPE) \
		-t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)

test:
	docker run --rm -v `pwd`:/go/src/k8s.io/autoscaler/addon-resizer/:Z \
	golang:${GOLANG_VERSION} \
        /bin/bash -c "\
            cd /go/src/k8s.io/autoscaler/addon-resizer/ && \
            go test ./nanny -v"

push: .push-$(ARCH)
.push-$(ARCH): .container-$(ARCH)
	gcloud docker -- push $(MULTI_ARCH_IMG):$(TAG)

push-multi-arch:
	docker manifest create --amend $(IMAGE):$(TAG) $(shell echo $(ALL_ARCH) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(TAG)~g")
	@for arch in $(ALL_ARCH); do docker manifest annotate --arch $${arch} $(IMAGE):$(TAG) $(IMAGE)-$${arch}:$(TAG); done
	gcloud docker -- manifest push --purge $(IMAGE):$(TAG)

clean: $(addprefix sub-clean-,$(ALL_ARCH))
sub-clean-%:
	docker rmi -f $(IMAGE)-$*:$(TAG) || true

