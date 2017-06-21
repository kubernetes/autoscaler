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

FROM gcr.io/google-containers/ubuntu-slim:0.8
MAINTAINER Marcin Wielgus "mwielgus@google.com"

RUN apt-get update && apt-get install --yes git wget make gcc \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
RUN wget https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz \
   && tar -xvf go1.8.3.linux-amd64.tar.gz \
   && rm go1.8.3.linux-amd64.tar.gz

ENV GOROOT /go
ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
RUN go version
RUN go get github.com/tools/godep
RUN godep version
CMD ["/bin/bash"]
