/*
Copyright 2025 The Kubernetes Authors.

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

package resourcequotas

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	cacontext "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

type nodeResourcesCache struct {
	crp       customresources.CustomResourcesProcessor
	resources map[string]resourceList
}

func newNodeResourcesCache(crp customresources.CustomResourcesProcessor) *nodeResourcesCache {
	return &nodeResourcesCache{
		crp:       crp,
		resources: make(map[string]resourceList),
	}
}

// totalNodeResources returns the map of resources that a node contains.
// If the resources have been calculated previously for a node within the same
// node group, the resources will be retrieved from the cache. Otherwise,
// the resources will be calculated. If the node belongs to any node group,
// the resources will be cached for the entire node group.
func (nc *nodeResourcesCache) totalNodeResources(
	autoscalingCtx *cacontext.AutoscalingContext, node *corev1.Node, nodeGroup cloudprovider.NodeGroup,
) (resourceList, error) {
	if nodeGroup != nil {
		if resources, ok := nc.resources[nodeGroup.Id()]; ok {
			return resources, nil
		}
	}
	resources, err := totalNodeResources(autoscalingCtx, nc.crp, node, nodeGroup)
	if err != nil {
		return nil, err
	}
	if nodeGroup != nil {
		nc.resources[nodeGroup.Id()] = resources
	}
	return resources, nil
}

// totalNodeResources calculates the amount of resources that a node contains.
func totalNodeResources(autoscalingCtx *cacontext.AutoscalingContext, crp customresources.CustomResourcesProcessor, node *corev1.Node, nodeGroup cloudprovider.NodeGroup) (resourceList, error) {
	// TODO: storage?
	nodeCPU, nodeMemory := utils.GetNodeCoresAndMemory(node)
	nodeResources := resourceList{
		string(corev1.ResourceCPU):    nodeCPU,
		string(corev1.ResourceMemory): nodeMemory,
		ResourceNodes:                 1,
	}

	resourceTargets, err := crp.GetNodeResourceTargets(autoscalingCtx, node, nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom resources: %w", err)
	}

	for _, resourceTarget := range resourceTargets {
		if resourceTarget.ResourceType != "" && resourceTarget.ResourceCount > 0 {
			nodeResources[resourceTarget.ResourceType] = resourceTarget.ResourceCount
		}
	}

	return nodeResources, nil
}
