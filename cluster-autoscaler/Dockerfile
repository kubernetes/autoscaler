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
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

WORKDIR /workspace

# Copy go.mod and go.sum files first to cache dependencies
COPY go.mod go.sum ./
COPY apis/go.mod apis/go.sum ./apis/

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

ARG GOARCH
ARG LDFLAGS
ARG BUILD_TAGS

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    if [ -n "$BUILD_TAGS" ]; then \
        CGO_ENABLED=0 GOOS=linux go build -o cluster-autoscaler-$GOARCH -ldflags="$LDFLAGS" -tags="$BUILD_TAGS"; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -o cluster-autoscaler-$GOARCH -ldflags="$LDFLAGS"; \
    fi

FROM gcr.io/distroless/static:nonroot
ARG GOARCH
COPY --from=builder /workspace/cluster-autoscaler-$GOARCH /cluster-autoscaler

WORKDIR /
CMD ["/cluster-autoscaler"]
