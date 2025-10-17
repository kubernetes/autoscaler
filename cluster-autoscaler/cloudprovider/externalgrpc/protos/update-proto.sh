#!/usr/bin/env bash

# canonicalize the path
CLUSTER_AUTOSCALER_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)

pushd "${CLUSTER_AUTOSCALER_ROOT}/cloudprovider/externalgrpc/protos/third_party" >/dev/null
  go mod tidy
  go mod vendor
popd >/dev/null

protoc_binary="cloudprovider/externalgrpc/protos/third_party/protobuf/bin/protoc"
if [ ! -f "${CLUSTER_AUTOSCALER_ROOT}/${protoc_binary}" ]; then
  echo "No ${protoc_binary} binary found"
  echo "Unpack latest release from https://github.com/protocolbuffers/protobuf/tags"
  exit 1
fi

protoc_include="cloudprovider/externalgrpc/protos/third_party/protobuf/include"
if [ ! -d "${CLUSTER_AUTOSCALER_ROOT}/${protoc_include}" ]; then
  echo "No ${protoc_include} directory found"
  echo "Unpack latest release from https://github.com/protocolbuffers/protobuf/tags"
  exit 1
fi

pushd "${CLUSTER_AUTOSCALER_ROOT}" >/dev/null
"${CLUSTER_AUTOSCALER_ROOT}/${protoc_binary}" \
  -I . \
  -I ./cloudprovider/externalgrpc/protos/third_party/vendor \
  -I "${CLUSTER_AUTOSCALER_ROOT}/${protoc_include}" \
  --go_out=. \
  --go-grpc_out=. \
  ./cloudprovider/externalgrpc/protos/externalgrpc.proto
popd >/dev/null

rm -fr "${CLUSTER_AUTOSCALER_ROOT}/cloudprovider/externalgrpc/protos/third_party/vendor"
