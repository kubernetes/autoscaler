# /*
# Copyright (c) 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#      http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# */


#HOW TO CALL 
#```````````````````````````````````````````````````````````````````````````````````````````

# ./local_setup.sh --PROJECT <project-name> --SEED <seed-name> --SHOOT <shoot-name>

#```````````````````````````````````````````````````````````````````````````````````````````
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

while [ $# -gt 0 ]; do

   if [[ $1 == *"--"* ]]; then
        v="${1/--/}"
        declare $v="$2"
   fi

  shift
done

CURRENT_DIR=$(pwd)
PROJECT_ROOT="${CURRENT_DIR}"
KUBECONFIG_PATH=$PROJECT_ROOT/dev/kubeconfigs

#setting kuebconfig of control cluster

gardenctl target --garden sap-landscape-dev --project garden

eval $(gardenctl kubectl-env bash)

echo "$(kubectl get secret/$SEED.kubeconfig --template={{.data.kubeconfig}} | base64 -d)" > $KUBECONFIG_PATH/kubeconfig_control.yaml

#setting up the target config

gardenctl target --garden sap-landscape-dev --project $PROJECT --shoot $SHOOT --control-plane

eval $(gardenctl kubectl-env bash)

SECRET=$(kubectl get secrets | awk '{ print $1 }' |grep user-kubeconfig-)

echo "$(kubectl get secret/$SECRET -n shoot--$PROJECT--$SHOOT --template={{.data.kubeconfig}} | base64 -d)" > $KUBECONFIG_PATH/kubeconfig_target.yaml

# All the kubeconfigs are at place

echo "kubeconfigs have been downloaded and kept at /dev/kubeconfigs/kubeconfig_<target/control>.yaml"

export CONTROL_NAMESPACE=shoot--$PROJECT--$SHOOT
export CONTROL_KUBECONFIG=$KUBECONFIG_PATH/kubeconfig_control.yaml
export TARGET_KUBECONFIG=$KUBECONFIG_PATH/kubeconfig_target.yaml
