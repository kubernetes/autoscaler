#!/bin/bash
set -euxo pipefail

REPO="${GOPATH}/src/k8s.io/autoscaler"
IMAGE="autoscaler-dependency-builder:latest"

docker build --compress -t "${IMAGE}" -f dependencies/Dockerfile .
docker run --rm -v "${REPO}/cluster-autoscaler:/tmp/out" "${IMAGE}" sh -c "cp cluster-autoscaler /tmp/out/"
( cd "${REPO}/cluster-autoscaler" && make make-image )
