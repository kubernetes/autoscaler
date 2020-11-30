/*
Copyright 2019 The Kubernetes Authors.

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

package nodegroupset

import (
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// AzureNodepoolLabel is a label specifying which Azure node pool a particular node belongs to.
const AzureNodepoolLabel = "agentpool"

func nodesFromSameAzureNodePool(n1, n2 *schedulerframework.NodeInfo) bool {
	n1AzureNodePool := n1.Node().Labels[AzureNodepoolLabel]
	n2AzureNodePool := n2.Node().Labels[AzureNodepoolLabel]
	return n1AzureNodePool != "" && n1AzureNodePool == n2AzureNodePool
}

// CreateAzureNodeInfoComparator returns a comparator that checks if two nodes should be considered
// part of the same NodeGroupSet. This is true if they either belong to the same Azure agentpool
// or match usual conditions checked by IsCloudProviderNodeInfoSimilar, even if they have different agentpool labels.
func CreateAzureNodeInfoComparator(extraIgnoredLabels []string) NodeInfoComparator {
	azureIgnoredLabels := make(map[string]bool)
	for k, v := range BasicIgnoredLabels {
		azureIgnoredLabels[k] = v
	}
	azureIgnoredLabels[AzureNodepoolLabel] = true
	for _, k := range extraIgnoredLabels {
		azureIgnoredLabels[k] = true
	}

	return func(n1, n2 *schedulerframework.NodeInfo) bool {
		if nodesFromSameAzureNodePool(n1, n2) {
			return true
		}
		return IsCloudProviderNodeInfoSimilar(n1, n2, azureIgnoredLabels)
	}
}
