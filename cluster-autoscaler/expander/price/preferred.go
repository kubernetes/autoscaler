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

package price

import (
	"fmt"
	"math"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

// SimplePreferredNodeProvider returns preferred node based on the cluster size.
type SimplePreferredNodeProvider struct {
	nodeLister kube_util.NodeLister
}

// NewSimplePreferredNodeProvider returns simple PreferredNodeProvider
func NewSimplePreferredNodeProvider(nodeLister kube_util.NodeLister) *SimplePreferredNodeProvider {
	return &SimplePreferredNodeProvider{
		nodeLister: nodeLister,
	}
}

// Node returns preferred node.
func (spnp *SimplePreferredNodeProvider) Node() (*apiv1.Node, error) {
	nodes, err := spnp.nodeLister.List()
	if err != nil {
		return nil, err
	}
	size := len(nodes)

	mb := int64(1024 * 1024)
	cpu := int64(1000)

	// Double node size with every time the cluster size increases 3x.
	if size <= 2 {
		return buildNode(1*cpu, 3750*mb), nil
	} else if size <= 6 {
		return buildNode(2*cpu, 7500*mb), nil
	} else if size <= 20 {
		return buildNode(4*cpu, 15000*mb), nil
	} else if size <= 60 {
		return buildNode(8*cpu, 30000*mb), nil
	} else if size <= 200 {
		return buildNode(16*cpu, 60000*mb), nil
	}
	return buildNode(32*cpu, 120000*mb), nil
}

func buildNode(millicpu int64, mem int64) *apiv1.Node {
	name := "CA-PreferredNode"
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourcePods:   *resource.NewQuantity(100, resource.DecimalSI),
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(millicpu, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	return node
}

// SimpleNodeUnfitness returns unfitness based on cpu only.
func SimpleNodeUnfitness(preferredNode, evaluatedNode *apiv1.Node) float64 {
	preferredCpu := preferredNode.Status.Capacity[apiv1.ResourceCPU]
	evaluatedCpu := evaluatedNode.Status.Capacity[apiv1.ResourceCPU]
	return math.Max(float64(preferredCpu.MilliValue())/float64(evaluatedCpu.MilliValue()),
		float64(evaluatedCpu.MilliValue())/float64(preferredCpu.MilliValue()))
}
