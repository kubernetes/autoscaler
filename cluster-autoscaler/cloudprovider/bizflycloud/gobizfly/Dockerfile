

FROM golang:1.13 AS build_base

RUN apt-get update && apt-get install -y git pkg-config

# stage 2
from build_base AS build_go

ENV GO111MODULE=on

WORKDIR $GOPATH/src/github.com/bizflycloud/gobizfly
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod vendor
# # RUN CGO_ENABLED=0 GOOS=linux go get

# stage 3
FROM build_go AS server_builder

ENV GO111MODULE=on

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -gcflags="-N -l" -o /bin/gobizfly *.go

ENTRYPOINT [ "/bin/gobizfly" ]


