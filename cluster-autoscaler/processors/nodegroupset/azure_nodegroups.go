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
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// AzureNodepoolLegacyLabel is a label specifying which Azure node pool a particular node belongs to.
const AzureNodepoolLegacyLabel = "agentpool"

// AzureNodepoolLabel is an AKS label specifying which nodepool a particular node belongs to
const AzureNodepoolLabel = "kubernetes.azure.com/agentpool"

// AzureDiskTopologyKey is the topology key of Azure Disk CSI driver
const AzureDiskTopologyKey = "topology.disk.csi.azure.com/zone"

// Those labels are added on the VMSS and shouldn't affect nodepool similarity
const aksEngineVersionLabel = "aksEngineVersion"
const creationSource = "creationSource"
const poolName = "poolName"
const resourceNameSuffix = "resourceNameSuffix"
const aksConsolidatedAdditionalProperties = "kubernetes.azure.com/consolidated-additional-properties"

// AKS node image version
const aksNodeImageVersion = "kubernetes.azure.com/node-image-version"

func nodesFromSameAzureNodePool(n1, n2 *framework.NodeInfo) bool {
	n1AzureNodePool := n1.Node().Labels[AzureNodepoolLabel]
	n2AzureNodePool := n2.Node().Labels[AzureNodepoolLabel]
	return (n1AzureNodePool != "" && n1AzureNodePool == n2AzureNodePool) || nodesFromSameAzureNodePoolLegacy(n1, n2)
}

func nodesFromSameAzureNodePoolLegacy(n1, n2 *framework.NodeInfo) bool {
	n1AzureNodePool := n1.Node().Labels[AzureNodepoolLegacyLabel]
	n2AzureNodePool := n2.Node().Labels[AzureNodepoolLegacyLabel]
	return n1AzureNodePool != "" && n1AzureNodePool == n2AzureNodePool
}

// CreateAzureNodeInfoComparator returns a comparator that checks if two nodes should be considered
// part of the same NodeGroupSet. This is true if they either belong to the same Azure agentpool
// or match usual conditions checked by IsCloudProviderNodeInfoSimilar, even if they have different agentpool labels.
func CreateAzureNodeInfoComparator(extraIgnoredLabels []string, ratioOpts config.NodeGroupDifferenceRatios) NodeInfoComparator {
	azureIgnoredLabels := make(map[string]bool)
	for k, v := range BasicIgnoredLabels {
		azureIgnoredLabels[k] = v
	}
	azureIgnoredLabels[AzureNodepoolLegacyLabel] = true
	azureIgnoredLabels[AzureNodepoolLabel] = true
	azureIgnoredLabels[AzureDiskTopologyKey] = true
	azureIgnoredLabels[aksEngineVersionLabel] = true
	azureIgnoredLabels[creationSource] = true
	azureIgnoredLabels[poolName] = true
	azureIgnoredLabels[resourceNameSuffix] = true
	azureIgnoredLabels[aksNodeImageVersion] = true
	azureIgnoredLabels[aksConsolidatedAdditionalProperties] = true

	for _, k := range extraIgnoredLabels {
		azureIgnoredLabels[k] = true
	}

	return func(n1, n2 *framework.NodeInfo) bool {
		if nodesFromSameAzureNodePool(n1, n2) {
			return true
		}
		return IsCloudProviderNodeInfoSimilar(n1, n2, azureIgnoredLabels, ratioOpts)
	}
}
