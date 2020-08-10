#env /usr/bin/bash

set -o errexit
set -o pipefail
set -o nounset
shopt -s lastpipe

if [[ $(basename $(pwd)) != "cluster-autoscaler" ]];then
  echo "The script must be run in cluster-autoscaler directory"
  exit 1
fi

if ! which jq > /dev/null; then
  echo "This script requires jq command to be available"
  exit 1
fi

SCRIPT_NAME=$(basename "$0")
K8S_FORK=${K8S_FORK:-"git@github.com:kubernetes/kubernetes.git"}
K8S_REV="master"
BATCH_MODE="false"
TARGET_MODULE=${TARGET_MODULE:-k8s.io/autoscaler/cluster-autoscaler}
VERIFY_COMMAND=${VERIFY_COMMAND:-"go test -mod=vendor ./..."}
OVERRIDE_GO_VERSION="false"

ARGS="$@"
OPTS=`getopt -o f::r::d::v::b::o:: --long k8sfork::,k8srev::,workdir::,batch::,override-go-version:: -n $SCRIPT_NAME -- "$@"`
if [ $? != 0 ] ; then echo "Failed parsing options." >&2 ; exit 1 ; fi
eval set -- "$OPTS"
while true; do
  case "$1" in
    -f | --k8sfork ) K8S_FORK="$2"; shift; shift ;;
    -r | --k8srev ) K8S_REV="$2"; shift; shift ;;
    -d | --workdir ) WORK_DIR="$2"; shift; shift ;;
    -b | --batch ) BATCH_MODE="true"; shift; shift ;;
    -o | --override-go-version) OVERRIDE_GO_VERSION="true"; shift; shift ;;
    -v ) VERBOSE=1; shift; if [[ "$1" == "v" ]]; then VERBOSE=2; shift; fi; ;;
    -- ) shift; break ;;
    * ) break ;;
  esac
done

export GO111MODULE=on

set -o errexit
WORK_DIR="${WORK_DIR:-$(mktemp -d /tmp/ca-update-vendor.XXXX)}"
echo "Operating in ${WORK_DIR}"

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

EXPECTED_ERROR_MARKER="${WORK_DIR}/expected_error"

# Try
set +o errexit
(
  set -o errexit
  rm -f $EXPECTED_ERROR_MARKER
  K8S_REPO="${WORK_DIR}/kubernetes"
  if [ -d ${K8S_REPO} ]; then
    pushd ${K8S_REPO} >/dev/null
    if [[ "$(git remote get-url origin)" != "${K8S_FORK}" ]]; then
      echo "Mismated checked out k8s repo; deleting"
      rm -rf "${K8S_REPO}"
    fi
    popd >/dev/null
  fi

  echo "Updating vendor against ${K8S_FORK}:${K8S_REV}"

  if [ ! -d ${K8S_REPO} ]; then
    echo "Cloning ${K8S_FORK} into ${K8S_REPO}"
    git clone --depth 1 ${K8S_FORK} ${K8S_REPO} >&${BASH_XTRACEFD} 2>&1
  fi

  pushd ${K8S_REPO} >/dev/null
  git fetch --depth 1 origin ${K8S_REV} >&${BASH_XTRACEFD} 2>&1
  git checkout FETCH_HEAD >&${BASH_XTRACEFD} 2>&1
  K8S_REV_PARSED=$(git rev-parse FETCH_HEAD)
  popd >/dev/null


  function err_rerun() {
    touch ${EXPECTED_ERROR_MARKER}
    echo "$*"
    if [[ "${BATCH_MODE}" == "false" ]]; then
      echo "Fix errors and rerun script:"
      echo " $0 -d${WORK_DIR} -f${K8S_FORK} -r${K8S_REV}"
    fi
    exit 1
  }

  # Deleting old stuff
  rm -rf vendor
  rm -f go.mod
  rm -f go.sum

  # Base CA go.mod on one from k8s.io/kuberntes
  cp $K8S_REPO/go.mod .

  # Check go version
  REQUIRED_GO_VERSION=$(cat go.mod  |grep '^go ' |tr -s ' ' |cut -d ' '  -f 2)
  USED_GO_VERSION=$(go version |sed 's/.*go\([0-9]\+\.[0-9]\+\).*/\1/')


  if [[ "${REQUIRED_GO_VERSION}" != "${USED_GO_VERSION}" ]];then
    if [[ "${OVERRIDE_GO_VERSION}" == "false" ]]; then
      err_rerun "Invalid go version ${USED_GO_VERSION}; required go version is ${REQUIRED_GO_VERSION}."
    else
      echo "Overriding go version found in go.mod file. Expected go version ${REQUIRED_GO_VERSION}, using ${USED_GO_VERSION}"
    fi
  fi

  # Fix module name and staging modules links
  sed -i "s#module k8s.io/kubernetes#module ${TARGET_MODULE}#" go.mod
  sed -i "s#\\./staging#${K8S_REPO}/staging#" go.mod

  function list_dependencies() {
    local_tmp_dir=$(mktemp -d "${WORK_DIR}/list_dependencies.XXXX")
    local go_dep_file="$1"
    local tmp_file="${local_tmp_dir}/list_dependencies.tmp"
    rm -f ${tmp_file}
    go mod edit -json ${go_dep_file} |jq -r '.Replace[]? | select(.New.Version != null)| "\(.Old.Path) \(.New.Version)"' >> ${tmp_file}
    go mod edit -json ${go_dep_file} |jq -r '.Require[]? | "\(.Path) \(.Version)"' >> ${tmp_file}
    cat ${tmp_file} |sort |uniq
  }

  function version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1";
  }

  GO_MOD_EXTRA_FILES="$(shopt -s nullglob;echo go.mod-extra*)"
  OLD_EXTRA_FOUND="false"
  for go_mod_extra in ${GO_MOD_EXTRA_FILES}; do
    list_dependencies ${go_mod_extra} | while read extra_path extra_version; do
      list_dependencies go.mod | while read source_path source_version; do
        if [[ "${source_path}" == "${extra_path}" ]]; then
          if ! version_gt $extra_version $source_version; then
            echo "Extra dependency ${source_path} already used by k8s in >= version ${source_version}"
            OLD_EXTRA_FOUND="true"
          fi
        fi
      done
    done
  done
  if [[ "${OLD_EXTRA_FOUND}" == "true" ]]; then
    err_rerun "Extra dependencies found in one of go.mod-extra files"
  fi

  # Add dependencies from go.mod-extra to go.mod
  # Propagate require entries to both require and replace
  for go_mod_extra in ${GO_MOD_EXTRA_FILES}; do
    go mod edit -json ${go_mod_extra} | jq -r '.Require[]? | "-require \(.Path)@\(.Version)"' | xargs -t -r go mod edit >&${BASH_XTRACEFD} 2>&1
    go mod edit -json ${go_mod_extra} | jq -r '.Require[]? | "-replace \(.Path)=\(.Path)@\(.Version)"' | xargs -t -r go mod edit >&${BASH_XTRACEFD} 2>&1
    # And add explicit replace entries
    go mod edit -json ${go_mod_extra} | jq -r '.Replace[]? | "-replace \(.Old.Path)=\(.New.Path)@\(.New.Version)"' | sed "s/@null//g" |xargs -t -r go mod edit >&${BASH_XTRACEFD} 2>&1
  done
  # Add k8s.io/kubernetes dependency
  go mod edit -require k8s.io/kubernetes@v0.0.0
  go mod edit -replace k8s.io/kubernetes=${K8S_REPO}

  # Fail if there are implicit dependencies
  list_dependencies go.mod > ${WORK_DIR}/packages-before-tidy
  go mod tidy -v >&${BASH_XTRACEFD} 2>&1
  list_dependencies go.mod > ${WORK_DIR}/packages-after-tidy

  IMPLICIT_FOUND="false"
  set +o pipefail
  diff -u ${WORK_DIR}/packages-before-tidy ${WORK_DIR}/packages-after-tidy | grep -v '\+\+\+ ' | grep '^\+' | cut -b 2- |while read line; do
    IMPLICIT_FOUND="true"
    echo "Implicit dependency found: ${line}"
  done
  set -o pipefail

  if [[ "${IMPLICIT_FOUND}" == "true" ]]; then
    err_rerun "Implicit dependencies missing from go.mod-extra"
  fi

  echo "Running go mod vendor"
  go mod vendor

  echo "Running ${VERIFY_COMMAND}"
  if ! ${VERIFY_COMMAND} >&${BASH_XTRACEFD} 2>&1; then
    err_rerun "Verify command failed"
  fi

  # Commit go.mod* and vendor
  git reset . >&${BASH_XTRACEFD} 2>&1
  git add vendor go.mod go.sum >&${BASH_XTRACEFD} 2>&1
  if ! git diff --quiet --cached; then
    echo "Commiting vendor, go.mod and go.sum"
    git commit -m "Updating vendor against ${K8S_FORK}:${K8S_REV} (${K8S_REV_PARSED})" >&${BASH_XTRACEFD} 2>&1
  else
    echo "No changes after vendor update; skipping commit"
  fi


  if ! git diff --quiet; then
    echo "Uncommited changes (manual fixes?) still present in repository - please commit those"
  fi

  echo "Operation finished successfully"
  if [[ "$(basename "${WORK_DIR}" | cut -d '.' -f 1)" == "ca-update-vendor" ]];then
    echo "Deleting working directory ${WORK_DIR}"
    rm -rf ${WORK_DIR}
  else
    echo "Preserving working directory ${WORK_DIR}"
  fi
)

# Catch
err=$?
if [[ $err -ne 0 ]]; then
  if [ ! -f "${EXPECTED_ERROR_MARKER}" ]; then
    echo
    echo "Unexpected error occured; check $LOG_FILE"
  fi
fi
exit $err
