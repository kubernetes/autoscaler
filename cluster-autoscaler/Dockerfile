#############      builder                                  #############
FROM golang:1.23.2 AS builder

WORKDIR /go/src/github.com/gardener/autoscaler
COPY . .

RUN .ci/build

#############      base                                     #############
FROM gcr.io/distroless/static-debian11:nonroot as base
WORKDIR /

#############      cluster-autoscaler               #############
FROM base AS cluster-autoscaler

COPY --from=builder /go/src/github.com/gardener/autoscaler/cluster-autoscaler/cluster-autoscaler /cluster-autoscaler

# Copy the VERSION file which helps in printing g/autoscaler version
COPY --from=builder /go/src/github.com/gardener/autoscaler/VERSION /VERSION

CMD ["/cluster-autoscaler"]
