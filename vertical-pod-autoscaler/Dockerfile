FROM golang:1.21.0 as builder

ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$PATH

ARG GOARCH
ARG LDFLAGS

RUN go version
RUN go install github.com/tools/godep@latest
RUN godep version

WORKDIR /src
COPY . .
RUN go mod download
WORKDIR /src/pkg/recommender
RUN GOOS=linux CGO_ENABLED=0 go build -o /out

FROM gcr.io/distroless/static:latest
COPY --from=builder /out /recommender
ENTRYPOINT ["/recommender"]
