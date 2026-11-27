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

# Exercise the validator as a CLI. No registry, daemon, or emulation is needed.
set -euo pipefail
hack=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
validator="$hack/validate-multiarch-image.sh"
repo=localhost:5050/test/cluster-autoscaler
tag="$repo:candidate"
root=$(mktemp -d "${TMPDIR:-/tmp}/ca-image-test.XXXXXX")
trap 'rm -rf "$root"' EXIT
mkdir "$root/bin" "$root/containers" "$root/work" "$root/keep"
export FAKE_ROOT="$root" FAKE_FAIL='' CONTAINER_RUNTIME=docker TMPDIR="$root/work"
export PATH="$root/bin:$PATH"
unset MAKEFLAGS MFLAGS

# Generate independent, byte-exact ELF32/ELF64 headers and registry responses.
python3 - "$root" <<'PY'
import hashlib
import json
from pathlib import Path
import struct
import sys

root = Path(sys.argv[1])

def elf(machine, bits=64, endian="little", kind=2):
    ident = b"\x7fELF" + bytes((2 if bits == 64 else 1, 1 if endian == "little" else 2, 1)) + bytes(9)
    layout = ("<" if endian == "little" else ">") + ("HHIQQQIHHHHHH" if bits == 64 else "HHIIIIIHHHHHH")
    return ident + struct.pack(layout, kind, machine, 1, 0, 0, 0, 0, 16 + struct.calcsize(layout), 0, 0, 0, 0, 0)

def image(name, arch, binary, variant=None, os="linux"):
    digest = "sha256:" + hashlib.sha256(name.encode()).hexdigest()
    (root / digest).write_bytes(binary)
    (root / (name + ".digest")).write_text(digest)
    platform = {"os": os, "architecture": arch}
    if variant:
        platform["variant"] = variant
    return {"mediaType": "application/vnd.oci.image.manifest.v1+json", "digest": digest, "platform": platform}

def index(name, *entries):
    (root / (name + ".json")).write_text(json.dumps({"schemaVersion": 2, "manifests": entries}))

amd = image("amd", "amd64", elf(62))
arm = image("arm", "arm64", elf(183, kind=3), "v8")
s390 = image("s390", "s390x", elf(22, endian="big"))
index("good", amd, arm, s390, image("attestation", "unknown", b"", os="unknown"), image("win", "amd64", b"", os="windows"))
index("partial", amd)
index("variants", amd, arm, image("bad-v9", "arm64", elf(62), "v9"))
index("missing-digest", {"platform": amd["platform"]})
for name, header in {"short": elf(62)[:20], "short63": elf(62)[:63], "x32": elf(62, bits=32),
                     "endian": elf(62, endian="big"), "object": elf(62, kind=1), "swapped": elf(183), "text": b"not ELF"}.items():
    index(name, image(name, "amd64", header))
(root / "single.json").write_text(json.dumps({"schemaVersion": 2, "config": {}, "layers": []}))
(root / "linux-amd64").write_bytes(elf(62))  # tag-based spot check without an index
PY

cat >"$root/bin/docker" <<'RUNTIME'
#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >>"$FAKE_ROOT/calls"
unexpected() { echo "$*" >>"$FAKE_ROOT/unexpected"; exit 1; }
case "$1 ${2:-}" in
  'buildx imagetools'|'manifest inspect') [[ $FAKE_FAIL != inspect ]] && cat "$FAKE_INDEX" ;;
  'manifest create') : ;;
  'manifest push') printf 'Pushed child manifest\nsha256:%064x\n' 42; [[ $FAKE_FAIL != push ]] ;;
  'pull --platform'|'create --platform')
    [[ $# == 4 ]] || unexpected "$@"
    [[ $FAKE_FAIL != "$1" ]] || exit 1
    if [[ $4 == *@* ]]; then key=${4##*@}; else key=${3//\//-}; fi
    [[ -f $FAKE_ROOT/$key ]] || exit 1
    if [[ $1 == create ]]; then
      cid=$(printf '%064x' "$$")
      printf '%s\n' "$key" >"$FAKE_ROOT/containers/$cid"
      echo "$cid"
    fi ;;
  'cp '*)
    [[ $FAKE_FAIL != cp ]] || exit 1
    if [[ $FAKE_FAIL == interrupt-cp ]]; then kill -TERM "$PPID"; exit 1; fi
    [[ ${2#*:} == /cluster-autoscaler ]] || unexpected "$@"
    key=$(cat "$FAKE_ROOT/containers/${2%%:*}")
    cp "$FAKE_ROOT/$key" "$3" ;;
  'rm '*)
    [[ $# == 2 && $2 =~ ^[0-9a-f]{64}$ && -f $FAKE_ROOT/containers/$2 ]] || unexpected "$@"
    rm "$FAKE_ROOT/containers/$2" ;;
  *) unexpected "$@" ;;
esac
RUNTIME
chmod +x "$root/bin/docker"
ln -s docker "$root/bin/podman"

fail() { cat "$root/output" >&2; echo "FAIL: $*" >&2; exit 1; }
# check <index fixture> <expected exit> <command> [args...]
check() {
  local fixture=$1 expected=$2 rc=0 container
  shift 2
  : >"$root/calls"; : >"$root/unexpected"
  FAKE_INDEX="$root/$fixture.json" "$@" >"$root/output" 2>&1 || rc=$?
  (( rc == expected )) || fail "$fixture: expected exit $expected, got $rc from $*"
  [[ ! -s $root/unexpected ]] || fail "unexpected runtime call: $(cat "$root/unexpected")"
  for container in "$root"/containers/*; do
    [[ ! -e $container ]] || fail "leaked test container: $container"
  done
}
called() { grep -Fxq -- "$*" "$root/calls" || fail "runtime did not receive: $*"; }
no_pulls() { if grep -q '^pull ' "$root/calls"; then fail 'pulled before validating the index'; fi; }

pinned="$repo@sha256:$(printf '%064x' 42)"
for runtime in docker podman; do
  check good 0 "$validator" -r "$runtime" "$tag@${pinned##*@}"
  called pull --platform linux/amd64 "$repo@$(cat "$root/amd.digest")"
  called pull --platform linux/arm64/v8 "$repo@$(cat "$root/arm.digest")"
  called pull --platform linux/s390x "$repo@$(cat "$root/s390.digest")"
  [[ $(grep -c '^pull ' "$root/calls") == 3 ]] || fail 'expected only runnable Linux images'
done
check partial 1 "$validator" "$tag"; no_pulls
check partial 0 "$validator" --arches amd64 "$tag"
check partial 1 "$validator" --arches amd64,arm64 "$tag"; no_pulls
check good 2 "$validator" --arches
check good 2 "$validator" -- "$tag" unexpected
check good 2 "$validator" --arches amd64 --require-arches amd64 "$tag"

# Even an additional (non-required) architecture must have every variant checked.
TMPDIR="$root/keep" check variants 1 "$validator" --require-arches amd64 --keep "$tag"
called pull --platform linux/arm64/v9 "$repo@$(cat "$root/bad-v9.digest")"
grep -Eq 'linux/arm64/v9 +amd64 +FAIL' "$root/output" || fail 'wrong variant absent from summary'
kept=("$root"/keep/*/cluster-autoscaler-*)
(( ${#kept[@]} == 3 )) || fail 'extracted variant files overwritten'

for fixture in short short63 x32 endian object swapped text; do
  check "$fixture" 1 "$validator" --arches amd64 "$tag"
  grep -Eq 'linux/amd64 +[^ ]+ +FAIL' "$root/output" || fail "$fixture: missing failure row"
done
for failure in inspect ''; do
  FAKE_FAIL=$failure check single 1 "$validator" --require-arches amd64 "$tag"; no_pulls
  FAKE_FAIL=$failure check single 0 "$validator" -a amd64 "$tag"
  called pull --platform linux/amd64 "$tag"
done
check missing-digest 1 "$validator" --arches amd64 "$tag"; no_pulls
for failure in pull create cp; do FAKE_FAIL=$failure check good 1 "$validator" -a amd64 "$tag"; done
FAKE_FAIL=interrupt-cp check good 143 "$validator" -a amd64 "$tag"
for dir in "$TMPDIR"/ca-multiarch-validate.*; do
  [[ ! -d $dir ]] || fail "leaked validator temp directory: $dir"
done

# Exercise the real Make recipe, but fake every Docker operation, including push.
push=(make -s -C "$hack/.." push-manifest REGISTRY=localhost:5050/test TAG=candidate)
check partial 0 "${push[@]}" ALL_ARCH=amd64
called buildx imagetools inspect --raw "$pinned"
check partial 2 "${push[@]}" 'ALL_ARCH=amd64 arm64'; no_pulls
FAKE_FAIL=push check partial 2 "${push[@]}" ALL_ARCH=amd64
if grep -q '^buildx ' "$root/calls"; then fail 'validated after an unsuccessful push'; fi

echo 'PASS: multi-architecture image validation regression tests'
