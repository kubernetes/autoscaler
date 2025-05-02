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

package dynamicresources

import (
	"reflect"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

// DynamicResourcesProcessor handles dynamic resource.
// dynamic resources may not be allocatable immediately after the node creation.
// It compares the expected resourceslices with the existing resourceslices to assess node readiness.
type DynamicResourcesProcessor interface {
	// FilterOutNodesWithUnreadyResources removes nodes that should have dynamic resources, but have not published resourceslices yet.
	FilterOutNodesWithUnreadyResources(
		context *ca_context.AutoscalingContext,
		allNodes, readyNodes []*apiv1.Node,
		resourceSlices []*resourceapi.ResourceSlice,
	) ([]*apiv1.Node, []*apiv1.Node, error)
	CleanUp()
}

// NewDefaultDynamicResourcesProcessor returns a default instance of DynamicResourcesProcessor.
func NewDefaultDynamicResourcesProcessor() DynamicResourcesProcessor {
	return &dynamicResourcesProcessor{}
}

type dynamicResourcesProcessor struct{}

// FilterOutNodesWithUnreadyResources removes nodes that should have dynamic resources, but have not published resourceslices yet.
func (p *dynamicResourcesProcessor) FilterOutNodesWithUnreadyResources(
	context *ca_context.AutoscalingContext,
	allNodes, readyNodes []*apiv1.Node,
	resourceSlices []*resourceapi.ResourceSlice,
) ([]*apiv1.Node, []*apiv1.Node, error) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyResources := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		isReady, err := p.checkNodeReadiness(context, node, resourceSlices)
		if err != nil {
			return nil, nil, err
		}
		if isReady {
			newReadyNodes = append(newReadyNodes, node)
		} else {
			nodesWithUnreadyResources[node.Name] = node
		}
	}
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyResources[node.Name]; found {
			newAllNodes = append(newAllNodes, kubernetes.GetUnreadyNodeCopy(newNode, kubernetes.ResourceUnready))
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes, nil
}

func (p *dynamicResourcesProcessor) checkNodeReadiness(
	context *ca_context.AutoscalingContext,
	node *apiv1.Node,
	resourceSlices []*resourceapi.ResourceSlice,
) (bool, error) {
	nodegroup, err := context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return false, err
	}
	if nodegroup == nil { // Node is not by autoscaler
		return true, nil
	}
	nodeTemplate, err := nodegroup.TemplateNodeInfo()
	if err != nil {
		return false, err
	}
	templateResourceSlices, _, err := drautils.SanitizedNodeResourceSlices(nodeTemplate.LocalResourceSlices, node.Name, "")
	if err != nil {
		return false, err
	}
	if len(templateResourceSlices) == 0 {
		return true, nil
	}
	clusterResourceSlices := resourceSlices
	nodeResourceSlices := make([]*resourceapi.ResourceSlice, 0)
	for _, rs := range clusterResourceSlices {
		if rs != nil && rs.Spec.NodeName == node.Name {
			nodeResourceSlices = append(nodeResourceSlices, rs)
		}
	}

	nodeResourceSlices, _, err = drautils.SanitizedNodeResourceSlices(nodeResourceSlices, node.Name, "")
	if err != nil {
		return false, err
	}

	if len(templateResourceSlices) != len(nodeResourceSlices) {
		return false, nil // Different number of slices means not ready/matched yet
	}

	for _, templateResourceSlice := range templateResourceSlices {
		var matched bool = false
		for _, nodeResourceSlice := range nodeResourceSlices {
			if compareResourceSlices(templateResourceSlice, nodeResourceSlice) {
				matched = true
				break
			}
			if !matched {
				return false, nil // No match was found for this template slice on the node
			}
		}
	}

	return true, nil // All template slices found a match

}

func compareResourceSlices(
	resourceSlice1, resourceSlice2 *resourceapi.ResourceSlice,
) bool {
	// In order to assess whether the expected resourceslices have been published
	// we only need to compare the spec
	if resourceSlice1 == nil && resourceSlice2 == nil {
		return true
	}
	if resourceSlice1 == nil || resourceSlice2 == nil {
		return false
	}
	if resourceSlice1.Spec.Driver != resourceSlice2.Spec.Driver {
		return false
	}
	if resourceSlice1.Spec.NodeName != resourceSlice2.Spec.NodeName {
		return false
	}
	devices1 := resourceSlice1.Spec.Devices
	devices2 := resourceSlice2.Spec.Devices
	if len(devices1) != len(devices2) {
		return false
	}

	matched2 := make([]bool, len(devices2))
	for _, d1 := range devices1 {
		foundMatch := false
		for j, d2 := range devices2 {
			if !matched2[j] && compareDevicesIgnoringName(d1, d2) {
				matched2[j] = true
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return false
		}
	}

	if !reflect.DeepEqual(resourceSlice1.Spec.AllNodes, resourceSlice2.Spec.AllNodes) {
		return false
	}
	if !reflect.DeepEqual(resourceSlice1.Spec.NodeSelector, resourceSlice2.Spec.NodeSelector) {
		return false
	}

	return true
}

// CleanUp cleans up processor's internal structures.
func (p *dynamicResourcesProcessor) CleanUp() {
}

type mockDynamicResourcesProcessor struct {
}

func (m *mockDynamicResourcesProcessor) FilterOutNodesWithUnreadyResources(
	context *ca_context.AutoscalingContext,
	allNodes, readyNodes []*apiv1.Node,
	resourceSlices []*resourceapi.ResourceSlice,
) ([]*apiv1.Node, []*apiv1.Node, error) {
	return allNodes, readyNodes, nil
}

func (m *mockDynamicResourcesProcessor) CleanUp() {
}

// NewMockDynamicResourcesProcessor returns a mock instance of DynamicResourcesProcessor.
func NewMockDynamicResourcesProcessor() DynamicResourcesProcessor {
	return &mockDynamicResourcesProcessor{}
}

func compareDevicesIgnoringName(d1, d2 resourceapi.Device) bool {
	return reflect.DeepEqual(d1.Basic, d2.Basic)
}
