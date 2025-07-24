# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder

WORKDIR /app
COPY ../../.. .

# Build only the startup-script-server
RUN CGO_ENABLED=0 GOOS=linux go build -o startup-script-server cloudprovider/datacrunch/examples/startup-script-server.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/startup-script-server /startup-script-server

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/startup-script-server"] 