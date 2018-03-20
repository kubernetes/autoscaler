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
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"
)

const (
	// MaxAllocatableDifferenceRatio describes how Node.Status.Allocatable can differ between
	// groups in the same NodeGroupSet
	MaxAllocatableDifferenceRatio = 0.05
	// MaxFreeDifferenceRatio describes how free resources (allocatable - daemon and system pods)
	// can differ between groups in the same NodeGroupSet
	MaxFreeDifferenceRatio = 0.05
)

func compareResourceMapsWithTolerance(resources map[apiv1.ResourceName][]resource.Quantity,
	maxDifferenceRatio float64) bool {
	for _, qtyList := range resources {
		if len(qtyList) != 2 {
			return false
		}
		larger := math.Max(float64(qtyList[0].MilliValue()), float64(qtyList[1].MilliValue()))
		smaller := math.Min(float64(qtyList[0].MilliValue()), float64(qtyList[1].MilliValue()))
		if larger-smaller > larger*maxDifferenceRatio {
			return false
		}
	}
	return true
}

// IsNodeInfoSimilar returns true if two NodeInfos are similar enough to consider
// that the NodeGroups they come from are part of the same NodeGroupSet. The criteria are
// somewhat arbitrary, but generally we check if resources provided by both nodes
// are similar enough to likely be the same type of machine and if the set of labels
// is the same (except for a pre-defined set of labels like hostname or zone).
func IsNodeInfoSimilar(n1, n2 *schedulercache.NodeInfo) bool {
	capacity := make(map[apiv1.ResourceName][]resource.Quantity)
	allocatable := make(map[apiv1.ResourceName][]resource.Quantity)
	free := make(map[apiv1.ResourceName][]resource.Quantity)
	nodes := []*schedulercache.NodeInfo{n1, n2}
	for _, node := range nodes {
		for res, quantity := range node.Node().Status.Capacity {
			capacity[res] = append(capacity[res], quantity)
		}
		for res, quantity := range node.Node().Status.Allocatable {
			allocatable[res] = append(allocatable[res], quantity)
		}
		requested := node.RequestedResource()
		for res, quantity := range (&requested).ResourceList() {
			freeRes := node.Node().Status.Allocatable[res].DeepCopy()
			freeRes.Sub(quantity)
			free[res] = append(free[res], freeRes)
		}
	}
	// For capacity we require exact match.
	// If this is ever changed, enforcing MaxCoresTotal and MaxMemoryTotal limits
	// as it is now may no longer work.
	for _, qtyList := range capacity {
		if len(qtyList) != 2 || qtyList[0].Cmp(qtyList[1]) != 0 {
			return false
		}
	}
	// For allocatable and free we allow resource quantities to be within a few % of each other
	if !compareResourceMapsWithTolerance(allocatable, MaxAllocatableDifferenceRatio) {
		return false
	}
	if !compareResourceMapsWithTolerance(free, MaxFreeDifferenceRatio) {
		return false
	}

	labels := make(map[string][]string)
	for _, node := range nodes {
		for label, value := range node.Node().ObjectMeta.Labels {
			if label == kubeletapis.LabelHostname {
				continue
			}
			if label == kubeletapis.LabelZoneFailureDomain {
				continue
			}
			if label == kubeletapis.LabelZoneRegion {
				continue
			}
			labels[label] = append(labels[label], value)
		}
	}
	for _, labelValues := range labels {
		if len(labelValues) != 2 || labelValues[0] != labelValues[1] {
			return false
		}
	}
	return true
}
