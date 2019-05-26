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

FROM golang:1.11.2
ENV GOPATH /gopath
ENV PATH $GOPATH/bin:$PATH
ADD ./cluster-autoscaler/ $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/
RUN go get github.com/tools/godep
RUN rm -Rf $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/cluster-autoscaler && cd $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler && GOOS=linux make build-binary
RUN chmod 755 $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/cluster-autoscaler
RUN cp $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/cluster-autoscaler /cluster-autoscaler
RUN cp $GOPATH/src/k8s.io/autoscaler/cluster-autoscaler/run.sh /run.sh

FROM k8s.gcr.io/debian-base-amd64:v1.0.0
ENV DEBIAN_FRONTEND noninteractive
ADD cluster-autoscaler cluster-autoscaler
RUN clean-install ca-certificates tzdata
ADD cluster-autoscaler/run.sh /run.sh

CMD ./run.sh
