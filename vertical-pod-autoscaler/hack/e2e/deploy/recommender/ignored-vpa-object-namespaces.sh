#!/bin/bash

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

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/../../../..

export REGISTRY=${REGISTRY:-localhost:5001}
export TAG=${TAG:-latest}


REGISTRY=${REGISTRY} TAG=${TAG} ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh ${SCRIPT_ROOT}/hack/e2e/deploy/recommender/ignored-vpa-object-namespaces.yaml | kubectl apply -f -

CHECK_CYCLES=15
CHECK_SLEEP=15

DEPLOYED=0

for i in $(seq 1 $CHECK_CYCLES); do
    if [[ $(kubectl get deployments vpa-recommender -n kube-system -o jsonpath='{.status.unavailableReplicas}') -eq 0 ]]; then
        echo "Verified vpa-recommender pod was scheduled and started!"
        DEPLOYED=1
        break
    fi
    echo "Assertion Loop $i/$CHECK_CYCLES, sleeping for $CHECK_SLEEP seconds"
    sleep $CHECK_SLEEP
done

if [[ $DEPLOYED -eq 0 ]]; then
    exit 1
fi

EXIT_STATUS=2
for i in $(seq 1 $CHECK_CYCLES); do
    if kubectl describe vpa -n included-namespace included-vpa | grep Status; then
        echo "Verified included namespaces are being tracked!"
        EXIT_STATUS=0
        break
    fi
    echo "Assertion Loop $i/$CHECK_CYCLES, sleeping for $CHECK_SLEEP seconds"
    sleep $CHECK_SLEEP
done

if [[ $EXIT_STATUS -ne 0 ]];then
    exit $EXIT_STATUS
fi

CHECK_CYCLES=5

for i in $(seq 1 $CHECK_CYCLES); do
    if kubectl describe vpa -n ignored-namespace ignored-vpa | grep Status; then
        echo "Failed! ignored namespaces should not be tracked!"
        EXIT_STATUS=3
        break
    fi
    echo "Assertion Loop $i/$CHECK_CYCLES, sleeping for $CHECK_SLEEP seconds"
    sleep $CHECK_SLEEP
done

if [[ $EXIT_STATUS -ne 0 ]];then
    exit $EXIT_STATUS
fi

echo "Verified ignored namespaces are still not tracked after wait time!"

REGISTRY=${REGISTRY} TAG=${TAG} ${SCRIPT_ROOT}/hack/vpa-process-yaml.sh ${SCRIPT_ROOT}/hack/e2e/deploy/recommender/ignored-vpa-object-namespaces.yaml | kubectl delete -f -
