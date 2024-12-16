# Copyright 2016 The Kubernetes Authors. All rights reserved
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

FROM --platform=$BUILDPLATFORM golang:1.23.4 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY multidimensional-pod-autoscaler/go.mod go.mod
COPY multidimensional-pod-autoscaler/go.sum go.sum
# TODO: This is temporary until the VPA has cut a new release
COPY vertical-pod-autoscaler /vertical-pod-autoscaler

RUN go mod download

COPY multidimensional-pod-autoscaler/common common
COPY multidimensional-pod-autoscaler/pkg pkg

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=0 LD_FLAGS=-s GOARCH=$TARGETARCH GOOS=$TARGETOS go build -C pkg/updater -o updater-$TARGETARCH

FROM gcr.io/distroless/static:nonroot

ARG TARGETARCH

COPY --from=builder /workspace/pkg/updater/updater-$TARGETARCH /updater

ENTRYPOINT ["/updater"]
