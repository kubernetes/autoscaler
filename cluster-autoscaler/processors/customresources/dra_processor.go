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

package customresources

import (
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/dra"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// DraCustomResourcesProcessor handles DRA custom resource. It assumes,
// that the DRA resources may not become allocatable immediately after the node creation.
type DraCustomResourcesProcessor struct {
}

// FilterOutNodesWithUnreadyResources removes nodes that should have DRA resource, but don't have
// it in allocatable from ready nodes list and updates their status to unready on all nodes list.
func (p *DraCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(autoscalingCtx *ca_context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, draSnapshot *snapshot.Snapshot, _ *csisnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyDraResources := make(map[string]*apiv1.Node)
	if draSnapshot == nil {
		klog.Warningf("Cannot filter out nodes with unready DRA resources. The DRA snapshot is nil. Processing will be skipped.")
		return allNodes, readyNodes
	}

	for _, node := range readyNodes {
		ng, err := autoscalingCtx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get node group for node %s, Skipping DRA readiness check and keeping node in ready list. Error: %v", node.Name, err)
			continue
		}
		if ng == nil {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		nodeInfo, err := getNodeInfo(autoscalingCtx, ng)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get template node info for node group %s with error: %v", ng.Id(), err)
			continue
		}

		nodeResourcesSlices, _ := draSnapshot.NodeResourceSlices(node.Name)
		if isEqualResourceSlices(nodeResourcesSlices, nodeInfo.LocalResourceSlices) {
			newReadyNodes = append(newReadyNodes, node)
		} else {
			nodesWithUnreadyDraResources[node.Name] = kubernetes.GetUnreadyNodeCopy(node, kubernetes.ResourceUnready)
		}
	}

	// Override any node with unready DRA resources with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyDraResources[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

func getNodeInfo(autoscalingCtx *ca_context.AutoscalingContext, ng cloudprovider.NodeGroup) (*framework.NodeInfo, error) {
	// Prefer the cached template from the registry. This template may contain enrichments (e.g.
	// custom DRA slices) that are not present in the raw CloudProvider template.
	if ni, found := autoscalingCtx.TemplateNodeInfoRegistry.GetNodeInfo(ng.Id()); found {
		return ni, nil
	}
	return ng.TemplateNodeInfo()
}

type resourceSliceSpecs struct {
	driver string
	pool   string
}

func isEqualResourceSlices(nodeResourcesSlices []*resourceapi.ResourceSlice, templateResourcesSlices []*resourceapi.ResourceSlice) bool {
	tempSlicesByPools := getDevicesBySpecs(templateResourcesSlices)
	nodeSlicesByPools := getDevicesBySpecs(nodeResourcesSlices)

	for templSpecs, tempDevicesSet := range tempSlicesByPools {
		matched := false
		for nodeSpecs, nodeDevicesSet := range nodeSlicesByPools {
			if templSpecs.driver == nodeSpecs.driver && nodeDevicesSet.Equal(tempDevicesSet) {
				delete(nodeSlicesByPools, nodeSpecs)
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func getDevicesBySpecs(resourcesSlices []*resourceapi.ResourceSlice) map[resourceSliceSpecs]sets.Set[string] {
	slicesGroupedByPoolAndDriver := make(map[resourceSliceSpecs]sets.Set[string])
	for _, rs := range resourcesSlices {
		rsSpecs := resourceSliceSpecs{
			pool:   rs.Spec.Pool.Name,
			driver: rs.Spec.Driver,
		}
		slicesGroupedByPoolAndDriver[rsSpecs] = getResourceSliceDevicesSet(rs)
	}
	return slicesGroupedByPoolAndDriver
}

func getResourceSliceDevicesSet(resourcesSlice *resourceapi.ResourceSlice) sets.Set[string] {
	devices := sets.New[string]()
	for _, device := range resourcesSlice.Spec.Devices {
		devices.Insert(device.Name)
	}
	return devices
}

// GetNodeResourceTargets returns the resource targets for DRA resource slices.
// It counts unique devices by their UID across all resource slices in the node template.
func (p *DraCustomResourcesProcessor) GetNodeResourceTargets(_ *ca_context.AutoscalingContext, node *apiv1.Node, ng cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	// deviceResourceMap maps a resourceType to a set of unique device UIDs exposed as this resource type.
	// Using a map of sets prevents double-counting the same device.
	deviceResourceMap := make(map[string]sets.Set[string])

	nodeInfo, err := ng.TemplateNodeInfo()
	if err != nil {
		// Template building is not supported by this cloud provider, return empty targets.
		// This is expected for cloud providers that don't support DRA resource templates yet.
		klog.V(4).Infof("Template not available for node %v: %v, skipping DRA resource target calculation", node.Name, err)
		return []CustomResourceTarget{}, nil
	}

	for _, rs := range nodeInfo.LocalResourceSlices {
		uidAttributeName, productNameAttributeName, supported := getDriverAttributes(rs.Spec.Driver)
		if !supported {
			klog.Warningf("Unknown DRA driver %s for node %s, skipping DRA resource target calculation", rs.Spec.Driver, node.Name)
			continue
		}

		for _, device := range rs.Spec.Devices {
			uid, foundUID := device.Attributes[resourceapi.QualifiedName(uidAttributeName)]
			if !foundUID {
				klog.Warningf("Device %s on node %s is missing uid attribute, this might indicate that autoscaler is out of date with the driver, skipping", device.Name, node.Name)
				continue
			}

			productName, foundProduct := device.Attributes[resourceapi.QualifiedName(productNameAttributeName)]
			if !foundProduct {
				klog.Warningf("Device %s on node %s is missing product name attribute, this might indicate that autoscaler is out of date with the driver, skipping", device.Name, node.Name)
				continue
			}

			resourceName := dra.GetDraResourceName(rs.Spec.Driver, *productName.StringValue)
			if _, exists := deviceResourceMap[resourceName]; !exists {
				deviceResourceMap[resourceName] = sets.New[string]()
			}
			deviceResourceMap[resourceName].Insert(uid.String())
		}
	}

	customResourceTargets := make([]CustomResourceTarget, 0, len(deviceResourceMap))
	for resourceName, deviceUIDs := range deviceResourceMap {
		customResourceTargets = append(customResourceTargets, CustomResourceTarget{
			ResourceType:  resourceName,
			ResourceCount: int64(deviceUIDs.Len()),
		})
	}

	return customResourceTargets, nil
}

// getDriverAttributes returns the UID and product name attribute keys for a given driver.
// Returns empty strings and false if the driver is not supported.
func getDriverAttributes(driver string) (uidAttribute string, productNameAttribute string, supported bool) {
	switch driver {
	case dra.DriverNvidiaGPUName:
		return dra.DriverNvidiaGPUUid, dra.DriverNvidiaGPUProductName, true
	default:
		return "", "", false
	}
}

// CleanUp cleans up processor's internal structures.
func (p *DraCustomResourcesProcessor) CleanUp() {
}
