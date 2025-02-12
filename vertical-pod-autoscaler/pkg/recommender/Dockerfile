# Copyright 2017 The Kubernetes Authors. All rights reserved
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

FROM --platform=$BUILDPLATFORM golang:1.23.6 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY common common
COPY pkg pkg

ARG TARGETOS TARGETARCH
ARG BUILD_LD_FLAGS="-s"

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=$TARGETOS go build -C pkg/recommender -ldflags="${BUILD_LD_FLAGS}" -o recommender-$TARGETARCH

FROM gcr.io/distroless/static:nonroot

ARG TARGETARCH

COPY --from=builder /workspace/pkg/recommender/recommender-$TARGETARCH /recommender

ENTRYPOINT ["/recommender"]
