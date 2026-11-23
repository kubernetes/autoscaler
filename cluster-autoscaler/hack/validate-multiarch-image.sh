#!/usr/bin/env bash

# Copyright The Kubernetes Authors.
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
# Validate that a multi-arch cluster-autoscaler image is built correctly, i.e.
# that the binary shipped for each published platform is *actually* compiled for
# that architecture.
#
# This guards against the class of bug reported in
# https://github.com/kubernetes/autoscaler/issues/9226 where the linux/amd64
# image for v1.32.6 actually contained an aarch64 (arm64) binary. Because the
# manifest list can advertise the correct platform while the binary inside is
# wrong, we cannot trust the manifest alone: we pull each platform's image,
# extract /cluster-autoscaler and inspect its real ELF machine type.
#
# It generalizes the manual reproduction steps from the issue:
#
#   docker pull   --platform linux/amd64 <image>
#   docker create --platform linux/amd64 --name ca <image>
#   docker cp ca:/cluster-autoscaler /tmp/ca
#   file /tmp/ca
#   docker rm ca
#
# across every architecture in the manifest list, and turns the result into a
# pass/fail check suitable for release validation / CI.

set -o errexit
set -o nounset
set -o pipefail

readonly DEFAULT_BINARY="/cluster-autoscaler"
BINARY="$DEFAULT_BINARY"
RUNTIME="${CONTAINER_RUNTIME:-}"
ARCH_SPEC=""
MODE=release
KEEP=0
REF=""

usage() {
  cat <<EOF
Usage: $(basename "$0") [options] <image-ref>

Check the ELF architecture of /cluster-autoscaler without running the image.
The default (release) check requires an inspectable index, checks every published
Linux image by digest, and fails if any architecture in this branch's ALL_ARCH
is missing. Each variant is checked for its CPU architecture, not its ISA level.
For promotion, pass the index digest, not a mutable tag.

  --require-arches LIST  Release check using this required matrix instead of
                         asking the adjacent Makefile for ALL_ARCH (used by CI).
  -a, --arches LIST      Manual spot check of just these architectures. If the
                         index is unavailable, try explicit platform pulls instead.
                         Lists may be space- or comma-separated.
  -b, --binary PATH      Image path to inspect (default: ${DEFAULT_BINARY}).
  -r, --runtime RUNTIME  docker or podman (default: \$CONTAINER_RUNTIME, then
                         docker, then podman).
  -k, --keep             Keep extracted binaries in the reported temp directory.
  -h, --help             Show this help.

Index inspection uses Python 3 (standard library); ELF inspection uses od.
Exit: 0 = all selected images matched; 1 = failed validation; 2 = usage/setup error.

Examples:
  $(basename "$0") registry.k8s.io/autoscaling/cluster-autoscaler:v1.32.7
  $(basename "$0") --arches amd64,arm64 registry.k8s.io/autoscaling/cluster-autoscaler:v1.32.7
EOF
}

log()  { printf '%s\n' "$*" >&2; }
die()  { printf 'error: %s\n' "$*" >&2; exit 2; }

# Colors, only when stderr is a terminal.
if [ -t 2 ]; then
  C_RED=$'\033[31m'; C_GREEN=$'\033[32m'; C_YELLOW=$'\033[33m'; C_BOLD=$'\033[1m'; C_OFF=$'\033[0m'
else
  C_RED=""; C_GREEN=""; C_YELLOW=""; C_BOLD=""; C_OFF=""
fi

set_ref() {
  [[ -z $REF && -n $1 && $1 != -* ]] || die "unexpected image argument: $1"
  REF="$1"
}

parse_args() {
  local arg
  while (( $# )); do
    case "$1" in
      -a|--arches|--require-arches|-b|--binary|-r|--runtime)
        [[ -n ${2:-} && $2 != -* ]] || die "$1 requires a value"
        case "$1" in
          -a|--arches|--require-arches)
            [[ -z $ARCH_SPEC ]] || die 'choose only one architecture list'
            ARCH_SPEC="$2"
            if [[ $1 != --require-arches ]]; then MODE=spot; fi ;;
          -b|--binary) BINARY="$2" ;;
          -r|--runtime) RUNTIME="$2" ;;
        esac
        shift 2 ;;
      -k|--keep) KEEP=1; shift ;;
      -h|--help) usage; exit 0 ;;
      --) shift; break ;;
      -*) die "unknown option: $1 (see --help)" ;;
      *) set_ref "$1"; shift ;;
    esac
  done
  for arg in "$@"; do set_ref "$arg"; done
  [[ -n $REF ]] || { usage >&2; die 'missing <image-ref>'; }
  [[ $BINARY == /* ]] || die '--binary must be an absolute path in the image'
  if [[ $REF == *@* && ! ${REF##*@} =~ ^[a-z0-9]+([+._-][a-z0-9]+)*:[A-Za-z0-9=_-]+$ ]]; then
    die "invalid index digest in $REF"
  fi
}

detect_runtime() {
  if [[ -z $RUNTIME ]]; then
    if command -v docker >/dev/null; then RUNTIME=docker; else RUNTIME=podman; fi
  fi
  case "${RUNTIME##*/}" in
    docker|podman) ;;
    *) die "unsupported runtime: $RUNTIME" ;;
  esac
  command -v "$RUNTIME" >/dev/null || die "runtime '$RUNTIME' not found (need docker or podman)"
}

# Strip any tag and/or digest from a reference, preserving registry:port.
repo_no_tag() {
  local r="$1"
  r="${r%@*}"                 # drop @sha256:... digest if present
  local last="${r##*/}"       # component after the last slash
  if [ "${last#*:}" != "$last" ]; then
    echo "${r%:*}"            # last component has a ':', it's the tag -> strip it
  else
    echo "$r"
  fi
}

inspect_index() {
  if [[ ${RUNTIME##*/} == docker ]]; then
    "$RUNTIME" buildx imagetools inspect --raw "$REF" >"$1" 2>/dev/null && return 0
  fi
  "$RUNTIME" manifest inspect "$REF" >"$1"
}

# Emit one platform<TAB>digest per runnable Linux descriptor. Never select a
# single representative digest for an entire CPU family. Exit 3 for a single
# image, which can still be checked with --arches but is not a release index.
index_platforms() {
  python3 - "$@" <<'PY'
import json
import re
import sys

path, mode, *arches = sys.argv[1:]
required = set(arches)
seen = set()
indexes = {"application/vnd.oci.image.index.v1+json",
           "application/vnd.docker.distribution.manifest.list.v2+json"}
try:
    with open(path) as f:
        index = json.load(f)
    if "manifests" not in index and "config" in index and "layers" in index:
        sys.exit(3)
    if index.get("schemaVersion") != 2 or not isinstance(index.get("manifests"), list):
        raise ValueError("not a manifest list or OCI image index")
    for entry in index["manifests"]:
        platform = entry.get("platform", {})
        os = platform.get("os")
        if os and os != "linux":  # includes BuildKit's unknown/unknown attestations
            continue
        if entry.get("mediaType") in indexes:
            raise ValueError("nested image indexes are not supported")
        if not os and entry.get("artifactType"):
            continue  # non-platform OCI artifact
        arch = platform.get("architecture", "")
        variant = platform.get("variant", "")
        if not os or arch == "unknown" or not re.fullmatch(r"[a-z0-9][a-z0-9_.-]*", arch):
            raise ValueError("missing or invalid Linux platform in descriptor")
        if mode == "spot" and arch not in required:
            continue
        if variant and not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]*", variant):
            raise ValueError(f"invalid platform variant for linux/{arch}")
        digest = entry.get("digest", "")
        if not re.fullmatch(r"[a-z0-9]+(?:[+._-][a-z0-9]+)*:[A-Za-z0-9=_-]+", digest):
            raise ValueError(f"missing or invalid manifest digest for linux/{arch}")
        seen.add(arch)
        print("/".join(filter(None, (os, arch, variant))), digest, sep="\t")
    if required - seen:
        raise ValueError("index is missing required Linux architecture(s): " + ", ".join(sorted(required - seen)))
except (OSError, ValueError, TypeError, AttributeError) as err:
    sys.exit(f"error inspecting index: {err}")
PY
}

plan_platforms() {
  local index="$WORKDIR/state/index" rc
  if command -v python3 >/dev/null; then
    if inspect_index "$index"; then
      index_platforms "$index" "$MODE" "$@" && return 0
      rc=$?
      (( rc == 3 )) || return 1  # Malformed or incomplete index: never fall back.
    fi
  elif [[ $MODE == release ]]; then
    die 'python3 is required to inspect the release index'
  fi
  if [[ $MODE != spot ]]; then
    log 'FAIL: release validation requires an inspectable multi-platform index'
    return 1
  fi
  log "${C_YELLOW}warning:${C_OFF} no inspectable index; spot-checking requested platforms only"
  printf 'linux/%s\n' "$@"
}

# od -v prevents repeated zero rows from being compressed to '*'. Read enough
# bytes for a full ELF64 header (64 bytes); ELF32 headers only need 52.
elf_arch() {
  local dump size size_offset lo hi machine kind
  local -a b
  dump=$(LC_ALL=C od -An -v -tu1 -N64 "$1") || return 1
  read -r -a b <<< "${dump//$'\n'/ }"
  if [[ ${b[0]:-} != 127 || ${b[1]:-} != 69 || ${b[2]:-} != 76 || ${b[3]:-} != 70 || ${b[6]:-} != 1 ]]; then
    log 'not a valid ELF header'; return 1
  fi
  case ${b[4]} in
    1) size=52; size_offset=40 ;;
    2) size=64; size_offset=52 ;;
    *) log 'invalid ELF class'; return 1 ;;
  esac
  (( ${#b[@]} >= size )) || { log 'truncated ELF header'; return 1; }
  case ${b[5]} in
    1) lo=0; hi=1 ;;
    2) lo=1; hi=0 ;;
    *) log 'invalid ELF byte order'; return 1 ;;
  esac
  (( b[size_offset+lo] + 256*b[size_offset+hi] == size )) || { log 'invalid ELF header size'; return 1; }
  kind=$(( b[16+lo] + 256*b[16+hi] ))
  (( kind == 2 || kind == 3 )) || { log 'expected an ET_EXEC or ET_DYN ELF file'; return 1; }
  machine=$(( b[18+lo] + 256*b[18+hi] ))
  case "$machine:${b[4]}:${b[5]}" in  # e_machine : EI_CLASS : EI_DATA
    3:1:1)   echo 386 ;;
    20:1:2)  echo ppc ;;
    21:2:1)  echo ppc64le ;;
    21:2:2)  echo ppc64 ;;
    22:2:2)  echo s390x ;;
    40:1:1)  echo arm ;;
    62:2:1)  echo amd64 ;;
    183:2:1) echo arm64 ;;
    243:2:1) echo riscv64 ;;
    *) log "unsupported ELF architecture (machine=$machine, class=${b[4]}, data=${b[5]})"; return 1 ;;
  esac
}

cleanup() {
  local cidfile cid
  for cidfile in "$WORKDIR"/state/*.cid; do
    [[ -s $cidfile ]] || continue
    cid=$(<"$cidfile")
    if [[ $cid =~ ^[0-9a-f]{64}$ ]]; then
      "$RUNTIME" rm "$cid" >/dev/null || log "warning: could not remove container $cid"
    fi
  done
  if (( KEEP )); then
    rm -rf "$WORKDIR/state"
    log "Extracted binaries kept in: $WORKDIR"
  else
    rm -rf "$WORKDIR"
  fi
}

validate_platform() {
  local platform="$1" ref="$2" n="$3" arch="${1#linux/}" cid
  local out="$WORKDIR/cluster-autoscaler-${platform//\//-}-$n"
  local cidfile="$WORKDIR/state/$n.cid"
  log "Checking $platform: $ref"
  "$RUNTIME" pull --platform "$platform" "$ref" >/dev/null || return 1
  # create never starts the image. Its stdout is the ID of our container;
  # persist it for the EXIT trap even if cp or the header check is interrupted.
  "$RUNTIME" create --platform "$platform" "$ref" >"$cidfile" || return 1
  cid=$(<"$cidfile")
  [[ $cid =~ ^[0-9a-f]{64}$ ]] || { log 'runtime did not return a container ID'; return 1; }
  "$RUNTIME" cp "$cid:$BINARY" "$out" >/dev/null || return 1
  [[ -f $out && ! -L $out ]] || { log "expected a regular file at $BINARY"; return 1; }
  DETECTED=$(elf_arch "$out") || return 1
  [[ $DETECTED == "${arch%%/*}" ]]
}

main() {
  parse_args "$@"
  detect_runtime
  command -v od >/dev/null || die "'od' is required"
  if [[ -z $ARCH_SPEC ]]; then
    ARCH_SPEC=$(make -s --no-print-directory -C "$(dirname "${BASH_SOURCE[0]}")/.." print-release-arches) ||
      die 'could not read ALL_ARCH from the release Makefile; supply --require-arches or --arches'
  fi
  local -a arches rows
  local arch platform digest repo ref result checked=0 failures=0
  ARCH_SPEC=${ARCH_SPEC//,/ }
  read -r -a arches <<< "${ARCH_SPEC//$'\n'/ }"
  [[ -n ${arches[0]:-} ]] || die 'empty architecture list'
  for arch in "${arches[@]}"; do
    [[ $arch =~ ^[a-z0-9][a-z0-9_.-]*$ ]] || die "invalid architecture: $arch"
  done

  WORKDIR=$(mktemp -d "${TMPDIR:-/tmp}/ca-multiarch-validate.XXXXXX") || die 'could not create temp directory'
  trap cleanup EXIT
  trap 'exit 129' HUP; trap 'exit 130' INT; trap 'exit 143' TERM
  mkdir "$WORKDIR/state"
  local plan="$WORKDIR/state/platforms"
  log "Image: $REF ($MODE check using $RUNTIME)"
  plan_platforms "${arches[@]}" >"$plan" || return 1

  repo=$(repo_no_tag "$REF")
  while IFS=$'\t' read -r platform digest; do
    ref=$REF
    if [[ -n $digest ]]; then ref="$repo@$digest"; fi
    checked=$((checked + 1))
    DETECTED=''
    if validate_platform "$platform" "$ref" "$checked" </dev/null; then
      result="${C_GREEN}PASS${C_OFF}"
    else
      result="${C_RED}FAIL${C_OFF}"
      failures=$((failures + 1))
    fi
    rows+=("$(printf '  %-20s %-12s %s' "$platform" "${DETECTED:--}" "$result")")
  done <"$plan"
  (( checked )) || { log 'FAIL: no platforms checked'; return 1; }
  log "${C_BOLD}Summary for $REF${C_OFF}"
  printf '  %-20s %-12s %s\n' PLATFORM BINARY-IS RESULT >&2
  printf '%s\n' "${rows[@]}" >&2
  (( failures == 0 )) || { log "FAILED: $failures/$checked platform image(s) mismatched or unverifiable"; return 1; }
  log "OK: verified $checked platform image(s) for $REF ($MODE check)"
}

main "$@"
