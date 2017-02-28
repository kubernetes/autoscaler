#!/bin/sh

# Copyright 2017 The Kubernetes Authors.
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

# A wrapper script trapping SIGTERM (docker stop) and passing the signal to
# cluster-autoscaler binary.

if [ -z "$LOG_OUTPUT" ]; then
  LOG_OUTPUT="/var/log/cluster_autoscaler.log"
fi

./cluster-autoscaler $@ 1>>$LOG_OUTPUT 2>&1 &
pid="$!"
trap "kill -15 $pid" 15

# We need a loop here, because receiving signal breaks out of wait.
# kill -0 doesn't send any signal, but it still checks if the process is running.
while kill -0 $pid > /dev/null 2>&1; do
  wait $pid
done
exit "$?"
