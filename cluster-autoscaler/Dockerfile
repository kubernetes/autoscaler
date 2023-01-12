#############      builder                                  #############
FROM golang:1.19.5 AS builder

WORKDIR /go/src/github.com/gardener/autoscaler
COPY . .

RUN .ci/build

#############      base                                     #############
FROM gcr.io/distroless/static-debian11:nonroot as base
WORKDIR /

#############      cluster-autoscaler               #############
FROM base AS cluster-autoscaler

COPY --from=builder /go/src/github.com/gardener/autoscaler/cluster-autoscaler/cluster-autoscaler /cluster-autoscaler

CMD ["/cluster-autoscaler"]
