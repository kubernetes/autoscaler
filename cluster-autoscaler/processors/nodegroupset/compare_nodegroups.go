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

package nodegroupset

import (
	"math"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	// MaxAllocatableDifferenceRatio describes how Node.Status.Allocatable can differ between
	// groups in the same NodeGroupSet
	MaxAllocatableDifferenceRatio = 0.05
	// MaxFreeDifferenceRatio describes how free resources (allocatable - daemon and system pods)
	// can differ between groups in the same NodeGroupSet
	MaxFreeDifferenceRatio = 0.05
	// MaxCapacityMemoryDifferenceRatio describes how Node.Status.Capacity.Memory can differ between
	// groups in the same NodeGroupSet
	MaxCapacityMemoryDifferenceRatio = 0.015
)

// BasicIgnoredLabels define a set of basic labels that should be ignored when comparing the similarity
// of two nodes. Customized IgnoredLabels can be implemented in the corresponding codes of
// specific cloud provider and the BasicIgnoredLabels should always be considered part of them.
var BasicIgnoredLabels = map[string]bool{
	apiv1.LabelHostname:                   true,
	apiv1.LabelZoneFailureDomain:          true,
	apiv1.LabelZoneRegion:                 true,
	apiv1.LabelZoneFailureDomainStable:    true,
	apiv1.LabelZoneRegionStable:           true,
	"beta.kubernetes.io/fluentd-ds-ready": true, // this is internal label used for determining if fluentd should be installed as deamon set. Used for migration 1.8 to 1.9.
	"kops.k8s.io/instancegroup":           true, // this is a label used by kops to identify "instance group" names. it's value is variable, defeating check of similar node groups
}

// NodeInfoComparator is a function that tells if two nodes are from NodeGroups
// similar enough to be considered a part of a single NodeGroupSet.
type NodeInfoComparator func(n1, n2 *schedulerframework.NodeInfo) bool

func resourceMapsWithinTolerance(resources map[apiv1.ResourceName][]resource.Quantity,
	maxDifferenceRatio float64) bool {
	for _, qtyList := range resources {
		if !resourceListWithinTolerance(qtyList, maxDifferenceRatio) {
			return false
		}
	}
	return true
}

func resourceListWithinTolerance(qtyList []resource.Quantity, maxDifferenceRatio float64) bool {
	if len(qtyList) != 2 {
		return false
	}
	larger := math.Max(float64(qtyList[0].MilliValue()), float64(qtyList[1].MilliValue()))
	smaller := math.Min(float64(qtyList[0].MilliValue()), float64(qtyList[1].MilliValue()))
	return larger-smaller <= larger*maxDifferenceRatio
}

func compareLabels(nodes []*schedulerframework.NodeInfo, ignoredLabels map[string]bool) bool {
	labels := make(map[string][]string)
	for _, node := range nodes {
		for label, value := range node.Node().ObjectMeta.Labels {
			ignore, _ := ignoredLabels[label]
			if !ignore {
				labels[label] = append(labels[label], value)
			}
		}
	}
	for _, labelValues := range labels {
		if len(labelValues) != 2 || labelValues[0] != labelValues[1] {
			return false
		}
	}
	return true
}

// CreateGenericNodeInfoComparator returns a generic comparator that checks for node group similarity
func CreateGenericNodeInfoComparator(extraIgnoredLabels []string) NodeInfoComparator {
	genericIgnoredLabels := make(map[string]bool)
	for k, v := range BasicIgnoredLabels {
		genericIgnoredLabels[k] = v
	}
	for _, k := range extraIgnoredLabels {
		genericIgnoredLabels[k] = true
	}

	return func(n1, n2 *schedulerframework.NodeInfo) bool {
		return IsCloudProviderNodeInfoSimilar(n1, n2, genericIgnoredLabels)
	}
}

// IsCloudProviderNodeInfoSimilar returns true if two NodeInfos are similar enough to consider
// that the NodeGroups they come from are part of the same NodeGroupSet. The criteria are
// somewhat arbitrary, but generally we check if resources provided by both nodes
// are similar enough to likely be the same type of machine and if the set of labels
// is the same (except for a set of labels passed in to be ignored like hostname or zone).
func IsCloudProviderNodeInfoSimilar(n1, n2 *schedulerframework.NodeInfo, ignoredLabels map[string]bool) bool {
	capacity := make(map[apiv1.ResourceName][]resource.Quantity)
	allocatable := make(map[apiv1.ResourceName][]resource.Quantity)
	free := make(map[apiv1.ResourceName][]resource.Quantity)
	nodes := []*schedulerframework.NodeInfo{n1, n2}
	for _, node := range nodes {
		for res, quantity := range node.Node().Status.Capacity {
			capacity[res] = append(capacity[res], quantity)
		}
		for res, quantity := range node.Node().Status.Allocatable {
			allocatable[res] = append(allocatable[res], quantity)
		}
		for res, quantity := range node.Requested.ResourceList() {
			freeRes := node.Node().Status.Allocatable[res].DeepCopy()
			freeRes.Sub(quantity)
			free[res] = append(free[res], freeRes)
		}
	}

	for kind, qtyList := range capacity {
		if len(qtyList) != 2 {
			return false
		}
		switch kind {
		case apiv1.ResourceMemory:
			if !resourceListWithinTolerance(qtyList, MaxCapacityMemoryDifferenceRatio) {
				return false
			}
		default:
			// For other capacity types we require exact match.
			// If this is ever changed, enforcing MaxCoresTotal limits
			// as it is now may no longer work.
			if qtyList[0].Cmp(qtyList[1]) != 0 {
				return false
			}
		}
	}

	// For allocatable and free we allow resource quantities to be within a few % of each other
	if !resourceMapsWithinTolerance(allocatable, MaxAllocatableDifferenceRatio) {
		return false
	}
	if !resourceMapsWithinTolerance(free, MaxFreeDifferenceRatio) {
		return false
	}

	if !compareLabels(nodes, ignoredLabels) {
		return false
	}

	return true
}
