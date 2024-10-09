ARG GO_VERSION=1.19.1
ARG GO_IMAGE=${GO_VERSION}-alpine

FROM golang:${GO_IMAGE} AS install
ARG IFACEMAKER_VERSION=v1.1.0
WORKDIR /local
RUN go install github.com/vburenin/ifacemaker@$IFACEMAKER_VERSION

FROM scratch AS binary
COPY --from=install /go/bin/ifacemaker /usr/local/bin/ifacemaker

ENTRYPOINT ["ifacemaker"]
