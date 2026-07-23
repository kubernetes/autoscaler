#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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
# Fallback arch list when the manifest cannot be enumerated. Mirrors ALL_ARCH in
# cluster-autoscaler/Makefile.
readonly DEFAULT_ARCHES="amd64 arm64 s390x"
# Architectures we always expect a release image to publish (used only to warn
# about a missing arch). Mirrors ALL_ARCH in cluster-autoscaler/Makefile.
readonly EXPECTED_ARCHES="amd64 arm64 s390x"

BINARY="${DEFAULT_BINARY}"
RUNTIME="${CONTAINER_RUNTIME:-}"
ARCHES_OVERRIDE=""
KEEP=0
REF=""
WORKDIR=""

usage() {
  cat <<EOF
Validate that every architecture of a multi-arch cluster-autoscaler image
contains a binary actually compiled for that architecture.

Usage:
  $(basename "$0") [options] <image-ref>

Arguments:
  <image-ref>   Fully-qualified image reference (a tag or an index digest), e.g.
                  registry.k8s.io/autoscaling/cluster-autoscaler:v1.32.7
                  gcr.io/k8s-staging-autoscaling/cluster-autoscaler:v1.32.7

Options:
  -a, --arches "amd64 arm64 s390x"
                Space/comma separated arch list to validate. Default: auto-detect
                from the image's manifest list, falling back to "${DEFAULT_ARCHES}".
  -b, --binary PATH
                Path of the binary inside the image. Default: "${DEFAULT_BINARY}".
  -r, --runtime docker|podman
                Container runtime to use. Default: auto-detect (\$CONTAINER_RUNTIME,
                then docker, then podman).
  -k, --keep    Keep the extracted binaries (in a temp dir) instead of deleting them.
  -h, --help    Show this help.

Exit status:
  0  all checked architectures match
  1  at least one architecture mismatched / could not be validated
  2  usage or environment error

Examples:
  # Validate the staging image before it is promoted:
  $(basename "$0") gcr.io/k8s-staging-autoscaling/cluster-autoscaler:v1.32.7

  # Validate a promoted release image, only amd64 and arm64:
  $(basename "$0") -a "amd64 arm64" registry.k8s.io/autoscaling/cluster-autoscaler:v1.32.7
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

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      -a|--arches)  ARCHES_OVERRIDE="${2:-}"; shift 2 ;;
      -b|--binary)  BINARY="${2:-}"; shift 2 ;;
      -r|--runtime) RUNTIME="${2:-}"; shift 2 ;;
      -k|--keep)    KEEP=1; shift ;;
      -h|--help)    usage; exit 0 ;;
      --)           shift; break ;;
      -*)           die "unknown option: $1 (see --help)" ;;
      *)            if [ -n "$REF" ]; then die "unexpected extra argument: $1"; fi; REF="$1"; shift ;;
    esac
  done
  [ $# -gt 0 ] && [ -z "$REF" ] && REF="$1"
  [ -n "$REF" ] || { usage >&2; die "missing <image-ref>"; }
}

detect_runtime() {
  if [ -n "$RUNTIME" ]; then
    command -v "$RUNTIME" >/dev/null 2>&1 || die "requested runtime '$RUNTIME' not found in PATH"
    return
  fi
  if command -v docker >/dev/null 2>&1; then RUNTIME=docker
  elif command -v podman >/dev/null 2>&1; then RUNTIME=podman
  else die "no container runtime found (need docker or podman)"; fi
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

# Echo the raw manifest-list / index JSON for a ref (best effort; may be empty).
inspect_index_json() {
  local ref="$1"
  case "$RUNTIME" in
    docker)
      docker buildx imagetools inspect --raw "$ref" 2>/dev/null && return 0
      docker manifest inspect "$ref" 2>/dev/null && return 0
      ;;
    podman)
      podman manifest inspect "$ref" 2>/dev/null && return 0
      ;;
  esac
  return 1
}

# List linux architectures present in the manifest list JSON on stdin.
detect_arches() {
  jq -r '.manifests[]?
           | select(.platform.os == "linux")
           | select(.platform.architecture != null and .platform.architecture != "unknown")
           | .platform.architecture' 2>/dev/null | sort -u
}

# Echo the manifest digest for a given linux arch from the index JSON on stdin.
digest_for_arch() {
  jq -r --arg a "$1" '.manifests[]?
           | select(.platform.os == "linux" and .platform.architecture == $a)
           | .digest' 2>/dev/null | head -n1
}

# Read the ELF header of a file and set the globals:
#   ELF_OK       1 if the file is an ELF binary, else 0
#   ELF_CLASS    1 (32-bit) or 2 (64-bit)
#   ELF_DATA     1 (little-endian) or 2 (big-endian)
#   ELF_MACHINE  e_machine value (decimal)
# Uses od so it does not depend on `file` and works identically on macOS/Linux.
elf_read() {
  local f="$1" magic
  ELF_OK=0; ELF_CLASS=""; ELF_DATA=""; ELF_MACHINE=""
  magic="$(od -An -tx1 -N4 "$f" 2>/dev/null | tr -d ' \n')"
  [ "$magic" = "7f454c46" ] || return 0
  ELF_OK=1
  ELF_CLASS="$(od -An -tu1 -j4  -N1 "$f" | tr -d ' ')"
  ELF_DATA="$(od  -An -tu1 -j5  -N1 "$f" | tr -d ' ')"
  local b18 b19
  b18="$(od -An -tu1 -j18 -N1 "$f" | tr -d ' ')"
  b19="$(od -An -tu1 -j19 -N1 "$f" | tr -d ' ')"
  if [ "$ELF_DATA" = "2" ]; then
    ELF_MACHINE=$(( b18 * 256 + b19 ))   # big-endian (e.g. s390x)
  else
    ELF_MACHINE=$(( b19 * 256 + b18 ))   # little-endian
  fi
}

# Map an ELF e_machine value (+ endianness) to a Go GOARCH name.
arch_from_machine() {
  local m="$1" data="$2"
  case "$m" in
    3)   echo "386" ;;
    20)  echo "ppc" ;;
    21)  if [ "$data" = "1" ]; then echo "ppc64le"; else echo "ppc64"; fi ;;
    22)  echo "s390x" ;;
    40)  echo "arm" ;;
    62)  echo "amd64" ;;
    183) echo "arm64" ;;
    243) echo "riscv64" ;;
    "")  echo "unknown" ;;
    *)   echo "machine-$m" ;;
  esac
}

cleanup() {
  if [ -n "$WORKDIR" ] && [ -d "$WORKDIR" ]; then
    if [ "$KEEP" = "1" ]; then
      log "Extracted binaries kept in: $WORKDIR"
    else
      rm -rf "$WORKDIR"
    fi
  fi
}

# Validate a single architecture. Returns 0 on match, 1 otherwise.
validate_arch() {
  local arch="$1"
  local platform="linux/${arch}"
  local pull_ref="$REF" digest=""
  local cname="ca-validate-${arch}-$$"
  local out="${WORKDIR}/cluster-autoscaler-${arch}"

  if [ -n "$INDEX_JSON" ]; then
    digest="$(printf '%s' "$INDEX_JSON" | digest_for_arch "$arch")"
  fi
  if [ -n "$digest" ]; then
    pull_ref="$(repo_no_tag "$REF")@${digest}"
  fi

  log ""
  log "${C_BOLD}== ${arch} (${platform})${C_OFF}"
  [ -n "$digest" ] && log "   manifest digest: ${digest}"
  log "   image:           ${pull_ref}"

  if ! "$RUNTIME" pull --platform "$platform" "$pull_ref" >/dev/null 2>&1; then
    log "   ${C_RED}pull failed${C_OFF} (no image for ${platform}?)"
    return 1
  fi

  "$RUNTIME" rm -f "$cname" >/dev/null 2>&1 || true
  if ! "$RUNTIME" create --platform "$platform" --name "$cname" "$pull_ref" >/dev/null 2>&1; then
    # Fall back without --platform (we may have pulled by exact digest already).
    if ! "$RUNTIME" create --name "$cname" "$pull_ref" >/dev/null 2>&1; then
      log "   ${C_RED}create failed${C_OFF}"
      return 1
    fi
  fi

  local cp_rc=0
  "$RUNTIME" cp "${cname}:${BINARY}" "$out" >/dev/null 2>&1 || cp_rc=$?
  "$RUNTIME" rm -f "$cname" >/dev/null 2>&1 || true
  if [ "$cp_rc" -ne 0 ]; then
    log "   ${C_RED}cp failed${C_OFF} (is '${BINARY}' present in the image?)"
    return 1
  fi

  # Informational: show `file` output, mirroring the issue's reproduction.
  if command -v file >/dev/null 2>&1; then
    log "   file:            $(file -b "$out")"
  fi

  elf_read "$out"
  if [ "$ELF_OK" != "1" ]; then
    log "   ${C_RED}FAIL${C_OFF}: '${BINARY}' is not an ELF binary"
    return 1
  fi

  local detected
  detected="$(arch_from_machine "$ELF_MACHINE" "$ELF_DATA")"
  log "   ELF e_machine:   ${ELF_MACHINE} -> ${detected} (class=${ELF_CLASS}, data=${ELF_DATA})"

  if [ "$detected" = "$arch" ]; then
    log "   ${C_GREEN}PASS${C_OFF}: binary architecture matches ${platform}"
    RESULT_ROWS="${RESULT_ROWS}${arch}|${detected}|PASS"$'\n'
    return 0
  fi

  log "   ${C_RED}FAIL${C_OFF}: expected ${arch} but binary is ${detected}"
  RESULT_ROWS="${RESULT_ROWS}${arch}|${detected}|FAIL"$'\n'
  return 1
}

main() {
  parse_args "$@"
  detect_runtime
  command -v od >/dev/null 2>&1 || die "'od' is required but not found in PATH"

  log "Runtime:   ${RUNTIME}"
  log "Image:     ${REF}"
  log "Binary:    ${BINARY}"

  # Enumerate the published manifest list (best effort).
  INDEX_JSON=""
  if command -v jq >/dev/null 2>&1; then
    INDEX_JSON="$(inspect_index_json "$REF" || true)"
  else
    log "${C_YELLOW}note:${C_OFF} 'jq' not found; skipping manifest inspection (digests/auto-detect disabled)"
  fi

  # Decide which architectures to validate.
  local arches=""
  if [ -n "$ARCHES_OVERRIDE" ]; then
    arches="$(printf '%s' "$ARCHES_OVERRIDE" | tr ',' ' ')"
  elif [ -n "$INDEX_JSON" ]; then
    arches="$(printf '%s' "$INDEX_JSON" | detect_arches | tr '\n' ' ')"
  fi
  if [ -z "${arches// }" ]; then
    log "${C_YELLOW}note:${C_OFF} could not determine architectures from manifest; using default '${DEFAULT_ARCHES}'"
    arches="$DEFAULT_ARCHES"
  fi

  log "Arches:    ${arches}"

  # Warn about any expected arch that is not advertised in the manifest at all.
  if [ -n "$INDEX_JSON" ]; then
    local published; published="$(printf '%s' "$INDEX_JSON" | detect_arches | tr '\n' ' ')"
    local want
    for want in $EXPECTED_ARCHES; do
      case " $published " in
        *" $want "*) ;;
        *) log "${C_YELLOW}warning:${C_OFF} expected arch '${want}' is not published in the manifest list" ;;
      esac
    done
  fi

  WORKDIR="$(mktemp -d "${TMPDIR:-/tmp}/ca-multiarch-validate.XXXXXX")"
  trap cleanup EXIT

  RESULT_ROWS=""
  local checked=0 failures=0 arch
  for arch in $arches; do
    checked=$((checked + 1))
    if ! validate_arch "$arch"; then
      failures=$((failures + 1))
    fi
  done

  # Summary table.
  log ""
  log "${C_BOLD}Summary for ${REF}${C_OFF}"
  printf '  %-10s %-12s %s\n' "ARCH" "BINARY-IS" "RESULT" >&2
  printf '%s' "$RESULT_ROWS" | while IFS='|' read -r a d r; do
    [ -n "$a" ] || continue
    local color="$C_GREEN"; [ "$r" = "FAIL" ] && color="$C_RED"
    printf '  %-10s %-12s %s%s%s\n' "$a" "$d" "$color" "$r" "$C_OFF" >&2
  done

  if [ "$checked" -eq 0 ]; then
    log "${C_RED}No architectures were validated.${C_OFF}"
    return 1
  fi
  if [ "$failures" -ne 0 ]; then
    log ""
    log "${C_RED}${C_BOLD}FAILED${C_OFF}: ${failures}/${checked} architecture(s) are mismatched or unverifiable."
    return 1
  fi
  log ""
  log "${C_GREEN}${C_BOLD}OK${C_OFF}: all ${checked} architecture(s) contain a correctly-built binary."
  return 0
}

main "$@"
