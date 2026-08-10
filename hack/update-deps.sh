#!/usr/bin/env bash

# Copyright 2021 The Kubernetes Authors.
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

###
# This script is to be used when updating Kubernetes and its staging
# repositories to *tagged* releases. This is the ideal case, but another
# script, submodule-k8s.sh, is available as a break-glass solution if we must
# switch to an unreleased commit.
###

set -o errexit
set -o pipefail

KUBE_ROOT="$(dirname "${BASH_SOURCE[0]}")/.."
cd "${KUBE_ROOT}"

VERSION=${1#"v"}
FORK=${2:-git@github.com:kubernetes/kubernetes.git}
SED=${3:-sed}

# $1: The k8s version to download.
cluster_autoscaler:list_mods:init() {
  k8s_version="${1:-${VERSION}}"

  workdir="$(mktemp -d)"
  repo="${workdir}/kubernetes"
  git clone --depth 1 "${FORK}" "${repo}"
  pushd "${repo}"
  git fetch --depth 1 origin "v${k8s_version}"
  git checkout FETCH_HEAD
}

cluster_autoscaler:list_mods:cleanup() {
  popd
  rm -rf "${workdir}"
}

# $1: The k8s version to download.
cluster_autoscaler:list_mods() {
  k8s_version="${1:-${VERSION}}"

  if [ -z "${k8s_version}" ]; then
    echo "Usage: hack/update-deps.sh <k8s version> <k8s fork:-git@github.com:kubernetes/kubernetes.git>"
    exit 1
  fi
  cluster_autoscaler:list_mods:init "${k8s_version}" > /dev/null
  mods=($(
        cat go.mod | sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
  ))
  cluster_autoscaler:list_mods:cleanup > /dev/null
  echo "${mods[@]}"
}

# $1: The package path.
# $2: The k8s version.
# $3: and later: The module names.
cluster_autoscaler:update_deps() {
  args=("${@}")
  pkg="${args[0]}"
  k8s_version="${args[1]}"
  mods=("${args[@]:2}")

  pushd "${pkg}"

  for mod in "${mods[@]}"; do
    echo "${pkg}: ${mod}"
    gomod="${mod}@kubernetes-${k8s_version}"
    if ! gomod_json="$(go mod download -json "${gomod}")"; then
      echo "Error downloading module ${gomod}: the tag might not be published in the staging repo."
      exit 1
    fi
    mod_version=$(echo "${gomod_json}" | "${SED}" -n 's|.*"Version": "\(.*\)".*|\1|p')
    if [ "${pkg}" = "." ] || [ "${pkg}" = "./pkg/e2e" ]; then
      go mod edit "-replace=${mod}=${mod}@${mod_version}"
    else
      go get "${mod}@${mod_version}"
    fi
  done

  if [ "${pkg}" = "." ] || [ "${pkg}" = "./pkg/e2e" ]; then
    go get "k8s.io/kubernetes@v${k8s_version}"
  fi

  go mod tidy

  git rm -r --force --ignore-unmatch kubernetes
  popd
}

# sigs.k8s.io/cluster-autoscaler/go.mod
mods=($(cluster_autoscaler:list_mods "${VERSION}"))
cluster_autoscaler:update_deps "." "${VERSION}" "${mods[@]}"

# sigs.k8s.io/cluster-autoscaler/pkg/e2e/go.mod
cluster_autoscaler:update_deps "./pkg/e2e" "${VERSION}" "${mods[@]}"
