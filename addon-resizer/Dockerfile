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

FROM --platform=$BUILDPLATFORM golang:1.24 AS builder

WORKDIR /workspace

COPY . .

ARG ARCH GOARM
ARG TAG

RUN CGO_ENABLED=0 GOARM=$GOARM GOARCH=$ARCH go build -a -installsuffix cgo --ldflags "-w -X k8s.io/autoscaler/addon-resizer/nanny.AddonResizerVersion=$TAG" -o pod_nanny nanny/main/pod_nanny.go

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /workspace/pod_nanny /pod_nanny

USER nobody

ENTRYPOINT ["/pod_nanny"]
