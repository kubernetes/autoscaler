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

# Procedure for change:
# 1. Bump the tag number.
# 2. Create an image for the kubernetes user on Dockerhub and push there: make dev-release
# 3. Temporarily adjust any config files etc. that reference gcr.io/google_containers/addon-resizer:$(TAG)
#    to use kubernetes/addon-resizer:$(TAG) instead.
# 4. Test etc.
# 5. Issue PR.
# 6. When PR is approved make the gcr.io version of the image (Change REGISTRY variable below first): make release
# 7. Revert the references to kubernetes/addon-resizer:$(TAG) to gcr.io/google_containers/addon-resizer:$(TAG)

all: build

PACKAGE = k8s.io/autoscaler/addon-resizer
REGISTRY = kubernetes
# Uncomment for release
# REGISTRY = gcr.io/google_containers
TAG = 2.0

deps:
		go get -u github.com/tools/godep

compile: nanny/ deps
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -a -o pod_nanny nanny/main/pod_nanny.go

test: nanny/
		godep go test ${PACKAGE}/nanny -v

go: compile test

clean:
	rm -rf pod_nanny

build: go

dev-release: build execute-release
	echo "Release ${TAG} completed"

# Rules for building the real image for deployment to gcr.io

execute-release:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/addon-resizer:${TAG} .
	gcloud docker -- push ${REGISTRY}/addon-resizer:${TAG}

docker-builder:
	docker build -t autoscaling-builder ../builder

build-in-docker: clean docker-builder
	docker run -v `pwd`:/gopath/src/k8s.io/autoscaler/addon-resizer/ autoscaling-builder:latest bash -c 'cd /gopath/src/k8s.io/autoscaler/addon-resizer && make'

release: build-in-docker execute-release
	echo "Full in-docker release ${TAG} completed"


