#!/bin/bash

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

# Runs the given command, retrying with linear backoff on failure. Useful for
# wrapping flaky network operations such as `docker push`.
#
# Usage:
#   hack/retry.sh <command> [args...]
#
# Configuration via environment variables:
#   RETRY_ATTEMPTS  Maximum number of attempts (default: 5).
#   RETRY_BACKOFF   Base backoff in seconds; attempt N waits N*RETRY_BACKOFF
#                   seconds before retrying (default: 10).

set -o nounset
set -o pipefail

RETRY_ATTEMPTS="${RETRY_ATTEMPTS:-5}"
RETRY_BACKOFF="${RETRY_BACKOFF:-10}"

if [ "$#" -eq 0 ]; then
  echo "ERROR: hack/retry.sh requires a command to run" >&2
  exit 1
fi

attempt=1
while true; do
  if "$@"; then
    exit 0
  fi
  if [ "${attempt}" -ge "${RETRY_ATTEMPTS}" ]; then
    echo "Command failed after ${RETRY_ATTEMPTS} attempts: $*" >&2
    exit 1
  fi
  delay=$((attempt * RETRY_BACKOFF))
  echo "Command failed (attempt ${attempt}/${RETRY_ATTEMPTS}); retrying in ${delay}s: $*" >&2
  sleep "${delay}"
  attempt=$((attempt + 1))
done
