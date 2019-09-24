#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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

function check_obsolete_cluster_role()
{
    kubectl get clusterrole system:admission-controller &> /dev/null
    OBSOLETE_CLUSTERROLE_EXISTS=$?
    if [[ $ret -eq 0 ]]; then
        kubectl get clusterrole system:admission-controller -o yaml 2>&1 | grep verticalpodautoscalers &> /dev/null
        OBSOLETE_CLUSTERROLE_EXISTS=$?
    fi

    kubectl get clusterrolebinding system:admission-controller &> /dev/null
    OBSOLETE_CLUSTERROLE_BINDING_EXISTS=$?
}

check_obsolete_cluster_role

if [[ ${OBSOLETE_CLUSTERROLE_EXISTS} -eq 0 ]]; then
    echo
    echo "Older version of vpa-up.sh creates a ClusterRole object named system:admission-controller"
    echo "that is now renamed to system:vpa-admission-controller to avoid confusion.  A ClusterRole"
    echo "object of the same name still exists in the cluster and it appears to be created by vpa-up.sh."
    echo "Please inspect the object and delete manually if it is created by vpa-up.sh"
    echo
    echo "You can inspect the object content by running"
    echo
    echo "kubectl get clusterrole system:admission-controller -o yaml"
    echo
fi

if [[ ${OBSOLETE_CLUSTERROLE_BINDING_EXISTS} -eq 0 ]]; then
    echo
    echo "Older version of vpa-up.sh creates a ClusterRoleBinding object named system:admission-controller"
    echo "that is now renamed to system:vpa-admission-controller to avoid confusion.  A ClusterRoleBinding"
    echo "object of the same name still exists in the cluster."
    echo "Please inspect the object and delete manually if it is created by vpa-up.sh"
    echo
    echo "You can inspect the object content by running"
    echo
    echo "kubectl get clusterrolebinding system:admission-controller -o yaml"
    echo
fi
