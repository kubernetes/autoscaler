# Copyright 2018 The Kubernetes Authors. All rights reserved
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
# limitations under the License.

FROM --platform=$BUILDPLATFORM golang:1.23.1 as builder

ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$PATH

COPY . /gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler
WORKDIR /gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=0 LD_FLAGS=-s GOARCH=$TARGETARCH GOOS=$TARGETOS go build -C pkg/admission-controller -mod vendor -o admission-controller-$TARGETARCH

FROM gcr.io/distroless/static:latest
MAINTAINER Tomasz Kulczynski "tkulczynski@google.com"

ARG TARGETARCH

COPY --from=builder /gopath/src/k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/admission-controller-$TARGETARCH /admission-controller

ENTRYPOINT ["/admission-controller"]
CMD ["--v=4", "--stderrthreshold=info"]
