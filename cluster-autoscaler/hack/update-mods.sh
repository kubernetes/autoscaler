#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

set -e
set -o pipefail
set -o nounset
shopt -s lastpipe

err_report() {
    echo "Exited with error on line $1"
}
trap 'err_report $LINENO' ERR

SCRIPT_NAME=$(basename "$0")
K8S_FORK="kubernetes/kubernetes"
TARGET_MODULE=${TARGET_MODULE:-k8s.io/autoscaler/cluster-autoscaler}
PRINT_HELP="false"
UPDATE_GO_MOD="true"

# Packages that are allowed to drift from k8s/k8s
PERMITTED_PACKAGE_DIFFS=(
    github.com/aws/aws-sdk-go
    github.com/stretchr/testify
)


################################################################################
### Flag and directory validation ##############################################
################################################################################

if [[ $(basename $(pwd)) != "cluster-autoscaler" ]];then
  echo "The script must be run in cluster-autoscaler directory"
  exit 1
fi
AUTOSCALER_DIR=$(pwd)

if ! which jq > /dev/null; then
  echo "This script requires jq command to be available"
  exit 1
fi

K8S_REV="$(cat ./hack/k8s_rev)"

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h | --help ) PRINT_HELP="true"; break;;
        -f | --k8sfork ) K8S_FORK="$2"; shift; shift ;;
        -r | --k8srev ) K8S_REV="$2"; shift; shift ;;
        -d | --workdir ) WORK_DIR="$2"; shift; shift ;;
        -n | --no-update ) UPDATE_GO_MOD="false"; shift ;;
        * ) break ;;
    esac
done

if [[ "${PRINT_HELP}" == "true" ]]; then
    cat << EOF >> /dev/stderr
$SCRIPT_NAME - Updates or detects differences in go.mod, go.sum, /vendor

    -f --k8sfork    The Kubernetes fork to use. Defaults to $K8S_FORK
    -r --k8srev     Kubernetes revision to use. Defaults to ${K8S_REV}
    -d --workdir    Working dir. Defaults to a temporary dir
    -n --no-update  Set flag to only view differences in go.mod, /vendor
    -h --help       Print this help output
EOF
    exit 0
fi

export GO111MODULE=on

WORK_DIR="${WORK_DIR:-$(mktemp -d /tmp/ca-update-vendor.XXXX)}"
echo "Operating in $WORK_DIR"
if [ ! -d $WORK_DIR ]; then
  echo "Work dir ${WORK_DIR} does not exist"
  exit 1
fi

LOG_FILE="${LOG_FILE:-${WORK_DIR}/ca-update-vendor.log}"
echo "Sending logs to: ${LOG_FILE}"
if [ -z "${BASH_XTRACEFD:-}" ]; then
  exec 19> "${LOG_FILE}"
  export BASH_XTRACEFD="19"
fi
set -x

USED_GO_VERSION=$(go version | grep -oE '[0-9].[0-9]{2}')
AUTOSCALER_GO_VERSION=$(go mod edit -json | jq -r '.Go')

################################################################################
### Get k8s.io/kubernetes/go.mod@$K8S_REV ######################################
################################################################################
cd $WORK_DIR
K8S_GO_MOD_JSON="k8s-go-mod.json"
curl -o $WORK_DIR/go.mod -s "https://raw.githubusercontent.com/${K8S_FORK}/${K8S_REV}/go.mod"
go mod edit -json > $WORK_DIR/$K8S_GO_MOD_JSON 2>&1

# Read all `./staging` replacements from Kubernetes go.mod
K8S_STAGING_MODULES=()
while IFS= read -r line; do
    K8S_STAGING_MODULES+=( "$line" )
done < <(go mod edit -json | jq -r '.Replace[] | select(.New.Path | contains("./staging")) | .Old.Path' 2>&1)

K8S_GO_VERSION=$(go mod edit -json | jq -r '.Go')

################################################################################
### Ensure Go version ##########################################################
################################################################################

if [[ "${USED_GO_VERSION}" != "${AUTOSCALER_GO_VERSION}" ]];then
    echo "Invalid go version: using ${USED_GO_VERSION}; required go version is ${AUTOSCALER_GO_VERSION}."
    exit 1
fi
if [[ "${K8S_GO_VERSION}" != "${AUTOSCALER_GO_VERSION}" ]];then
    echo "Invalid go version: Autoscaler is using ${AUTOSCALER_GO_VERSION}; kubernetes is using ${K8S_GO_VERSION}."
    exit 1
fi

################################################################################
### Ensure k8s.io/kubernetes version and overrides in mod file #################
################################################################################
cd $AUTOSCALER_DIR
K8S_PSEUDO_VERSION=""
if [ "$UPDATE_GO_MOD" == "true" ]; then
    # Drop commit hash as version
    go mod edit -require=k8s.io/kubernetes@$K8S_REV 2>&1
    # Get go toolchain to replace commit hash with pseudo version
    go mod tidy 2>&1
    K8S_PSEUDO_VERSION=$(go mod edit -json | jq -r '.Require[] | select(.Path == "k8s.io/kubernetes") | .Version')

    ## Add staging module overrides.
    ## Extraneous replacements are harmless
    for mod in ${K8S_STAGING_MODULES[*]}; do
        go mod edit -require=${mod}@v0.0.0 2>&1
        go mod edit -replace=${mod}=k8s.io/kubernetes/staging/src/${mod}@${K8S_PSEUDO_VERSION} 2>&1
    done
fi

################################################################################
### Ensure k8s.io/kubernetes version and overrides in mod file #################
################################################################################

# Get Cluster Autoscaler required packages
CA_GO_MOD_JSON="ca-go-mod.json"
go mod edit -json > $WORK_DIR/$CA_GO_MOD_JSON 2>&1
CA_REQS=$(cat $WORK_DIR/$CA_GO_MOD_JSON | jq -r .Require[].Path)

# Record Cluster Autoscaler differences from k8s/k8s
CA_DIFFS=()
CA_ONLY_MODS=()
echo "Cluster Autoscaler & Kubernetes Dependencies"
for req in $CA_REQS; do
    # Get corresponding require version from k8s/k8s
    K8S_REQUIRE_VERSION=$(cat $WORK_DIR/$K8S_GO_MOD_JSON | jq --arg req $req -r '.Require[] | select(.Path == $req) | .Version')
    # Get corresponding replace version from k8s/k8s
    K8S_REPLACE_VERSION=$(go mod edit -json | jq --arg req $req -r '.Replace[] | select(.New.Path == $req)| .New.Version')
    # If k8s sets a require, choose that over a replace
    K8S_VERSION=${K8S_REPLACE_VERSION:-$K8S_REQUIRE_VERSION}

    # Skip modules staged in k8s/k8s
    SKIP="0"
    IS_K8S_STAGED="0"
    for STAGED_MOD in ${K8S_STAGING_MODULES[*]}; do
        if [ "$req" == "$STAGED_MOD" ]; then
            SKIP="1"
            IS_K8S_STAGED="1"
        fi
    done
    # Skip packages that are permitted to drift
    for ALLOWED_DIFF in ${PERMITTED_PACKAGE_DIFFS[*]}; do
        if [ "$req" == "$ALLOWED_DIFF" ]; then
            SKIP="1"
        fi
    done

    CA_VERSION=$(cat $WORK_DIR/$CA_GO_MOD_JSON | jq -r --arg req $req '.Require[] | select(.Path == $req) | .Version')
    # if cluster-autoscaler requires it and k8s/k8s does, use k8s/k8s version
    if [[ -n "$K8S_VERSION" && "$SKIP" -eq "0" ]]; then

        if [ "$CA_VERSION" != "$K8S_VERSION" ]; then
            CA_DIFFS+=("${req}@${CA_VERSION} != ${req}@${K8S_VERSION}")
        else
            echo "    ${req}@${K8S_VERSION}"
        fi

        if [ "$UPDATE_GO_MOD" == "true" ]; then
            echo "    * Pinning kubernetes dependency in autoscaler to ${req}@${K8S_VERSION}"
            # pin dependency to k8s/k8s & force replace
            go mod edit -require=${req}@${K8S_VERSION} 2>&1
            go mod edit -replace=${req}=${req}@${K8S_VERSION} 2>&1
        fi
    fi
    if [[ -z "$K8S_VERSION" && "$IS_K8S_STAGED" -eq "0" ]]; then
        CA_ONLY_MODS+=("${req}@${CA_VERSION}")
    fi
done

if [ "$UPDATE_GO_MOD" != "true" ]; then
    echo "Cluster Autoscaler dependencies that differ from k8s/k8s:"
    for mod in "${CA_DIFFS[@]}"; do
        echo "    $mod";
    done
fi

echo "Cluster Autoscaler dependencies not in k8s/k8s:"
for mod in "${CA_ONLY_MODS[@]}"; do
    echo "    $mod";
done

if [ "$UPDATE_GO_MOD" == "true" ]; then
    echo "go mod vendor"
    go mod vendor 2>&1
    echo "go mod tidy"
    go mod tidy -v 2>&1
    go test -mod=vendor -v

    # Save revision
    echo $K8S_REV > ./hack/k8s_rev
    git add hack/k8s_rev vendor go.mod go.sum >&${BASH_XTRACEFD} 2>&1
    if ! git diff --quiet --cached; then
        echo "Committing vendor, go.mod and go.sum"
        git ci -m "Updating vendor against ${K8S_FORK}:${K8S_REV} (${K8S_PSEUDO_VERSION})" >&${BASH_XTRACEFD} 2>&1
    else
        echo "No changes after vendor update; skipping commit"
    fi

fi
