#!/usr/bin/env bash

# Copyright 2025 The Kubernetes Authors.
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

# canonicalize the path
CLUSTER_AUTOSCALER_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)

# directory where protoc releases are unpacked
protoc_dir="${CLUSTER_AUTOSCALER_ROOT}/hack/genproto/protobuf"
# ensure this directory contains bin/protoc and include/...
protoc_bin_dir="${protoc_dir}/bin"
protoc_include_dir="${protoc_dir}/include"
if [[ ! -f "${protoc_bin_dir}/protoc" || ! -d "${protoc_include_dir}" ]]; then
  protoc_version="33.0"
  case "$(uname -om)" in
    x86_64*Linux)
      protoc_url="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-x86_64.zip"
      ;;
    i?86_64*Linux)
      protoc_url="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-x86_64.zip"
      ;;
    amd64*Linux)
      protoc_url="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-x86_64.zip"
      ;;
    aarch64*Linux)
      protoc_url="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-aarch_64.zip"
      ;;
    arm64*Linux)
      protoc_url="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-aarch_64.zip"
      ;;
    *Darwin)
      protoc_url="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-osx-universal_binary.zip"
      ;;
    *)
      echo "Unsupported host os/arch. Must be Darwin, or x86_64/arm64 Linux." >&2
      exit 1
      ;;
  esac
  echo "Downloading ${protoc_url} to ${protoc_dir}"
  mkdir -p "${protoc_dir}"
  curl -s -L "${protoc_url}" -o "${protoc_dir}/protoc.zip"
  unzip -q "${protoc_dir}/protoc.zip" -d "${protoc_dir}"
fi

# Build protoc-gen-go and protoc-gen-go-grpc into bin directory
pushd "${CLUSTER_AUTOSCALER_ROOT}/hack/genproto" >/dev/null
  go mod tidy
  go mod vendor
  GOBIN="${protoc_bin_dir}" go install google.golang.org/protobuf/cmd/protoc-gen-go
  GOBIN="${protoc_bin_dir}" go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
popd >/dev/null

pushd "${CLUSTER_AUTOSCALER_ROOT}" >/dev/null

PATH="${protoc_bin_dir}" "${protoc_bin_dir}/protoc" \
  -I . \
  -I ./hack/genproto/vendor \
  -I "${protoc_include_dir}" \
  --go_out=. \
  --go-grpc_out=. \
  ./cloudprovider/externalgrpc/protos/externalgrpc.proto

PATH="${protoc_bin_dir}" "${protoc_bin_dir}/protoc" \
  -I . \
  -I ./hack/genproto/vendor \
  -I "${protoc_include_dir}" \
  --go_out=. \
  --go-grpc_out=. \
  ./expander/grpcplugin/protos/expander.proto

popd >/dev/null

rm -fr "${CLUSTER_AUTOSCALER_ROOT}/hack/genproto/vendor"
