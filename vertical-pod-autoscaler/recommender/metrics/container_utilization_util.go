/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	clientapiv1 "k8s.io/client-go/pkg/api/v1"
	k8sapiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// type conversions are needed, since ResourceList is coming from different packages in containerUsageSnapshot and containerSpec
func convertResourceListToClientApi(input k8sapiv1.ResourceList) clientapiv1.ResourceList {
	output := make(clientapiv1.ResourceList, len(input))
	for name, quantity := range input {
		var newName clientapiv1.ResourceName = clientapiv1.ResourceName(name.String())
		output[newName] = quantity
	}
	return output
}

func convertResourceListToKubernetesApi(input clientapiv1.ResourceList) k8sapiv1.ResourceList {
	output := make(k8sapiv1.ResourceList, len(input))
	for name, quantity := range input {
		var newName k8sapiv1.ResourceName = k8sapiv1.ResourceName(name.String())
		output[newName] = quantity
	}
	return output
}
