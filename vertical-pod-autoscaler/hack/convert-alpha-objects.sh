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

outdir=`mktemp -d --tmpdir vpa-alpha-XXXXXXXXX`

kubectl get verticalpodautoscalers.poc.autoscaling.k8s.io --all-namespaces --no-headers -o=custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name >${outdir}/list
if [[ -z `cat ${outdir}/list` ]]
then
  echo "No alpha VPA objects to store"
  exit 0
fi

echo "Fetching alpha VPA objects into ${outdir}."
while read ns name
do
  outfile=`mktemp --tmpdir=${outdir} vpa-XXXXXXXXX.yaml`
  echo "Storing converted ${ns}:${name} into ${outfile}."
  kubectl get verticalpodautoscalers.poc.autoscaling.k8s.io -n ${ns} ${name} -o yaml >${outfile}
  sed -i -e 's|poc.autoscaling.k8s.io/v1alpha1|autoscaling.k8s.io/v1beta1|' ${outfile}
done <${outdir}/list
rm ${outdir}/list

echo
echo "Please have a look at converted VPA objects in ${outdir}."
echo "If everything looks OK you can migrate to beta VPA by:"
echo "1. disabling alpha VPA via vpa-down.sh script,"
echo "2. enabling beta VPA via vpa-up.sh script,"
echo "3. re-creating VPA objects by executing:"
echo "   kubectl create -f ${outdir}"
echo
echo "NOTE: The recommendations will NOT be kept between versions."
echo "There will be a disruption period until beta VPA computes new recommendations."
echo

