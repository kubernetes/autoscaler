ARG BASEIMAGE=gcr.io/distroless/static:nonroot

FROM golang:1.21 as build

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -v -o /build/cluster-autoscaler

FROM $BASEIMAGE

COPY --from=build /build/cluster-autoscaler /cluster-autoscaler
WORKDIR /
CMD ["/cluster-autoscaler"]