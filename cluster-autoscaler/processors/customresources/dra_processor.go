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
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/comparator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// resourceDiscrepancyReporter defines the interface for reporting DRA discrepancies.
type resourceDiscrepancyReporter interface {
	ReportResourceDiscrepancies(nodeNames []string, templateSlices [][]*resourceapi.ResourceSlice, nodeSlices [][]*resourceapi.ResourceSlice)
}

// DraCustomResourcesProcessor handles DRA custom resource. It assumes,
// that the DRA resources may not become allocatable immediately after the node creation.
type DraCustomResourcesProcessor struct {
	resourcesComparator resourceDiscrepancyReporter
}

// NewDraCustomResourcesProcessor returns an instance of DraCustomResourcesProcessor properly initialized.
func NewDraCustomResourcesProcessor() *DraCustomResourcesProcessor {
	return &DraCustomResourcesProcessor{
		resourcesComparator: comparator.NewNodeResourcesComparator(metrics.DefaultMetrics),
	}
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

	readyNodeNames := make([]string, 0, len(readyNodes))
	readyTemplateSlices := make([][]*resourceapi.ResourceSlice, 0, len(readyNodes))
	readyNodeSlices := make([][]*resourceapi.ResourceSlice, 0, len(readyNodes))

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

		templateNodeInfo, err := getNodeInfo(autoscalingCtx, ng)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get template node info for node group %s with error: %v", ng.Id(), err)
			continue
		}

		nodeResourcesSlices, _ := draSnapshot.NodeResourceSlices(node.Name)
		if areResourcePoolsReady(nodeResourcesSlices, templateNodeInfo.LocalResourceSlices) {
			newReadyNodes = append(newReadyNodes, node)
			readyNodeNames = append(readyNodeNames, node.Name)
			readyTemplateSlices = append(readyTemplateSlices, templateNodeInfo.LocalResourceSlices)
			readyNodeSlices = append(readyNodeSlices, nodeResourcesSlices)
		} else {
			nodesWithUnreadyDraResources[node.Name] = kubernetes.GetUnreadyNodeCopy(node, kubernetes.ResourceUnready)
		}
	}

	p.resourcesComparator.ReportResourceDiscrepancies(readyNodeNames, readyTemplateSlices, readyNodeSlices)

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

// GetNodeResourceTargets returns the resource targets for DRA resource slices, not implemented.
func (p *DraCustomResourcesProcessor) GetNodeResourceTargets(_ *ca_context.AutoscalingContext, _ *apiv1.Node, _ cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	// TODO(DRA): Figure out resource limits for DRA here.
	return []CustomResourceTarget{}, nil
}

// CleanUp cleans up processor's internal structures.
func (p *DraCustomResourcesProcessor) CleanUp() {
}

type poolState struct {
	count      int64
	generation int64
	size       int64
}

type poolSpec struct {
	name   string
	driver string
}

// areResourcePoolsReady returns boolean indicating whether resource slices from a real node
// contain a minimal amount of ready resource pools as declared in the template.
func areResourcePoolsReady(resourceSlices []*resourceapi.ResourceSlice, templateNodeResourcesSlices []*resourceapi.ResourceSlice) bool {
	templatePools := getCompleteResourcePools(templateNodeResourcesSlices)
	realPools := getCompleteResourcePools(resourceSlices)

	for driver, count := range templatePools {
		if realPools[driver] < count {
			return false
		}
	}

	return true
}

// getCompleteResourcePools returns a map of drivers and count of ready resource pools mapped to it.
func getCompleteResourcePools(resourceSlices []*resourceapi.ResourceSlice) map[string]int {
	poolStates := make(map[poolSpec]poolState, len(resourceSlices))

	for _, rs := range resourceSlices {
		spec := poolSpec{name: rs.Spec.Pool.Name, driver: rs.Spec.Driver}
		state, found := poolStates[spec]

		// If the pool is new, or if we found a newer generation, overwrite/reset the state.
		if !found || rs.Spec.Pool.Generation > state.generation {
			state = poolState{
				count:      0,
				generation: rs.Spec.Pool.Generation,
				size:       rs.Spec.Pool.ResourceSliceCount,
			}
		} else if rs.Spec.Pool.Generation < state.generation {
			// Ignore slices from older generations
			continue
		}

		state.count++
		poolStates[spec] = state
	}

	// Pre-allocate map to avoid dynamic resize overhead during the loop.
	readyPoolsByDriver := make(map[string]int, len(poolStates))
	for spec, state := range poolStates {
		// A pool is ready if we have seen strictly the expected number of resource slices.
		if state.count == state.size {
			readyPoolsByDriver[spec.driver]++
		}
	}

	return readyPoolsByDriver
}
